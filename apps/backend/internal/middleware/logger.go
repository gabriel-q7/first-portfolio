package middleware

import (
	"bufio"
	"net"
	"net/http"
	"time"

	"github.com/gabriel-q7/portfolio/backend/pkg/logger"
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
	rw.statusCode = http.StatusSwitchingProtocols
	return hijacker.Hijack()
}

// RequestLogger logs each request as structured JSON.
func RequestLogger(log logger.Logger, trustedProxy *net.IPNet) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := newResponseWriter(w)
			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			log.Info("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", ClientIP(r, trustedProxy),
				"request_id", GetRequestID(r.Context()),
				"status", rw.statusCode,
				"duration_ms", duration.Milliseconds(),
			)
		})
	}
}
