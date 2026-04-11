package repository

import (
	"context"

	"github.com/gabriel-q7/portfolio/backend/internal/domain/entity"
	"github.com/google/uuid"
)

// ProjectRepository defines persistence operations for Project entities.
type ProjectRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Project, error)
	FindAll(ctx context.Context) ([]*entity.Project, error)
	FindFeatured(ctx context.Context) ([]*entity.Project, error)
	Save(ctx context.Context, p *entity.Project) error
	Update(ctx context.Context, p *entity.Project) error
	Delete(ctx context.Context, id uuid.UUID) error
}
