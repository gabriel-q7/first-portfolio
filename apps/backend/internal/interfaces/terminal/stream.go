package terminal

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gabriel-q7/portfolio/backend/internal/domain/protocol"
	"github.com/gorilla/websocket"
)

// OutputStream is the interface through which command handlers write output.
type OutputStream interface {
	Write(text string) error
	WriteError(text string) error
	Close() error
}

// wsOutputStream implements OutputStream by sending WebSocket messages.
type wsOutputStream struct {
	conn      *websocket.Conn
	requestID string
	mu        sync.Mutex
	closed    bool
}

// NewOutputStream creates an OutputStream backed by a WebSocket connection.
func NewOutputStream(conn *websocket.Conn, requestID string) OutputStream {
	return &wsOutputStream{conn: conn, requestID: requestID}
}

// Write sends an output message to the client.
func (s *wsOutputStream) Write(text string) error {
	return s.send(protocol.NewOutputMessage(s.requestID, text))
}

// WriteError sends an error output message to the client.
func (s *wsOutputStream) WriteError(text string) error {
	return s.send(protocol.NewErrorMessage(s.requestID, text))
}

// Close sends a done message to the client to signal command completion.
func (s *wsOutputStream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	return s.sendLocked(protocol.NewDoneMessage(s.requestID))
}

func (s *wsOutputStream) send(msg protocol.ServerMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sendLocked(msg)
}

func (s *wsOutputStream) sendLocked(msg protocol.ServerMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}
	s.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)) //nolint:errcheck
	return s.conn.WriteMessage(websocket.TextMessage, data)
}
