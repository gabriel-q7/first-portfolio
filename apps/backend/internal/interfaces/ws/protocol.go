package ws

// MessageType identifies the role of a WebSocket message.
type MessageType string

const (
	// Client → Server
	TypeCommand MessageType = "command"

	// Server → Client
	TypeOutput MessageType = "output"
	TypeDone   MessageType = "done"
	TypeError  MessageType = "error"
	TypeStatus MessageType = "status"
)

// ClientMessage is a message sent from the browser to the server.
type ClientMessage struct {
	Type    MessageType `json:"type"`
	Payload string      `json:"payload"`
}

// ServerMessage is a message sent from the server to the browser.
type ServerMessage struct {
	Type    MessageType `json:"type"`
	Payload string      `json:"payload"`
}
