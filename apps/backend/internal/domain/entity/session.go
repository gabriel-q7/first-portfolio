package entity

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

// TerminalSession represents an active WebSocket terminal session.
type TerminalSession struct {
	ID        uuid.UUID
	CreatedAt time.Time

	ctx    context.Context
	cancel context.CancelFunc

	mu     sync.Mutex
	closed bool
}

// NewTerminalSession creates a new terminal session with a cancellable context.
func NewTerminalSession(parent context.Context) *TerminalSession {
	ctx, cancel := context.WithCancel(parent)
	return &TerminalSession{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Context returns the session's context.
func (s *TerminalSession) Context() context.Context {
	return s.ctx
}

// Cancel cancels the session's context, signalling any running command to stop.
func (s *TerminalSession) Cancel() {
	s.cancel()
}

// Close closes the session, cancelling any running command.
func (s *TerminalSession) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.closed {
		s.closed = true
		s.cancel()
	}
}

// IsClosed returns true if the session has been closed.
func (s *TerminalSession) IsClosed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed
}
