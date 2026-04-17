package ws

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

// Session represents one active WebSocket connection with its own command lifecycle.
type Session struct {
	conn *websocket.Conn

	mu     sync.Mutex
	cancel context.CancelFunc // cancels the currently running command, if any
}

// newSession creates a Session wrapping the given connection.
func newSession(conn *websocket.Conn) *Session {
	return &Session{conn: conn}
}

// Send writes a ServerMessage to the WebSocket connection.
// It is safe to call from multiple goroutines.
func (s *Session) Send(msg ServerMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.conn.WriteMessage(websocket.TextMessage, data)
}

// SendOutput is a convenience wrapper that sends a TypeOutput message.
func (s *Session) SendOutput(line string) error {
	return s.Send(ServerMessage{Type: TypeOutput, Payload: line})
}

// SendDone signals that a command has finished successfully.
func (s *Session) SendDone() error {
	return s.Send(ServerMessage{Type: TypeDone, Payload: ""})
}

// SendError sends a TypeError message to the client.
func (s *Session) SendError(msg string) error {
	return s.Send(ServerMessage{Type: TypeError, Payload: msg})
}

// SendStatus sends an informational status line (e.g. "connecting…").
func (s *Session) SendStatus(msg string) error {
	return s.Send(ServerMessage{Type: TypeStatus, Payload: msg})
}

// startCommand sets a new active command cancel function,
// cancelling any previously running command first.
func (s *Session) startCommand(ctx context.Context) (context.Context, context.CancelFunc) {
	s.mu.Lock()
	if s.cancel != nil {
		s.cancel()
	}
	cmdCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.mu.Unlock()
	return cmdCtx, cancel
}

// cancelCommand interrupts the currently running command, if any.
func (s *Session) cancelCommand() {
	s.mu.Lock()
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	s.mu.Unlock()
}
