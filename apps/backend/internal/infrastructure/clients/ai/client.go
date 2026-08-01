package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gabriel-q7/portfolio/backend/pkg/logger"
	"golang.org/x/time/rate"
)

// AIProvider abstracts an AI service.
type AIProvider interface {
	Complete(ctx context.Context, prompt string, opts CompletionOptions) (*CompletionResult, error)
	Embed(ctx context.Context, text string) ([]float64, error)
	ProviderName() string
}

// CompletionOptions configures a completion request.
type CompletionOptions struct {
	Model        string
	MaxTokens    int
	Temperature  float64
	SystemPrompt string
}

// CompletionResult holds the AI response.
type CompletionResult struct {
	Content      string
	Model        string
	InputTokens  int
	OutputTokens int
	FinishReason string
}

type openAIClient struct {
	apiKey      string
	baseURL     string
	httpClient  *http.Client
	logger      logger.Logger
	rateLimiter *rate.Limiter
}

// NewOpenAIClient creates an OpenAI-compatible AI provider.
func NewOpenAIClient(apiKey, baseURL string, rateLimit float64, log logger.Logger) AIProvider {
	return &openAIClient{
		apiKey:      apiKey,
		baseURL:     baseURL,
		httpClient:  &http.Client{Timeout: 60 * time.Second},
		logger:      log,
		rateLimiter: rate.NewLimiter(rate.Limit(rateLimit), int(rateLimit)+1),
	}
}

func (c *openAIClient) ProviderName() string { return "openai" }

func (c *openAIClient) Complete(ctx context.Context, prompt string, opts CompletionOptions) (*CompletionResult, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter: %w", err)
	}

	model := opts.Model
	if model == "" {
		model = "gpt-4o-mini"
	}

	messages := []map[string]string{}
	if opts.SystemPrompt != "" {
		messages = append(messages, map[string]string{"role": "system", "content": opts.SystemPrompt})
	}
	messages = append(messages, map[string]string{"role": "user", "content": prompt})

	reqBody := map[string]any{
		"model":       model,
		"messages":    messages,
		"max_tokens":  opts.MaxTokens,
		"temperature": opts.Temperature,
	}

	var result *CompletionResult
	var lastErr error
	backoff := 100 * time.Millisecond

	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
				backoff *= 2
			}
		}

		result, lastErr = c.doComplete(ctx, reqBody, model)
		if lastErr == nil {
			break
		}
		c.logger.Warn("AI complete attempt failed", "attempt", attempt+1, "error", lastErr)
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return result, nil
}

func (c *openAIClient) doComplete(ctx context.Context, reqBody map[string]any, model string) (*CompletionResult, error) {
	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai status %d: %s", resp.StatusCode, body)
	}

	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Model string `json:"model"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &CompletionResult{
		Content:      apiResp.Choices[0].Message.Content,
		Model:        apiResp.Model,
		InputTokens:  apiResp.Usage.PromptTokens,
		OutputTokens: apiResp.Usage.CompletionTokens,
		FinishReason: apiResp.Choices[0].FinishReason,
	}, nil
}

func (c *openAIClient) Embed(ctx context.Context, text string) ([]float64, error) {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
	}

	reqBody := map[string]any{
		"model": "text-embedding-3-small",
		"input": text,
	}
	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal embed request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/embeddings", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("build embed request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http embed request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("read embed body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai embed status %d: %s", resp.StatusCode, body)
	}

	var apiResp struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal embed response: %w", err)
	}
	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data in response")
	}
	return apiResp.Data[0].Embedding, nil
}

// NoopAIClient returns empty results without calling any API.
type NoopAIClient struct{}

func (n *NoopAIClient) ProviderName() string { return "noop" }

func (n *NoopAIClient) Complete(_ context.Context, _ string, opts CompletionOptions) (*CompletionResult, error) {
	return &CompletionResult{
		Content:      "",
		Model:        opts.Model,
		FinishReason: "stop",
	}, nil
}

func (n *NoopAIClient) Embed(_ context.Context, _ string) ([]float64, error) {
	return []float64{}, nil
}
