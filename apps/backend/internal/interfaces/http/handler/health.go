package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gabriel-q7/portfolio/backend/pkg/logger"
)

// HealthHandler serves health check endpoints.
type HealthHandler struct {
	version   string
	startTime time.Time
	logger    logger.Logger
	database  databasePinger
}

type databasePinger interface {
	PingContext(context.Context) error
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(version string, log logger.Logger, database ...databasePinger) *HealthHandler {
	var db databasePinger
	if len(database) > 0 {
		db = database[0]
	}
	return &HealthHandler{version: version, startTime: time.Now().UTC(), logger: log, database: db}
}

// Health returns the full health status.
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(h.startTime).Round(time.Second).String()
	respondJSON(w, http.StatusOK, map[string]string{
		"status":    "ok",
		"version":   h.version,
		"uptime":    uptime,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// Live is the liveness probe.
func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "alive"})
}

// Ready is the readiness probe.
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	if h.database != nil {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()
		if err := h.database.PingContext(ctx); err != nil {
			h.logger.Error("readiness check failed", "error", err)
			respondJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
			return
		}
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

// respondJSON writes a JSON response with the given status code.
func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
