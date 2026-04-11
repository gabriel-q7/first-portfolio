package handler

import (
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
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(version string, log logger.Logger) *HealthHandler {
	return &HealthHandler{version: version, startTime: time.Now().UTC(), logger: log}
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
	respondJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

// respondJSON writes a JSON response with the given status code.
func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
