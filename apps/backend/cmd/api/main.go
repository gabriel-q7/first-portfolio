package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gabriel-q7/portfolio/backend/configs"
	"github.com/gabriel-q7/portfolio/backend/internal/domain/entity"
	"github.com/gabriel-q7/portfolio/backend/internal/domain/repository"
	aiClient "github.com/gabriel-q7/portfolio/backend/internal/infrastructure/clients/ai"
	extClient "github.com/gabriel-q7/portfolio/backend/internal/infrastructure/clients/external_apis"
	nosqlRepo "github.com/gabriel-q7/portfolio/backend/internal/infrastructure/persistence/nosql"
	sqlRepo "github.com/gabriel-q7/portfolio/backend/internal/infrastructure/persistence/sql"
	"github.com/gabriel-q7/portfolio/backend/internal/infrastructure/queue"
	httpHandler "github.com/gabriel-q7/portfolio/backend/internal/interfaces/http"
	"github.com/gabriel-q7/portfolio/backend/internal/interfaces/http/handler"
	"github.com/gabriel-q7/portfolio/backend/internal/interfaces/terminal"
	"github.com/gabriel-q7/portfolio/backend/internal/interfaces/terminal/commands"
	"github.com/gabriel-q7/portfolio/backend/internal/middleware"
	projectUseCase "github.com/gabriel-q7/portfolio/backend/internal/usecase/project"
	apperrors "github.com/gabriel-q7/portfolio/backend/pkg/errors"
	"github.com/gabriel-q7/portfolio/backend/pkg/logger"
	"github.com/gabriel-q7/portfolio/backend/pkg/metrics"
	"github.com/google/uuid"

	_ "github.com/lib/pq"
)

