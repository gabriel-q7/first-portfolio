package protocol

import "time"

// MessageType identifies the kind of WebSocket message.
type MessageType string

const (
	// MessageTypeCommand is sent by the client to execute a terminal command.
	MessageTypeCommand MessageType = "command"
	// MessageTypeOutput is sent by the server with incremental command output.
	MessageTypeOutput MessageType = "output"
	// MessageTypeDone is sent by the server when a command finishes.
	MessageTypeDone MessageType = "done"
	// MessageTypeError is sent by the server when a command fails.
	MessageTypeError MessageType = "error"
	// MessageTypeStatus is sent by the server with connection/session status.
	MessageTypeStatus MessageType = "status"
	// MessageTypeCancel is sent by the client to cancel a running command.
	MessageTypeCancel MessageType = "cancel"
)

// ClientMessage is a message sent from the client to the server.
type ClientMessage struct {
	Type      MessageType `json:"type"`
	RequestID string      `json:"request_id"`
	Command   string      `json:"command,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// ServerMessage is a message sent from the server to the client.
type ServerMessage struct {
	Type      MessageType `json:"type"`
	RequestID string      `json:"request_id,omitempty"`
	Content   string      `json:"content,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// NewOutputMessage creates a server output message.
func NewOutputMessage(requestID, content string) ServerMessage {
	return ServerMessage{
		Type:      MessageTypeOutput,
		RequestID: requestID,
		Content:   content,
		Timestamp: time.Now().UTC(),
	}
}

// NewDoneMessage creates a server done message.
func NewDoneMessage(requestID string) ServerMessage {
	return ServerMessage{
		Type:      MessageTypeDone,
		RequestID: requestID,
		Timestamp: time.Now().UTC(),
	}
}

// NewErrorMessage creates a server error message.
func NewErrorMessage(requestID, errMsg string) ServerMessage {
	return ServerMessage{
		Type:      MessageTypeError,
		RequestID: requestID,
		Content:   errMsg,
		Timestamp: time.Now().UTC(),
	}
}

// NewStatusMessage creates a server status message.
func NewStatusMessage(content string) ServerMessage {
	return ServerMessage{
		Type:      MessageTypeStatus,
		Content:   content,
		Timestamp: time.Now().UTC(),
	}
}
