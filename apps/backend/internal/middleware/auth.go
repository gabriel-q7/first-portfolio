package middleware

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"net/http"

	"github.com/gabriel-q7/portfolio/backend/pkg/logger"
)

type authMiddleware struct {
	apiKeys [][]byte
	log     logger.Logger
}

// ExportedAuthMiddleware is the exported alias used by the router.
type ExportedAuthMiddleware = authMiddleware

// NewAuthMiddleware creates an auth middleware instance.
func NewAuthMiddleware(apiKeys []string, log logger.Logger) *authMiddleware {
	keys := make([][]byte, 0, len(apiKeys))
	for _, k := range apiKeys {
		if k != "" {
			keys = append(keys, []byte(k))
		}
	}
	return &authMiddleware{
		apiKeys: keys,
		log:     log,
	}
}

type contextKeyAPIKey struct{}

// APIKeyAuth validates the X-API-Key header and rejects unauthenticated requests.
func (a *authMiddleware) APIKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-API-Key")
		if key == "" || !a.validAPIKey(key) {
			a.log.Warn("unauthorized request", "path", r.URL.Path, "remote_addr", r.RemoteAddr)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}
		ctx := context.WithValue(r.Context(), contextKeyAPIKey{}, key)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth validates the API key if provided, but allows requests through without one.
func (a *authMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-API-Key")
		if key != "" {
			if !a.validAPIKey(key) {
				a.log.Warn("invalid optional API key", "path", r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
				return
			}
			ctx := context.WithValue(r.Context(), contextKeyAPIKey{}, key)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

func (a *authMiddleware) validAPIKey(candidate string) bool {
	candidateBytes := []byte(candidate)
	valid := 0
	for _, expected := range a.apiKeys {
		if len(candidateBytes) == len(expected) {
			valid |= subtle.ConstantTimeCompare(candidateBytes, expected)
		}
	}
	return valid == 1
}

// GetAPIKey retrieves the API key stored in the context.
func GetAPIKey(ctx context.Context) string {
	if key, ok := ctx.Value(contextKeyAPIKey{}).(string); ok {
		return key
	}
	return ""
}
