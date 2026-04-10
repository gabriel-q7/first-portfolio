package external_apis

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gabriel-q7/portfolio/backend/pkg/logger"
	"github.com/gabriel-q7/portfolio/backend/pkg/metrics"
)

type circuitState int

const (
	stateClosed   circuitState = iota
	stateOpen
	stateHalfOpen
)

type circuitBreaker struct {
	mu               sync.Mutex
	state            circuitState
	failures         int
	failureThreshold int
	recoveryTime     time.Duration
	lastFailure      time.Time
}

func newCircuitBreaker() *circuitBreaker {
	return &circuitBreaker{
		state:            stateClosed,
		failureThreshold: 5,
		recoveryTime:     30 * time.Second,
	}
}

func (cb *circuitBreaker) allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	switch cb.state {
	case stateClosed:
		return true
	case stateOpen:
		if time.Since(cb.lastFailure) >= cb.recoveryTime {
			cb.state = stateHalfOpen
			return true
		}
		return false
	case stateHalfOpen:
		return true
	}
	return false
}

func (cb *circuitBreaker) recordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.state = stateClosed
}

func (cb *circuitBreaker) recordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	cb.lastFailure = time.Now()
	if cb.failures >= cb.failureThreshold {
		cb.state = stateOpen
	}
}

// Config holds external API client configuration.
type Config struct {
	BaseURL    string
	Timeout    time.Duration
	MaxRetries int
	APIKey     string
	UserAgent  string
}

// Client is a reusable HTTP client for external APIs.
type Client struct {
	cfg            Config
	httpClient     *http.Client
	logger         logger.Logger
	metrics        *metrics.Metrics
	circuitBreaker *circuitBreaker
}

// New creates a new external API Client.
func New(cfg Config, log logger.Logger, m *metrics.Metrics) *Client {
	return &Client{
		cfg:            cfg,
		httpClient:     &http.Client{Timeout: cfg.Timeout},
		logger:         log,
		metrics:        m,
		circuitBreaker: newCircuitBreaker(),
	}
}

// Get performs an HTTP GET request.
func (c *Client) Get(ctx context.Context, path string, headers map[string]string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.cfg.BaseURL+path, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("build GET request: %w", err)
	}
	c.applyHeaders(req, headers)
	return c.do(req)
}

// Post performs an HTTP POST request.
func (c *Client) Post(ctx context.Context, path string, body []byte, headers map[string]string) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, 0, fmt.Errorf("build POST request: %w", err)
	}
	c.applyHeaders(req, headers)
	return c.do(req)
}

func (c *Client) applyHeaders(req *http.Request, extra map[string]string) {
	if c.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	}
	if c.cfg.UserAgent != "" {
		req.Header.Set("User-Agent", c.cfg.UserAgent)
	}
	for k, v := range extra {
		req.Header.Set(k, v)
	}
}

// do executes the request with retry and circuit breaker logic.
func (c *Client) do(req *http.Request) ([]byte, int, error) {
	if !c.circuitBreaker.allow() {
		return nil, 0, fmt.Errorf("circuit breaker open for %s", c.cfg.BaseURL)
	}

	base := 200 * time.Millisecond
	maxBackoff := 5 * time.Second
	var lastErr error

	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(float64(base) * math.Pow(2, float64(attempt-1)))
			jitter := time.Duration(rand.Int63n(int64(backoff / 2)))
			sleep := backoff + jitter
			if sleep > maxBackoff {
				sleep = maxBackoff
			}
			c.logger.Warn("Retrying external API request",
				"attempt", attempt,
				"url", req.URL.String(),
				"backoff_ms", sleep.Milliseconds(),
			)
			select {
			case <-req.Context().Done():
				return nil, 0, req.Context().Err()
			case <-time.After(sleep):
			}

			// Rebuild body reader for retries (if needed, body must be re-readable).
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			c.circuitBreaker.recordFailure()
			c.logger.Error("External API request failed", "url", req.URL.String(), "error", err)
			continue
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = fmt.Errorf("read response body: %w", readErr)
			c.circuitBreaker.recordFailure()
			continue
		}

		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("server error status %d", resp.StatusCode)
			c.circuitBreaker.recordFailure()
			c.logger.Error("External API server error", "status", resp.StatusCode, "url", req.URL.String())
			// For POST requests with a body, body re-reading on retry is not supported
			// because http.Request.Body is consumed after the first read. Callers that
			// need idempotent retry on POST should reconstruct the request themselves.
			continue
		}

		c.circuitBreaker.recordSuccess()
		return body, resp.StatusCode, nil
	}

	return nil, 0, fmt.Errorf("all retries exhausted: %w", lastErr)
}
