package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gabriel-q7/portfolio/backend/internal/domain/entity"
	apperrors "github.com/gabriel-q7/portfolio/backend/pkg/errors"
	"github.com/gabriel-q7/portfolio/backend/pkg/logger"
	"github.com/google/uuid"
)

// mockProjectUseCase is a test double for ProjectUseCaseInterface.
type mockProjectUseCase struct {
	projects []*entity.Project
	err      error
}

func (m *mockProjectUseCase) GetAll(_ context.Context) ([]*entity.Project, error) {
	return m.projects, m.err
}

func (m *mockProjectUseCase) GetByID(_ context.Context, id uuid.UUID) (*entity.Project, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, p := range m.projects {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, apperrors.NewNotFoundError("project not found", nil)
}

func (m *mockProjectUseCase) GetFeatured(_ context.Context) ([]*entity.Project, error) {
	return m.projects, m.err
}

func (m *mockProjectUseCase) Create(_ context.Context, name, desc string, tech []string, url string) (*entity.Project, error) {
	if m.err != nil {
		return nil, m.err
	}
	return entity.NewProject(name, desc, tech, url), nil
}

func (m *mockProjectUseCase) Update(_ context.Context, id uuid.UUID, name, desc string, tech []string, url string) (*entity.Project, error) {
	if m.err != nil {
		return nil, m.err
	}
	p := entity.NewProject(name, desc, tech, url)
	p.ID = id
	return p, nil
}

func (m *mockProjectUseCase) Delete(_ context.Context, _ uuid.UUID) error {
	return m.err
}

func TestHealthHandler_Health(t *testing.T) {
	h := NewHealthHandler("test", logger.New("debug"))
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status 'ok', got %q", body["status"])
	}
}

func TestProjectHandler_List_Empty(t *testing.T) {
	uc := &mockProjectUseCase{projects: []*entity.Project{}}
	h := NewProjectHandler(uc, logger.New("debug"))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	rec := httptest.NewRecorder()
	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	var body []*entity.Project
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(body) != 0 {
		t.Errorf("expected empty array, got %d items", len(body))
	}
}

func TestProjectHandler_GetByID_NotFound(t *testing.T) {
	uc := &mockProjectUseCase{}
	h := NewProjectHandler(uc, logger.New("debug"))

	// Use an invalid UUID to trigger a 400.
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/projects/{id}", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/not-a-uuid", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}
