package sql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gabriel-q7/portfolio/backend/internal/domain/entity"
	"github.com/gabriel-q7/portfolio/backend/internal/domain/repository"
	apperrors "github.com/gabriel-q7/portfolio/backend/pkg/errors"
	"github.com/gabriel-q7/portfolio/backend/pkg/logger"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type postgresProjectRepo struct {
	db     *sql.DB
	logger logger.Logger
}

// New returns a PostgreSQL-backed ProjectRepository.
func New(db *sql.DB, log logger.Logger) repository.ProjectRepository {
	return &postgresProjectRepo{db: db, logger: log}
}

const baseSelect = `SELECT id, name, description, tech, url, featured, created_at, updated_at FROM projects`

// FindAll returns all projects ordered by creation date.
func (r *postgresProjectRepo) FindAll(ctx context.Context) ([]*entity.Project, error) {
	rows, err := r.db.QueryContext(ctx, baseSelect+` ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("FindAll query: %w", err)
	}
	defer rows.Close()
	return scanProjects(rows)
}

// FindByID returns a project by its UUID.
func (r *postgresProjectRepo) FindByID(ctx context.Context, id uuid.UUID) (*entity.Project, error) {
	row := r.db.QueryRowContext(ctx, baseSelect+` WHERE id = $1`, id)
	p, err := scanProject(row)
	if err == sql.ErrNoRows {
		return nil, apperrors.NewNotFoundError(fmt.Sprintf("project %s not found", id), err)
	}
	if err != nil {
		return nil, fmt.Errorf("FindByID query: %w", err)
	}
	return p, nil
}

// FindFeatured returns projects marked as featured.
func (r *postgresProjectRepo) FindFeatured(ctx context.Context) ([]*entity.Project, error) {
	rows, err := r.db.QueryContext(ctx, baseSelect+` WHERE featured = true ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("FindFeatured query: %w", err)
	}
	defer rows.Close()
	return scanProjects(rows)
}

// Save inserts a new project, ignoring conflicts on id.
func (r *postgresProjectRepo) Save(ctx context.Context, p *entity.Project) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO projects (id, name, description, tech, url, featured, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 ON CONFLICT (id) DO NOTHING`,
		p.ID, p.Name, p.Description, pq.Array(p.Tech), p.URL, p.Featured, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("Save: %w", err)
	}
	return nil
}

// Update persists changes to an existing project.
func (r *postgresProjectRepo) Update(ctx context.Context, p *entity.Project) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE projects SET name=$1, description=$2, tech=$3, url=$4, featured=$5, updated_at=$6
		 WHERE id = $7`,
		p.Name, p.Description, pq.Array(p.Tech), p.URL, p.Featured, p.UpdatedAt, p.ID,
	)
	if err != nil {
		return fmt.Errorf("Update: %w", err)
	}
	return nil
}

// Delete removes a project by ID.
func (r *postgresProjectRepo) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM projects WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("Delete: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return apperrors.NewNotFoundError(fmt.Sprintf("project %s not found", id), nil)
	}
	return nil
}

// scanProject scans a single *sql.Row into a Project.
func scanProject(row *sql.Row) (*entity.Project, error) {
	p := &entity.Project{}
	var techArr pq.StringArray
	err := row.Scan(&p.ID, &p.Name, &p.Description, &techArr, &p.URL, &p.Featured, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	p.Tech = []string(techArr)
	return p, nil
}

// scanProjects scans *sql.Rows into a slice of Projects.
func scanProjects(rows *sql.Rows) ([]*entity.Project, error) {
	var projects []*entity.Project
	for rows.Next() {
		p := &entity.Project{}
		var techArr pq.StringArray
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &techArr, &p.URL, &p.Featured, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		p.Tech = []string(techArr)
		projects = append(projects, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return projects, nil
}

// InitSchema creates the projects table if it does not exist.
func InitSchema(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS projects (
		id          UUID PRIMARY KEY,
		name        TEXT NOT NULL,
		description TEXT NOT NULL,
		tech        TEXT[] NOT NULL DEFAULT '{}',
		url         TEXT NOT NULL DEFAULT '',
		featured    BOOLEAN NOT NULL DEFAULT FALSE,
		created_at  TIMESTAMPTZ NOT NULL,
		updated_at  TIMESTAMPTZ NOT NULL
	)`)
	return err
}
