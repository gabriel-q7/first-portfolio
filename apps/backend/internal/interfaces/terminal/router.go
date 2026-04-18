package terminal

import (
	"context"
	"fmt"
)

// HandlerFunc is a function that executes a terminal command.
// It receives the parsed command, an output stream, and a context for cancellation.
type HandlerFunc func(ctx context.Context, cmd ParsedCommand, out OutputStream) error

// CommandRouter maps command names to handler functions.
type CommandRouter struct {
	handlers map[string]HandlerFunc
}

// NewRouter creates a new CommandRouter.
func NewRouter() *CommandRouter {
	return &CommandRouter{handlers: make(map[string]HandlerFunc)}
}

// Register adds a command handler to the router.
func (r *CommandRouter) Register(name string, fn HandlerFunc) {
	r.handlers[name] = fn
}

// Dispatch looks up and executes the handler for the given command.
// Returns an error if the command is not found.
func (r *CommandRouter) Dispatch(ctx context.Context, cmd ParsedCommand, out OutputStream) error {
	fn, ok := r.handlers[cmd.Name]
	if !ok {
		return fmt.Errorf("command not found: %s. Type 'help' for available commands.", cmd.Name)
	}
	return fn(ctx, cmd, out)
}
