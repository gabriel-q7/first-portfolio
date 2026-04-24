package middleware

import (
	"bufio"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gabriel-q7/portfolio/backend/pkg/logger"
	"github.com/gabriel-q7/portfolio/backend/pkg/metrics"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Hijack implements http.Hijacker for WebSocket support.
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return hijacker.Hijack()
}

// RequestLogger logs each request and records metrics.
func RequestLogger(log logger.Logger, m *metrics.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := newResponseWriter(w)
			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			statusStr := strconv.Itoa(rw.statusCode)

			log.Info("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
				"request_id", GetRequestID(r.Context()),
				"status", rw.statusCode,
				"duration_ms", duration.Milliseconds(),
			)

			if m != nil {
				m.RecordHTTP(r.Method, r.URL.Path, statusStr, duration)
			}
		})
	}
}
