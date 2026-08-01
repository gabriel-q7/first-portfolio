package middleware

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gabriel-q7/portfolio/backend/pkg/logger"
	"golang.org/x/time/rate"
)

type visitorState struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	visitors     map[string]*visitorState
	mu           sync.Mutex
	rate         rate.Limit
	burst        int
	maxVisitors  int
	requests     uint64
	log          logger.Logger
	trustedProxy *net.IPNet
}

// NewRateLimiter creates a bounded per-client limiter. Cleanup is performed
// opportunistically so the service does not need a maintenance goroutine.
func NewRateLimiter(
	requestsPerSecond float64,
	burst int,
	maxVisitors int,
	trustedProxyCIDR string,
	log logger.Logger,
) (*RateLimiter, error) {
	if requestsPerSecond <= 0 || burst < 1 || maxVisitors < 1 {
		return nil, fmt.Errorf("rate limit values must be positive")
	}
	var trustedProxy *net.IPNet
	if trustedProxyCIDR != "" {
		_, network, err := net.ParseCIDR(trustedProxyCIDR)
		if err != nil {
			return nil, fmt.Errorf("parse TRUSTED_PROXY_CIDR: %w", err)
		}
		trustedProxy = network
	}
	return &RateLimiter{
		visitors:     make(map[string]*visitorState),
		rate:         rate.Limit(requestsPerSecond),
		burst:        burst,
		maxVisitors:  maxVisitors,
		log:          log,
		trustedProxy: trustedProxy,
	}, nil
}

func (rl *RateLimiter) TrustedProxy() *net.IPNet {
	return rl.trustedProxy
}

func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := ClientIP(r, rl.trustedProxy)
			if !rl.getVisitor(ip).Allow() {
				rl.log.Warn("rate limit exceeded", "ip", ip, "path", r.URL.Path)
				w.Header().Set("Retry-After", "1")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "rate limit exceeded"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (rl *RateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.requests++
	if rl.requests%256 == 0 {
		rl.removeStaleLocked(time.Now().Add(-3 * time.Minute))
	}
	if visitor, exists := rl.visitors[ip]; exists {
		visitor.lastSeen = time.Now()
		return visitor.limiter
	}
	if len(rl.visitors) >= rl.maxVisitors {
		rl.removeOldestLocked()
	}
	visitor := &visitorState{
		limiter:  rate.NewLimiter(rl.rate, rl.burst),
		lastSeen: time.Now(),
	}
	rl.visitors[ip] = visitor
	return visitor.limiter
}

func (rl *RateLimiter) removeStaleLocked(cutoff time.Time) {
	for ip, visitor := range rl.visitors {
		if visitor.lastSeen.Before(cutoff) {
			delete(rl.visitors, ip)
		}
	}
}

func (rl *RateLimiter) removeOldestLocked() {
	var oldestIP string
	var oldestTime time.Time
	for ip, visitor := range rl.visitors {
		if oldestIP == "" || visitor.lastSeen.Before(oldestTime) {
			oldestIP = ip
			oldestTime = visitor.lastSeen
		}
	}
	delete(rl.visitors, oldestIP)
}

func ClientIP(r *http.Request, trustedProxy *net.IPNet) string {
	remoteIP := parseRemoteIP(r.RemoteAddr)
	if trustedProxy != nil && remoteIP != nil && trustedProxy.Contains(remoteIP) {
		if realIP := net.ParseIP(r.Header.Get("X-Real-IP")); realIP != nil {
			return realIP.String()
		}
	}
	if remoteIP != nil {
		return remoteIP.String()
	}
	return r.RemoteAddr
}

func parseRemoteIP(remoteAddr string) net.IP {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		return net.ParseIP(host)
	}
	return net.ParseIP(remoteAddr)
}

func APIKeyRateLimit(keys map[string]float64) func(http.Handler) http.Handler {
	limiters := make(map[string]*rate.Limiter, len(keys))
	var mu sync.RWMutex
	for key, requestsPerSecond := range keys {
		limiters[key] = rate.NewLimiter(rate.Limit(requestsPerSecond), int(requestsPerSecond)+1)
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-API-Key")
			mu.RLock()
			limiter, ok := limiters[key]
			mu.RUnlock()
			if ok && !limiter.Allow() {
				w.Header().Set("Retry-After", strconv.Itoa(1))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "api key rate limit exceeded"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
