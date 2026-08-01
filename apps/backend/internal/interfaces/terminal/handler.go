package terminal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gabriel-q7/portfolio/backend/internal/domain/entity"
	"github.com/gabriel-q7/portfolio/backend/internal/domain/protocol"
	"github.com/gabriel-q7/portfolio/backend/pkg/logger"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin, err := url.Parse(r.Header.Get("Origin"))
		return err == nil && origin.Host != "" && strings.EqualFold(origin.Host, r.Host)
	},
}

// Handler manages WebSocket terminal sessions.
type Handler struct {
	router        *CommandRouter
	logger        logger.Logger
	parentContext context.Context
}

// NewHandler creates a new WebSocket terminal Handler.
func NewHandler(router *CommandRouter, log logger.Logger, parent ...context.Context) *Handler {
	parentContext := context.Background()
	if len(parent) > 0 && parent[0] != nil {
		parentContext = parent[0]
	}
	return &Handler{router: router, logger: log, parentContext: parentContext}
}

// ServeHTTP upgrades the HTTP connection to WebSocket and handles the terminal session.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Warn("websocket upgrade failed", "error", err)
		return
	}
	defer conn.Close()
	conn.SetReadLimit(4 << 10)

	session := entity.NewTerminalSession(h.parentContext)
	defer session.Close()

	h.logger.Info("terminal session started", "session_id", session.ID)

	// Send welcome status message.
	welcome := protocol.NewStatusMessage("connected — type 'help' for available commands")
	if data, err := json.Marshal(welcome); err == nil {
		conn.SetWriteDeadline(time.Now().Add(10 * time.Second)) //nolint:errcheck
		_ = conn.WriteMessage(websocket.TextMessage, data)
	}

	// Set up pong handler to keep the connection alive.
	conn.SetReadDeadline(time.Now().Add(60 * time.Second)) //nolint:errcheck
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second)) //nolint:errcheck
		return nil
	})

	// Start ping ticker.
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	// Channel to receive parsed client messages.
	msgCh := make(chan protocol.ClientMessage, 8)
	readErrCh := make(chan error, 1)

	// Read loop in a goroutine.
	go func() {
		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				readErrCh <- err
				return
			}
			var msg protocol.ClientMessage
			if err := json.Unmarshal(data, &msg); err != nil {
				h.logger.Warn("failed to unmarshal client message", "error", err)
				continue
			}
			msgCh <- msg
		}
	}()

	// Track whether a command is currently running.
	var cmdCancel context.CancelFunc

	for {
		select {
		case <-session.Context().Done():
			if cmdCancel != nil {
				cmdCancel()
			}
			_ = conn.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseGoingAway, "server shutting down"),
				time.Now().Add(time.Second),
			)
			return

		case <-pingTicker.C:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second)) //nolint:errcheck
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				h.logger.Debug("ping failed, closing session", "session_id", session.ID)
				return
			}

		case err := <-readErrCh:
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				h.logger.Warn("websocket read error", "session_id", session.ID, "error", err)
			}
			return

		case msg := <-msgCh:
			switch msg.Type {
			case protocol.MessageTypeCancel:
				if cmdCancel != nil {
					cmdCancel()
					cmdCancel = nil
				}

			case protocol.MessageTypeCommand:
				if msg.Command == "" {
					continue
				}
				// Cancel any in-progress command before starting a new one.
				if cmdCancel != nil {
					cmdCancel()
				}

				parsed := Parse(msg.Command)
				out := NewOutputStream(conn, msg.RequestID)

				cmdCtx, cancel := context.WithCancel(session.Context())
				cmdCancel = cancel

				go func(ctx context.Context, p ParsedCommand, o OutputStream, cancelFn context.CancelFunc) {
					defer cancelFn()
					defer func() { _ = o.Close() }()

					h.logger.Debug("executing command", "session_id", session.ID, "command", p.Raw)

					if err := h.router.Dispatch(ctx, p, o); err != nil {
						if ctx.Err() != nil {
							_ = o.WriteError("[cancelled]")
							return
						}
						_ = o.WriteError(err.Error())
					}
				}(cmdCtx, parsed, out, cancel)
				// Reset cmdCancel only after the goroutine finishes; the goroutine
				// holds its own reference so double-calling cancel is safe (idempotent).
				cmdCancel = cancel
			}
		}
	}
}
