package terminal

import (
	"context"
	"fmt"
	"strings"

	"github.com/gabriel-q7/portfolio/backend/internal/terminal/handler"
)

// Router dispatches parsed commands to the appropriate Handler.
type Router struct {
	handlers map[string]handler.Handler
}

// NewRouter creates an empty Router.
func NewRouter() *Router {
	return &Router{handlers: make(map[string]handler.Handler)}
}

// Register adds a handler under a command name (case-insensitive).
func (r *Router) Register(name string, h handler.Handler) {
	r.handlers[strings.ToLower(name)] = h
}

// Dispatch routes the command to the correct handler and executes it.
// Unknown commands produce a single error line followed by done.
func (r *Router) Dispatch(ctx context.Context, cmd ParsedCommand, w handler.Writer) error {
	if cmd.Name == "" {
		return w.Done()
	}

	h, ok := r.handlers[cmd.Name]
	if !ok {
		return w.Error(fmt.Sprintf("  command not found: %s — type 'help' for available commands", cmd.Name))
	}

	return h.Handle(ctx, cmd.Args, w)
}