func main() {
	cfg := configs.Load()
	log := logger.New(cfg.Observability.LogLevel)
	log.Info("starting portfolio backend", "env", cfg.Environment, "port", cfg.Server.Port)

	var m *metrics.Metrics
	if cfg.Observability.MetricsEnabled {
		m = metrics.New()
	} else {
		m = &metrics.Metrics{}
	}

	// Connect to PostgreSQL with retry.
	var projectRepo repository.ProjectRepository
	if cfg.Database.Postgres.DSN != "" {
		db, err := connectPostgres(cfg, log)
		if err != nil {
			log.Error("failed to connect to postgres, falling back to in-memory repo", "error", err)
			projectRepo = newInMemoryProjectRepo()
		} else {
			if err := sqlRepo.InitSchema(db); err != nil {
				log.Error("schema init failed", "error", err)
			}
			projectRepo = sqlRepo.New(db, log)
		}
	} else {
		log.Info("no POSTGRES_DSN set, using in-memory project repository")
		projectRepo = newInMemoryProjectRepo()
	}

	// Connect to Redis or use in-memory cache.
	var cacheRepo repository.CacheRepository
	if cfg.Database.Redis.Addr != "" {
		client, err := nosqlRepo.Connect(nosqlRepo.RedisConfig{
			Addr:     cfg.Database.Redis.Addr,
			Password: cfg.Database.Redis.Password,
			DB:       cfg.Database.Redis.DB,
			PoolSize: cfg.Database.Redis.PoolSize,
		})
		if err != nil {
			log.Error("failed to connect to redis, using in-memory cache", "error", err)
			cacheRepo = newInMemoryCache()
		} else {
			cacheRepo = nosqlRepo.New(client, log)
		}
	} else {
		log.Info("no REDIS_ADDR set, using in-memory cache")
		cacheRepo = newInMemoryCache()
	}

	// Initialize AI client.
	var ai aiClient.AIProvider
	if cfg.AI.APIKey != "" {
		ai = aiClient.NewOpenAIClient(cfg.AI.APIKey, cfg.AI.BaseURL, cfg.AI.RateLimit, log, m)
	} else {
		log.Info("no AI_API_KEY set, using noop AI client")
		ai = &aiClient.NoopAIClient{}
	}
	_ = ai // Used by terminal chat command.

	// Initialize external API client.
	extAPI := extClient.New(extClient.Config{
		BaseURL:    "",
		Timeout:    cfg.ExternalAPIs.Timeout,
		MaxRetries: cfg.ExternalAPIs.MaxRetries,
		UserAgent:  "portfolio-backend/1.0",
	}, log, m)
	_ = extAPI

	// Start job queue.
	q := queue.New(100, 4, log, m)
	ctx, cancel := context.WithCancel(context.Background())
	q.Start(ctx)

	// Initialize use cases.
	projService := projectUseCase.New(projectRepo, cacheRepo, log, m)

	// Initialize handlers.
	healthH := handler.NewHealthHandler("1.0.0", log)
	projectH := handler.NewProjectHandler(projService, log)

	// Initialize terminal command router.
	terminalRouter := terminal.NewRouter()
	commands.RegisterBasicCommands(terminalRouter)
	commands.RegisterProjectCommands(terminalRouter, projService)
	commands.RegisterPortfolioCommands(terminalRouter)
	commands.RegisterDBCommands(terminalRouter, projService)
	commands.RegisterChatCommands(terminalRouter, ai)
	terminalH := terminal.NewHandler(terminalRouter, log)

	// Initialize middleware.
	rl := middleware.NewRateLimiter(cfg.RateLimit.RequestsPerSecond, cfg.RateLimit.BurstSize, log, m)
	authMW := middleware.NewAuthMiddleware(cfg.Auth.APIKeys, cfg.Auth.JWTSecret, log)

	// Build router.
	router := httpHandler.New(httpHandler.Dependencies{
		HealthHandler:   healthH,
		ProjectHandler:  projectH,
		TerminalHandler: terminalH,
		RateLimiter:     rl,
		AuthMiddleware:  authMW,
		Metrics:         m,
		Logger:          log,
		IsProd:          cfg.IsProd(),
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router.SetupRoutes(),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server.
	errCh := make(chan error, 1)
	go func() {
		log.Info("HTTP server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	select {
	case err := <-errCh:
		log.Error("server error", "error", err)
	case sig := <-quit:
		log.Info("shutdown signal received", "signal", sig)
	}

	cancel() // Stop queue workers.

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("server shutdown error", "error", err)
	}

	q.Drain()
	log.Info("server stopped gracefully")
}

func connectPostgres(cfg *configs.Config, log logger.Logger) (*sql.DB, error) {
	var db *sql.DB
	var err error
	for i := 1; i <= 5; i++ {
		db, err = sql.Open("postgres", cfg.Database.Postgres.DSN)
		if err == nil {
			err = db.Ping()
		}
		if err == nil {
			db.SetMaxOpenConns(cfg.Database.Postgres.MaxOpenConns)
			db.SetMaxIdleConns(cfg.Database.Postgres.MaxIdleConns)
			db.SetConnMaxLifetime(cfg.Database.Postgres.ConnMaxLifetime)
			log.Info("connected to postgres")
			return db, nil
		}
		log.Warn("postgres connection attempt failed, retrying",
			"attempt", i, "error", err,
		)
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("failed to connect to postgres after 5 attempts: %w", err)
}

// --- In-memory project repository ---

type inMemoryProjectRepo struct {
	mu       sync.RWMutex
	projects map[uuid.UUID]*entity.Project
}

func newInMemoryProjectRepo() repository.ProjectRepository {
	return &inMemoryProjectRepo{projects: make(map[uuid.UUID]*entity.Project)}
}

func (r *inMemoryProjectRepo) FindByID(_ context.Context, id uuid.UUID) (*entity.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.projects[id]
	if !ok {
		return nil, apperrors.NewNotFoundError(fmt.Sprintf("project %s not found", id), nil)
	}
	return p, nil
}

func (r *inMemoryProjectRepo) FindAll(_ context.Context) ([]*entity.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]*entity.Project, 0, len(r.projects))
	for _, p := range r.projects {
		list = append(list, p)
	}
	return list, nil
}

func (r *inMemoryProjectRepo) FindFeatured(_ context.Context) ([]*entity.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []*entity.Project
	for _, p := range r.projects {
		if p.Featured {
			list = append(list, p)
		}
	}
	return list, nil
}

func (r *inMemoryProjectRepo) Save(_ context.Context, p *entity.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.projects[p.ID] = p
	return nil
}

func (r *inMemoryProjectRepo) Update(_ context.Context, p *entity.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.projects[p.ID]; !ok {
		return apperrors.NewNotFoundError(fmt.Sprintf("project %s not found", p.ID), nil)
	}
	r.projects[p.ID] = p
	return nil
}

func (r *inMemoryProjectRepo) Delete(_ context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.projects[id]; !ok {
		return apperrors.NewNotFoundError(fmt.Sprintf("project %s not found", id), nil)
	}
	delete(r.projects, id)
	return nil
}

// --- In-memory cache ---

type cacheEntry struct {
	value     []byte
	expiresAt time.Time
}

type inMemoryCache struct {
	mu    sync.RWMutex
	store map[string]cacheEntry
}

func newInMemoryCache() repository.CacheRepository {
	c := &inMemoryCache{store: make(map[string]cacheEntry)}
	go c.evict()
	return c
}

func (c *inMemoryCache) evict() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for k, v := range c.store {
			if !v.expiresAt.IsZero() && now.After(v.expiresAt) {
				delete(c.store, k)
			}
		}
		c.mu.Unlock()
	}
}

func (c *inMemoryCache) Get(_ context.Context, key string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.store[key]
	if !ok {
		return nil, nil
	}
	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		return nil, nil
	}
	return entry.value, nil
}

func (c *inMemoryCache) Set(_ context.Context, key string, value []byte, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	e := cacheEntry{value: value}
	if ttl > 0 {
		e.expiresAt = time.Now().Add(ttl)
	}
	c.store[key] = e
	return nil
}

func (c *inMemoryCache) Delete(_ context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, key)
	return nil
}

func (c *inMemoryCache) Exists(_ context.Context, key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.store[key]
	if !ok {
		return false, nil
	}
	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		return false, nil
	}
	return true, nil
}

func (c *inMemoryCache) FlushPattern(_ context.Context, pattern string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k := range c.store {
		if matchPattern(pattern, k) {
			delete(c.store, k)
		}
	}
	return nil
}

// matchPattern performs simple glob-style matching (* matches anything).
func matchPattern(pattern, s string) bool {
	if pattern == "*" {
		return true
	}
	if len(pattern) == 0 {
		return s == ""
	}
	if pattern[0] == '*' {
		rest := pattern[1:]
		for i := 0; i <= len(s); i++ {
			if matchPattern(rest, s[i:]) {
				return true
			}
		}
		return false
	}
	if len(s) == 0 {
		return false
	}
	if pattern[0] == s[0] {
		return matchPattern(pattern[1:], s[1:])
	}
	return false
}
