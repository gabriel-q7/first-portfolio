package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gabriel-q7/portfolio/backend/configs"
	"github.com/gabriel-q7/portfolio/backend/internal/domain/repository"
	memoryCache "github.com/gabriel-q7/portfolio/backend/internal/infrastructure/cache"
	aiClient "github.com/gabriel-q7/portfolio/backend/internal/infrastructure/clients/ai"
	sqliteRepo "github.com/gabriel-q7/portfolio/backend/internal/infrastructure/persistence/sqlite"
	httpHandler "github.com/gabriel-q7/portfolio/backend/internal/interfaces/http"
	"github.com/gabriel-q7/portfolio/backend/internal/interfaces/http/handler"
	"github.com/gabriel-q7/portfolio/backend/internal/interfaces/terminal"
	"github.com/gabriel-q7/portfolio/backend/internal/interfaces/terminal/commands"
	"github.com/gabriel-q7/portfolio/backend/internal/middleware"
	projectUseCase "github.com/gabriel-q7/portfolio/backend/internal/usecase/project"
	"github.com/gabriel-q7/portfolio/backend/pkg/logger"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		runHealthcheck()
		return
	}

	cfg := configs.Load()
	log := logger.New(cfg.Observability.LogLevel)
	log.Info("starting portfolio backend", "env", cfg.Environment, "port", cfg.Server.Port, "version", version)
	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

	db, err := sqliteRepo.Open(appCtx, sqliteRepo.Config{
		Path:         cfg.Database.SQLite.Path,
		BusyTimeout:  cfg.Database.SQLite.BusyTimeout,
		MaxOpenConns: cfg.Database.SQLite.MaxOpenConns,
	})
	if err != nil {
		log.Error("sqlite initialization failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := sqliteRepo.InitSchema(appCtx, db); err != nil {
		log.Error("sqlite schema initialization failed", "error", err)
		os.Exit(1)
	}
	log.Info("sqlite ready", "path", cfg.Database.SQLite.Path, "journal_mode", "WAL")

	projectRepo := sqliteRepo.NewProjectRepository(db, log)
	cacheRepo := memoryCache.New(256)

	var ai aiClient.AIProvider
	if cfg.AI.APIKey != "" {
		ai = aiClient.NewOpenAIClient(cfg.AI.APIKey, cfg.AI.BaseURL, cfg.AI.RateLimit, log)
	} else {
		log.Info("AI_API_KEY not set; chat command is disabled")
		ai = &aiClient.NoopAIClient{}
	}

	projService := projectUseCase.New(projectRepo, cacheRepo, log)
	healthH := handler.NewHealthHandler(version, log, db)
	projectH := handler.NewProjectHandler(projService, log)

	terminalRouter := terminal.NewRouter()
	commands.RegisterBasicCommands(terminalRouter)
	commands.RegisterProjectCommands(terminalRouter, projService)
	commands.RegisterPortfolioCommands(terminalRouter)
	commands.RegisterDBCommands(terminalRouter, projService)
	commands.RegisterChatCommands(terminalRouter, ai)
	terminalH := terminal.NewHandler(terminalRouter, log, appCtx)

	rl, err := middleware.NewRateLimiter(
		cfg.RateLimit.RequestsPerSecond,
		cfg.RateLimit.BurstSize,
		cfg.RateLimit.MaxVisitors,
		cfg.Server.TrustedProxyCIDR,
		log,
	)
	if err != nil {
		log.Error("invalid trusted proxy configuration", "error", err)
		os.Exit(1)
	}
	authMW := middleware.NewAuthMiddleware(cfg.Auth.APIKeys, log)

	router := httpHandler.New(httpHandler.Dependencies{
		HealthHandler:   healthH,
		ProjectHandler:  projectH,
		TerminalHandler: terminalH,
		RateLimiter:     rl,
		AuthMiddleware:  authMW,
		Logger:          log,
		TrustedProxy:    rl.TrustedProxy(),
		RequestMaxBytes: cfg.Server.RequestMaxBytes,
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           router.SetupRoutes(),
		ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
		MaxHeaderBytes:    cfg.Server.MaxHeaderBytes,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Info("HTTP server listening", "addr", srv.Addr)
		if serveErr := srv.ListenAndServe(); serveErr != nil && serveErr != http.ErrServerClosed {
			errCh <- serveErr
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	select {
	case serveErr := <-errCh:
		log.Error("server error", "error", serveErr)
	case sig := <-quit:
		log.Info("shutdown signal received", "signal", sig)
	}

	appCancel()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("server shutdown error", "error", err)
	}

	if err := checkpointSQLite(db); err != nil {
		log.Warn("sqlite WAL checkpoint failed during shutdown", "error", err)
	}
	log.Info("server stopped gracefully")
}

func runHealthcheck() {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://127.0.0.1:" + getEnv("SERVER_PORT", "8080") + "/health/ready")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "healthcheck returned %d\n", resp.StatusCode)
		os.Exit(1)
	}
}

func checkpointSQLite(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := db.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)")
	return err
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

var _ repository.ProjectRepository = (*sqliteRepo.ProjectRepository)(nil)
