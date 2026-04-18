package http

import (
	"net/http"
	"time"

	"github.com/gabriel-q7/portfolio/backend/internal/interfaces/http/handler"
	"github.com/gabriel-q7/portfolio/backend/internal/interfaces/terminal"
	"github.com/gabriel-q7/portfolio/backend/internal/middleware"
	"github.com/gabriel-q7/portfolio/backend/pkg/logger"
	"github.com/gabriel-q7/portfolio/backend/pkg/metrics"
)

// Dependencies holds all handler and middleware dependencies for the router.
type Dependencies struct {
	HealthHandler   *handler.HealthHandler
	ProjectHandler  *handler.ProjectHandler
	TerminalHandler *terminal.Handler
	RateLimiter     *middleware.RateLimiter
	AuthMiddleware  *middleware.ExportedAuthMiddleware
	Metrics         *metrics.Metrics
	Logger          logger.Logger
	IsProd          bool
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

	// Metrics (no auth).
	mux.Handle("GET /metrics", ro.deps.Metrics.Handler())

	// Public read routes.
	mux.HandleFunc("GET /api/v1/projects", ro.deps.ProjectHandler.List)
	mux.HandleFunc("GET /api/v1/projects/{id}", ro.deps.ProjectHandler.GetByID)

	// Authenticated write routes.
	mux.Handle("POST /api/v1/projects",
		ro.deps.AuthMiddleware.APIKeyAuth(http.HandlerFunc(ro.deps.ProjectHandler.Create)),
	)

	// WebSocket terminal endpoint (bypasses body-size and timeout middleware via hijacking).
	if ro.deps.TerminalHandler != nil {
		mux.Handle("GET /ws", ro.deps.TerminalHandler)
	}

	// Build middleware chain.
	var chain http.Handler = mux
	chain = ro.deps.RateLimiter.Middleware()(chain)

	if ro.deps.IsProd {
		chain = middleware.SecureHeadersProd(chain)
	} else {
		chain = middleware.SecureHeaders(chain)
	}

	chain = middleware.CORS(middleware.DefaultCORSConfig())(chain)
	chain = middleware.RequestLogger(ro.deps.Logger, ro.deps.Metrics)(chain)
	chain = middleware.MaxBodySize(1 << 20)(chain) // 1 MB
	chain = middleware.Timeout(30 * time.Second)(chain)
	chain = middleware.RequestID(chain)

	return chain
}
