package sqlite

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gabriel-q7/portfolio/backend/internal/domain/entity"
	"github.com/gabriel-q7/portfolio/backend/pkg/logger"
)

func TestProjectRepositoryRoundTrip(t *testing.T) {
	db := openTestDatabase(t)
	repository := NewProjectRepository(db, logger.New("error"))
	ctx := context.Background()

	project := entity.NewProject(
		"SQLite",
		"Persistent project",
		[]string{"Go", "SQLite"},
		"https://example.com",
	)
	if err := repository.Save(ctx, project); err != nil {
		t.Fatalf("save project: %v", err)
	}

	stored, err := repository.FindByID(ctx, project.ID)
	if err != nil {
		t.Fatalf("find project: %v", err)
	}
	if stored.Name != project.Name || len(stored.Tech) != 2 || stored.Tech[1] != "SQLite" {
		t.Fatalf("unexpected stored project: %#v", stored)
	}

	project.Update("Updated", "Updated description", []string{"Go"}, "https://example.org")
	if err := repository.Update(ctx, project); err != nil {
		t.Fatalf("update project: %v", err)
	}
	if err := repository.Delete(ctx, project.ID); err != nil {
		t.Fatalf("delete project: %v", err)
	}
	if _, err := repository.FindByID(ctx, project.ID); err == nil {
		t.Fatal("expected deleted project to be missing")
	}
}

func TestSQLiteProductionPragmas(t *testing.T) {
	db := openTestDatabase(t)

	assertPragma(t, db, "journal_mode", "wal")
	assertPragma(t, db, "foreign_keys", "1")
	assertPragma(t, db, "busy_timeout", "5000")
	assertPragma(t, db, "synchronous", "1")
}

func TestSQLitePermissions(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "portfolio.db")
	db, err := Open(context.Background(), Config{
		Path:         path,
		BusyTimeout:  5 * time.Second,
		MaxOpenConns: 2,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	directoryInfo, err := os.Stat(directory)
	if err != nil {
		t.Fatalf("stat sqlite directory: %v", err)
	}
	if directoryInfo.Mode().Perm() != 0o700 {
		t.Fatalf("expected sqlite directory mode 0700, got %04o", directoryInfo.Mode().Perm())
	}
	databaseInfo, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat sqlite database: %v", err)
	}
	if databaseInfo.Mode().Perm() != 0o600 {
		t.Fatalf("expected sqlite database mode 0600, got %04o", databaseInfo.Mode().Perm())
	}
}

func openTestDatabase(t *testing.T) *sql.DB {
	t.Helper()
	db, err := Open(context.Background(), Config{
		Path:         filepath.Join(t.TempDir(), "portfolio.db"),
		BusyTimeout:  5 * time.Second,
		MaxOpenConns: 2,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := InitSchema(context.Background(), db); err != nil {
		t.Fatalf("initialize schema: %v", err)
	}
	return db
}

func assertPragma(t *testing.T, db *sql.DB, pragma, expected string) {
	t.Helper()
	var actual string
	if err := db.QueryRow("PRAGMA " + pragma).Scan(&actual); err != nil {
		t.Fatalf("read PRAGMA %s: %v", pragma, err)
	}
	if actual != expected {
		t.Fatalf("PRAGMA %s: expected %q, got %q", pragma, expected, actual)
	}
}
