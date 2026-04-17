package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	aiClient "github.com/gabriel-q7/portfolio/backend/internal/infrastructure/clients/ai"
)

// ChatHandler handles the "chat ask" command by calling the AI provider.
type ChatHandler struct {
	ai aiClient.AIProvider
}

// NewChatHandler creates a ChatHandler backed by the given AI provider.
func NewChatHandler(ai aiClient.AIProvider) *ChatHandler {
	return &ChatHandler{ai: ai}
}

func (h *ChatHandler) Handle(ctx context.Context, args []string, w Writer) error {
	// Expect subcommand: ask "<question>"
	if len(args) == 0 || args[0] != "ask" {
		return w.Error("  usage: chat ask \"<question>\"")
	}

	// Everything after "ask" is the question (may have been split by the parser)
	question := strings.Join(args[1:], " ")
	// Strip surrounding quotes if present
	question = strings.Trim(question, "\"'")
	if strings.TrimSpace(question) == "" {
		return w.Error("  usage: chat ask \"<question>\"")
	}

	if err := w.Status("  Thinking…"); err != nil {
		return err
	}

	// Use a timeout so a stalled AI call does not hang the session forever.
	aiCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	result, err := h.ai.Complete(aiCtx, question, aiClient.CompletionOptions{
		SystemPrompt: "You are a helpful assistant embedded in a developer portfolio terminal. " +
			"Answer concisely in plain text. Keep answers under 200 words.",
		MaxTokens:   300,
		Temperature: 0.7,
	})
	if err != nil {
		return w.Error(fmt.Sprintf("  chat error: %v", err))
	}

	if err := w.Output(""); err != nil {
		return err
	}
	// Stream line by line so the terminal renders progressively
	for _, line := range strings.Split(result.Content, "\n") {
		if err := w.Output("  " + line); err != nil {
			return err
		}
	}
	if err := w.Output(""); err != nil {
		return err
	}
	return w.Done()
}
