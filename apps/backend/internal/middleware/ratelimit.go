package middleware

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gabriel-q7/portfolio/backend/pkg/logger"
	"github.com/gabriel-q7/portfolio/backend/pkg/metrics"
	"golang.org/x/time/rate"
)

type visitorState struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter enforces per-IP request rate limits.
type RateLimiter struct {
	visitors map[string]*visitorState
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
	log      logger.Logger
	metrics  *metrics.Metrics
}

// NewRateLimiter creates a RateLimiter and starts the cleanup goroutine.
func NewRateLimiter(requestsPerSec float64, burst int, log logger.Logger, m *metrics.Metrics) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitorState),
		rate:     rate.Limit(requestsPerSec),
		burst:    burst,
		log:      log,
		metrics:  m,
	}
	go rl.cleanupVisitors()
	return rl
}

// Middleware returns an http.Handler middleware that rate limits by IP.
func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			limiter := rl.getVisitor(ip)
			if !limiter.Allow() {
				rl.log.Warn("rate limit exceeded", "ip", ip, "path", r.URL.Path)
				retryAfter := strconv.FormatFloat(1/float64(rl.rate), 'f', 0, 64)
				w.Header().Set("Retry-After", retryAfter)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error": "rate limit exceeded",
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (rl *RateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	v, exists := rl.visitors[ip]
	if !exists {
		v = &visitorState{limiter: rate.NewLimiter(rl.rate, rl.burst)}
		rl.visitors[ip] = v
	}
	v.lastSeen = time.Now()
	return v.limiter
}

func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// APIKeyRateLimit returns a middleware that rate limits per API key.
func APIKeyRateLimit(keys map[string]float64) func(http.Handler) http.Handler {
	limiters := make(map[string]*rate.Limiter, len(keys))
	var mu sync.RWMutex
	for k, rps := range keys {
		limiters[k] = rate.NewLimiter(rate.Limit(rps), int(rps)+1)
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-API-Key")
			mu.RLock()
			limiter, ok := limiters[key]
			mu.RUnlock()
			if ok && !limiter.Allow() {
				retryAfter := strconv.Itoa(1)
				w.Header().Set("Retry-After", retryAfter)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "api key rate limit exceeded"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
