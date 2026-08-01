package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

type Config struct {
	Path         string
	BusyTimeout  time.Duration
	MaxOpenConns int
}

func Open(ctx context.Context, cfg Config) (*sql.DB, error) {
	if cfg.Path == "" {
		return nil, fmt.Errorf("sqlite path is required")
	}
	if cfg.MaxOpenConns < 1 {
		cfg.MaxOpenConns = 1
	}
	if cfg.BusyTimeout <= 0 {
		cfg.BusyTimeout = 5 * time.Second
	}

	directory := filepath.Dir(cfg.Path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return nil, fmt.Errorf("create sqlite directory: %w", err)
	}
	if err := os.Chmod(directory, 0o700); err != nil {
		return nil, fmt.Errorf("secure sqlite directory permissions: %w", err)
	}

	dsn := fmt.Sprintf(
		"file:%s?_pragma=busy_timeout%%28%d%%29&_pragma=foreign_keys%%281%%29&_pragma=journal_mode%%28WAL%%29&_pragma=synchronous%%28NORMAL%%29&_pragma=cache_size%%28-2000%%29&_pragma=journal_size_limit%%288388608%%29&_pragma=temp_store%%282%%29",
		cfg.Path,
		cfg.BusyTimeout.Milliseconds(),
	)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxOpenConns)
	db.SetConnMaxLifetime(0)

	pingCtx, cancel := context.WithTimeout(ctx, cfg.BusyTimeout)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	if err := os.Chmod(cfg.Path, 0o600); err != nil {
		db.Close()
		return nil, fmt.Errorf("secure sqlite file permissions: %w", err)
	}
	return db, nil
}

func InitSchema(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin schema transaction: %w", err)
	}
	defer tx.Rollback()

	const schema = `
	CREATE TABLE IF NOT EXISTS projects (
		id          TEXT PRIMARY KEY,
		name        TEXT NOT NULL,
		description TEXT NOT NULL,
		tech        TEXT NOT NULL DEFAULT '[]',
		url         TEXT NOT NULL DEFAULT '',
		featured    INTEGER NOT NULL DEFAULT 0 CHECK (featured IN (0, 1)),
		created_at  DATETIME NOT NULL,
		updated_at  DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_projects_featured_created
		ON projects(featured, created_at DESC);
	PRAGMA user_version = 1;`
	if _, err := tx.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("apply schema: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit schema: %w", err)
	}
	return nil
}
