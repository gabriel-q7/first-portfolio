package project

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/gabriel-q7/portfolio/backend/internal/domain/entity"
	"github.com/gabriel-q7/portfolio/backend/pkg/logger"
	"github.com/google/uuid"
)

// mockProjectRepo is an in-memory mock for ProjectRepository.
type mockProjectRepo struct {
	projects      map[uuid.UUID]*entity.Project
	findAllCalled bool
}

func newMockProjectRepo() *mockProjectRepo {
	return &mockProjectRepo{projects: make(map[uuid.UUID]*entity.Project)}
}

func (m *mockProjectRepo) FindByID(_ context.Context, id uuid.UUID) (*entity.Project, error) {
	p, ok := m.projects[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return p, nil
}

func (m *mockProjectRepo) FindAll(_ context.Context) ([]*entity.Project, error) {
	m.findAllCalled = true
	var list []*entity.Project
	for _, p := range m.projects {
		list = append(list, p)
	}
	return list, nil
}

func (m *mockProjectRepo) FindFeatured(_ context.Context) ([]*entity.Project, error) {
	var list []*entity.Project
	for _, p := range m.projects {
		if p.Featured {
			list = append(list, p)
		}
	}
	return list, nil
}

func (m *mockProjectRepo) Save(_ context.Context, p *entity.Project) error {
	m.projects[p.ID] = p
	return nil
}

func (m *mockProjectRepo) Update(_ context.Context, p *entity.Project) error {
	m.projects[p.ID] = p
	return nil
}

func (m *mockProjectRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.projects, id)
	return nil
}

// mockCacheRepo is an in-memory mock for CacheRepository.
type mockCacheRepo struct {
	data      map[string][]byte
	getCalled int
	setCalled int
}

func newMockCacheRepo() *mockCacheRepo {
	return &mockCacheRepo{data: make(map[string][]byte)}
}

func (m *mockCacheRepo) Get(_ context.Context, key string) ([]byte, error) {
	m.getCalled++
	v, ok := m.data[key]
	if !ok {
		return nil, nil
	}
	return v, nil
}

func (m *mockCacheRepo) Set(_ context.Context, key string, value []byte, _ time.Duration) error {
	m.setCalled++
	m.data[key] = value
	return nil
}

func (m *mockCacheRepo) Delete(_ context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func (m *mockCacheRepo) Exists(_ context.Context, key string) (bool, error) {
	_, ok := m.data[key]
	return ok, nil
}

func (m *mockCacheRepo) FlushPattern(_ context.Context, _ string) error {
	return nil
}

func newTestService(repo *mockProjectRepo, cache *mockCacheRepo) *ProjectService {
	return New(repo, cache, logger.New("debug"))
}

func TestGetAll_CacheHit(t *testing.T) {
	repo := newMockProjectRepo()
	cache := newMockCacheRepo()
	svc := newTestService(repo, cache)
	ctx := context.Background()

	// Pre-populate cache with a project list.
	projects := []*entity.Project{entity.NewProject("CacheProj", "desc", []string{"Go"}, "https://example.com")}
	data, _ := json.Marshal(projects)
	_ = cache.Set(ctx, cacheKeyAllProjects, data, cacheProjectTTL)

	result, err := svc.GetAll(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 project, got %d", len(result))
	}
	if repo.findAllCalled {
		t.Error("repo.FindAll should not have been called on cache hit")
	}
}

func TestGetAll_CacheMiss(t *testing.T) {
	repo := newMockProjectRepo()
	cache := newMockCacheRepo()
	svc := newTestService(repo, cache)
	ctx := context.Background()

	p := entity.NewProject("DBProj", "desc", []string{"Go"}, "https://example.com")
	_ = repo.Save(ctx, p)

	result, err := svc.GetAll(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 project, got %d", len(result))
	}
	if !repo.findAllCalled {
		t.Error("repo.FindAll should have been called on cache miss")
	}
	if cache.setCalled == 0 {
		t.Error("result should have been stored in cache")
	}
}

func TestGetByID_NotFound(t *testing.T) {
	repo := newMockProjectRepo()
	cache := newMockCacheRepo()
	svc := newTestService(repo, cache)
	ctx := context.Background()

	_, err := svc.GetByID(ctx, uuid.New())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreate_Success(t *testing.T) {
	repo := newMockProjectRepo()
	cache := newMockCacheRepo()
	svc := newTestService(repo, cache)
	ctx := context.Background()

	p, err := svc.Create(ctx, "NewProj", "A new project", []string{"Go", "React"}, "https://example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name != "NewProj" {
		t.Errorf("expected name 'NewProj', got %q", p.Name)
	}
	if _, ok := repo.projects[p.ID]; !ok {
		t.Error("project was not saved to repo")
	}
}
