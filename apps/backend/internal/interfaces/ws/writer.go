package ws

import (
	"github.com/gabriel-q7/portfolio/backend/internal/terminal/handler"
)

// sessionWriter adapts a *Session to the terminal handler.Writer interface.
type sessionWriter struct {
	s *Session
}

// NewWriter wraps a WebSocket session as a handler.Writer.
func NewWriter(s *Session) handler.Writer {
	return &sessionWriter{s: s}
}

func (w *sessionWriter) Output(line string) error { return w.s.SendOutput(line) }
func (w *sessionWriter) Done() error              { return w.s.SendDone() }
func (w *sessionWriter) Error(msg string) error {
	_ = w.s.SendError(msg)
	return w.s.SendDone()
}
func (w *sessionWriter) Status(msg string) error { return w.s.SendStatus(msg) }
