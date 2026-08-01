package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/gabriel-q7/portfolio/backend/internal/domain/entity"
	"github.com/gabriel-q7/portfolio/backend/internal/domain/repository"
	apperrors "github.com/gabriel-q7/portfolio/backend/pkg/errors"
	"github.com/gabriel-q7/portfolio/backend/pkg/logger"
	"github.com/google/uuid"
)

type ProjectRepository struct {
	db     *sql.DB
	logger logger.Logger
}

func NewProjectRepository(db *sql.DB, log logger.Logger) repository.ProjectRepository {
	return &ProjectRepository{db: db, logger: log}
}

const baseSelect = `SELECT id, name, description, tech, url, featured, created_at, updated_at FROM projects`

func (r *ProjectRepository) FindAll(ctx context.Context) ([]*entity.Project, error) {
	rows, err := r.db.QueryContext(ctx, baseSelect+` ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("find all projects: %w", err)
	}
	defer rows.Close()
	return scanProjects(rows)
}

func (r *ProjectRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Project, error) {
	project, err := scanProject(r.db.QueryRowContext(ctx, baseSelect+` WHERE id = ?`, id.String()))
	if err == sql.ErrNoRows {
		return nil, apperrors.NewNotFoundError(fmt.Sprintf("project %s not found", id), err)
	}
	if err != nil {
		return nil, fmt.Errorf("find project by id: %w", err)
	}
	return project, nil
}

func (r *ProjectRepository) FindFeatured(ctx context.Context) ([]*entity.Project, error) {
	rows, err := r.db.QueryContext(ctx, baseSelect+` WHERE featured = 1 ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("find featured projects: %w", err)
	}
	defer rows.Close()
	return scanProjects(rows)
}

func (r *ProjectRepository) Save(ctx context.Context, project *entity.Project) error {
	tech, err := json.Marshal(project.Tech)
	if err != nil {
		return fmt.Errorf("encode project tech: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO projects (id, name, description, tech, url, featured, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO NOTHING`,
		project.ID.String(), project.Name, project.Description, string(tech), project.URL,
		project.Featured, project.CreatedAt, project.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("save project: %w", err)
	}
	return nil
}

func (r *ProjectRepository) Update(ctx context.Context, project *entity.Project) error {
	tech, err := json.Marshal(project.Tech)
	if err != nil {
		return fmt.Errorf("encode project tech: %w", err)
	}
	result, err := r.db.ExecContext(ctx, `
		UPDATE projects
		SET name = ?, description = ?, tech = ?, url = ?, featured = ?, updated_at = ?
		WHERE id = ?`,
		project.Name, project.Description, string(tech), project.URL, project.Featured,
		project.UpdatedAt, project.ID.String(),
	)
	if err != nil {
		return fmt.Errorf("update project: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read updated row count: %w", err)
	}
	if affected == 0 {
		return apperrors.NewNotFoundError(fmt.Sprintf("project %s not found", project.ID), nil)
	}
	return nil
}

func (r *ProjectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, id.String())
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read deleted row count: %w", err)
	}
	if affected == 0 {
		return apperrors.NewNotFoundError(fmt.Sprintf("project %s not found", id), nil)
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanProject(row scanner) (*entity.Project, error) {
	project := &entity.Project{}
	var rawID string
	var tech string
	if err := row.Scan(
		&rawID, &project.Name, &project.Description, &tech, &project.URL,
		&project.Featured, &project.CreatedAt, &project.UpdatedAt,
	); err != nil {
		return nil, err
	}
	id, err := uuid.Parse(rawID)
	if err != nil {
		return nil, fmt.Errorf("parse stored project id: %w", err)
	}
	project.ID = id
	if err := json.Unmarshal([]byte(tech), &project.Tech); err != nil {
		return nil, fmt.Errorf("decode stored project tech: %w", err)
	}
	return project, nil
}

func scanProjects(rows *sql.Rows) ([]*entity.Project, error) {
	projects := make([]*entity.Project, 0)
	for rows.Next() {
		project, err := scanProject(rows)
		if err != nil {
			return nil, fmt.Errorf("scan project row: %w", err)
		}
		projects = append(projects, project)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate project rows: %w", err)
	}
	return projects, nil
}
