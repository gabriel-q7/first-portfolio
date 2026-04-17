package handler

import (
	"context"
)

// Writer is the abstraction given to every command handler so it can
// stream lines back to the client without knowing about WebSocket internals.
type Writer interface {
	// Output sends a single output line.
	Output(line string) error
	// Done signals successful completion.
	Done() error
	// Error sends an error message and signals completion.
	Error(msg string) error
	// Status sends an informational status line.
	Status(msg string) error
}

// Handler is the interface every command handler must implement.
type Handler interface {
	// Handle executes the command and streams output via w.
	// Implementations must respect ctx.Done() for cancellation.
	Handle(ctx context.Context, args []string, w Writer) error
}
