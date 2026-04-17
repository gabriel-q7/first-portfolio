package ws

import (
	"encoding/json"
	"net/http"

	"github.com/gabriel-q7/portfolio/backend/internal/terminal"
	"github.com/gabriel-q7/portfolio/backend/pkg/logger"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
	// Allow all origins in development; in production restrict via CORS middleware.
	CheckOrigin: func(r *http.Request) bool { return true },
}

// TerminalHandler upgrades HTTP connections to WebSocket and drives terminal sessions.
type TerminalHandler struct {
	router *terminal.Router
	logger logger.Logger
}

// NewTerminalHandler creates a TerminalHandler with the given command router.
func NewTerminalHandler(router *terminal.Router, log logger.Logger) *TerminalHandler {
	return &TerminalHandler{router: router, logger: log}
}

// ServeHTTP upgrades the request to a WebSocket connection and enters the read loop.
func (h *TerminalHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Warn("websocket upgrade failed", "error", err)
		return
	}
	defer conn.Close()

	sess := newSession(conn)
	h.logger.Info("terminal session started", "remote", r.RemoteAddr)

	// Send a status line so the client knows the connection is live.
	_ = sess.SendStatus("connected")

	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				h.logger.Warn("websocket read error", "error", err)
			}
			break
		}

		var msg ClientMessage
		if jsonErr := json.Unmarshal(raw, &msg); jsonErr != nil {
			h.logger.Debug("malformed client message", "error", jsonErr)
			_ = sess.SendError("malformed message")
			continue
		}

		if msg.Type != TypeCommand {
			h.logger.Debug("unexpected message type", "type", msg.Type)
			continue
		}

		// Each command runs in its own cancellable context so it can be interrupted.
		cmdCtx, cancel := sess.startCommand(r.Context())

		go func(payload string) {
			defer cancel()
			cmd := terminal.Parse(payload)
			w := NewWriter(sess)
			if dispatchErr := h.router.Dispatch(cmdCtx, cmd, w); dispatchErr != nil {
				h.logger.Debug("command dispatch error", "error", dispatchErr, "command", payload)
			}
		}(msg.Payload)
	}

	sess.cancelCommand()
	h.logger.Info("terminal session ended", "remote", r.RemoteAddr)
}
