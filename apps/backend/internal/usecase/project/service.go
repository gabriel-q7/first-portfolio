package project

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gabriel-q7/portfolio/backend/internal/domain/entity"
	"github.com/gabriel-q7/portfolio/backend/internal/domain/repository"
	apperrors "github.com/gabriel-q7/portfolio/backend/pkg/errors"
	"github.com/gabriel-q7/portfolio/backend/pkg/logger"
	"github.com/google/uuid"
)

const (
	cacheKeyAllProjects      = "projects:all"
	cacheKeyFeaturedProjects = "projects:featured"
	cacheProjectTTL          = 5 * time.Minute
)

// ProjectUseCase is the interface for project business operations.
type ProjectUseCase interface {
	GetAll(ctx context.Context) ([]*entity.Project, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Project, error)
	GetFeatured(ctx context.Context) ([]*entity.Project, error)
	Create(ctx context.Context, name, desc string, tech []string, url string) (*entity.Project, error)
	Update(ctx context.Context, id uuid.UUID, name, desc string, tech []string, url string) (*entity.Project, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// ProjectService implements ProjectUseCase.
type ProjectService struct {
	repo   repository.ProjectRepository
	cache  repository.CacheRepository
	logger logger.Logger
}

// New creates a new ProjectService.
func New(
	repo repository.ProjectRepository,
	cache repository.CacheRepository,
	log logger.Logger,
) *ProjectService {
	return &ProjectService{repo: repo, cache: cache, logger: log}
}

// GetAll returns all projects, using cache when available.
func (s *ProjectService) GetAll(ctx context.Context) ([]*entity.Project, error) {
	start := time.Now()
	defer func() {
		s.logger.Debug("GetAll completed", "duration", time.Since(start))
	}()

	if data, err := s.cache.Get(ctx, cacheKeyAllProjects); err == nil && data != nil {
		var projects []*entity.Project
		if jsonErr := json.Unmarshal(data, &projects); jsonErr == nil {
			s.logger.Debug("GetAll cache hit")
			return projects, nil
		}
	}

	projects, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, apperrors.NewInternalError("failed to fetch projects", err)
	}

	if data, jsonErr := json.Marshal(projects); jsonErr == nil {
		_ = s.cache.Set(ctx, cacheKeyAllProjects, data, cacheProjectTTL)
	}

	return projects, nil
}

// GetByID returns a single project by ID, using cache when available.
func (s *ProjectService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Project, error) {
	start := time.Now()
	defer func() {
		s.logger.Debug("GetByID completed", "id", id, "duration", time.Since(start))
	}()

	cacheKey := fmt.Sprintf("project:%s", id)
	if data, err := s.cache.Get(ctx, cacheKey); err == nil && data != nil {
		var p entity.Project
		if jsonErr := json.Unmarshal(data, &p); jsonErr == nil {
			return &p, nil
		}
	}

	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if data, jsonErr := json.Marshal(p); jsonErr == nil {
		_ = s.cache.Set(ctx, cacheKey, data, cacheProjectTTL)
	}

	return p, nil
}

// GetFeatured returns featured projects.
func (s *ProjectService) GetFeatured(ctx context.Context) ([]*entity.Project, error) {
	start := time.Now()
	defer func() {
		s.logger.Debug("GetFeatured completed", "duration", time.Since(start))
	}()

	if data, err := s.cache.Get(ctx, cacheKeyFeaturedProjects); err == nil && data != nil {
		var projects []*entity.Project
		if jsonErr := json.Unmarshal(data, &projects); jsonErr == nil {
			return projects, nil
		}
	}

	projects, err := s.repo.FindFeatured(ctx)
	if err != nil {
		return nil, apperrors.NewInternalError("failed to fetch featured projects", err)
	}

	if data, jsonErr := json.Marshal(projects); jsonErr == nil {
		_ = s.cache.Set(ctx, cacheKeyFeaturedProjects, data, cacheProjectTTL)
	}

	return projects, nil
}

// Create validates and persists a new project.
func (s *ProjectService) Create(ctx context.Context, name, desc string, tech []string, url string) (*entity.Project, error) {
	start := time.Now()
	s.logger.Info("Creating project", "name", name)
	defer func() {
		s.logger.Debug("Create completed", "duration", time.Since(start))
	}()

	p := entity.NewProject(name, desc, tech, url)
	if err := p.Validate(); err != nil {
		return nil, apperrors.NewBadRequestError(err.Error(), err)
	}

	if err := s.repo.Save(ctx, p); err != nil {
		return nil, apperrors.NewInternalError("failed to save project", err)
	}

	_ = s.cache.Delete(ctx, cacheKeyAllProjects)
	_ = s.cache.Delete(ctx, cacheKeyFeaturedProjects)
	return p, nil
}

// Update mutates an existing project.
func (s *ProjectService) Update(ctx context.Context, id uuid.UUID, name, desc string, tech []string, url string) (*entity.Project, error) {
	start := time.Now()
	s.logger.Info("Updating project", "id", id)
	defer func() {
		s.logger.Debug("Update completed", "duration", time.Since(start))
	}()

	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	p.Update(name, desc, tech, url)
	if err := p.Validate(); err != nil {
		return nil, apperrors.NewBadRequestError(err.Error(), err)
	}

	if err := s.repo.Update(ctx, p); err != nil {
		return nil, apperrors.NewInternalError("failed to update project", err)
	}

	_ = s.cache.Delete(ctx, cacheKeyAllProjects)
	_ = s.cache.Delete(ctx, cacheKeyFeaturedProjects)
	_ = s.cache.Delete(ctx, fmt.Sprintf("project:%s", id))
	return p, nil
}

// Delete removes a project by ID.
func (s *ProjectService) Delete(ctx context.Context, id uuid.UUID) error {
	start := time.Now()
	s.logger.Info("Deleting project", "id", id)
	defer func() {
		s.logger.Debug("Delete completed", "duration", time.Since(start))
	}()

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	_ = s.cache.Delete(ctx, cacheKeyAllProjects)
	_ = s.cache.Delete(ctx, cacheKeyFeaturedProjects)
	_ = s.cache.Delete(ctx, fmt.Sprintf("project:%s", id))
	return nil
}
