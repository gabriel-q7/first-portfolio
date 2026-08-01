package http

import (
	"net"
	"net/http"
	"time"

	"github.com/gabriel-q7/portfolio/backend/internal/interfaces/http/handler"
	"github.com/gabriel-q7/portfolio/backend/internal/interfaces/terminal"
	"github.com/gabriel-q7/portfolio/backend/internal/middleware"
	"github.com/gabriel-q7/portfolio/backend/pkg/logger"
)

// Dependencies holds all handler and middleware dependencies for the router.
type Dependencies struct {
	HealthHandler   *handler.HealthHandler
	ProjectHandler  *handler.ProjectHandler
	TerminalHandler *terminal.Handler
	RateLimiter     *middleware.RateLimiter
	AuthMiddleware  *middleware.ExportedAuthMiddleware
	Logger          logger.Logger
	TrustedProxy    *net.IPNet
	RequestMaxBytes int64
}

// Router assembles the HTTP routing table and middleware chain.
type Router struct {
	deps Dependencies
}

// New creates a Router with the given dependencies.
func New(deps Dependencies) *Router {
	return &Router{deps: deps}
}

// SetupRoutes returns the fully wired http.Handler.
func (ro *Router) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Health routes (no auth).
	mux.HandleFunc("GET /health", ro.deps.HealthHandler.Health)
	mux.HandleFunc("GET /health/live", ro.deps.HealthHandler.Live)
	mux.HandleFunc("GET /health/ready", ro.deps.HealthHandler.Ready)
	mux.HandleFunc("GET /api/health", ro.deps.HealthHandler.Health)

	// Public read routes.
	mux.HandleFunc("GET /api/v1/projects", ro.deps.ProjectHandler.List)
	mux.HandleFunc("GET /api/v1/projects/{id}", ro.deps.ProjectHandler.GetByID)

	// Authenticated write routes.
	mux.Handle("POST /api/v1/projects",
		ro.deps.AuthMiddleware.APIKeyAuth(http.HandlerFunc(ro.deps.ProjectHandler.Create)),
	)

	// WebSocket terminal endpoint.
	if ro.deps.TerminalHandler != nil {
		mux.Handle("GET /ws", ro.deps.TerminalHandler)
	}

	// Build middleware chain.
	var chain http.Handler = mux
	chain = ro.deps.RateLimiter.Middleware()(chain)
	chain = middleware.RequestLogger(ro.deps.Logger, ro.deps.TrustedProxy)(chain)
	chain = middleware.MaxBodySize(ro.deps.RequestMaxBytes)(chain)
	chain = middleware.Timeout(30 * time.Second)(chain)
	chain = middleware.RequestID(chain)

	return chain
}
