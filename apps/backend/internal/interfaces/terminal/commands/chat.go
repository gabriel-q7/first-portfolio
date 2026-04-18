package commands

import (
	"context"
	"fmt"
	"strings"

	aiClient "github.com/gabriel-q7/portfolio/backend/internal/infrastructure/clients/ai"
	"github.com/gabriel-q7/portfolio/backend/internal/interfaces/terminal"
)

// RegisterChatCommands registers the chat ask handler.
func RegisterChatCommands(r *terminal.CommandRouter, ai aiClient.AIProvider) {
	r.Register("chat", func(ctx context.Context, cmd terminal.ParsedCommand, out terminal.OutputStream) error {
		switch cmd.Sub {
		case "ask":
			return handleChatAsk(ctx, ai, cmd, out)
		default:
			if cmd.Sub == "" {
				return out.WriteError("  usage: chat ask \"<question>\"")
			}
			return out.WriteError(fmt.Sprintf("  unknown chat sub-command: %s. Use 'chat ask \"<question>\"'.", cmd.Sub))
		}
	})
}

func handleChatAsk(ctx context.Context, ai aiClient.AIProvider, cmd terminal.ParsedCommand, out terminal.OutputStream) error {
	// Reconstruct the question from the raw input after "chat ask"
	question := extractQuestion(cmd.Raw)
	if question == "" {
		return out.WriteError("  usage: chat ask \"<question>\"")
	}

	if err := out.Write("\x1b[2m  thinking...\x1b[0m"); err != nil {
		return err
	}

	result, err := ai.Complete(ctx, question, aiClient.CompletionOptions{
		Model:       "gpt-4o-mini",
		MaxTokens:   512,
		Temperature: 0.7,
		SystemPrompt: "You are an AI assistant embedded in a terminal-style portfolio website. " +
			"Keep your responses concise and formatted for a terminal (plain text, no markdown headers). " +
			"You can use simple formatting like dashes for lists.",
	})
	if err != nil {
		if ctx.Err() != nil {
			return out.WriteError("  [cancelled]")
		}
		return out.WriteError(fmt.Sprintf("  AI error: %s", err.Error()))
	}

	// Write each line of the response separately for a streaming-like feel
	lines := strings.Split(result.Content, "\n")
	for _, l := range lines {
		if ctx.Err() != nil {
			return out.WriteError("  [cancelled]")
		}
		if err := out.Write("  " + l); err != nil {
			return err
		}
	}
	if err := out.Write(""); err != nil {
		return err
	}
	return nil
}

// extractQuestion parses the question from the raw command input.
// It handles: chat ask "some question" or chat ask some question
func extractQuestion(raw string) string {
	// Find the part after "chat ask"
	lower := strings.ToLower(raw)
	idx := strings.Index(lower, "ask")
	if idx < 0 {
		return ""
	}
	rest := strings.TrimSpace(raw[idx+3:])

	// Strip surrounding quotes if present
	if len(rest) >= 2 {
		if (rest[0] == '"' && rest[len(rest)-1] == '"') ||
			(rest[0] == '\'' && rest[len(rest)-1] == '\'') {
			return rest[1 : len(rest)-1]
		}
	}

	return rest
}
