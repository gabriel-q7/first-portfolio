package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gabriel-q7/portfolio/backend/internal/domain/entity"
	apperrors "github.com/gabriel-q7/portfolio/backend/pkg/errors"
	"github.com/gabriel-q7/portfolio/backend/pkg/logger"
	"github.com/google/uuid"
)

// ProjectUseCaseInterface defines the handler's dependency on the use case layer.
type ProjectUseCaseInterface interface {
	GetAll(ctx context.Context) ([]*entity.Project, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Project, error)
	GetFeatured(ctx context.Context) ([]*entity.Project, error)
	Create(ctx context.Context, name, desc string, tech []string, url string) (*entity.Project, error)
	Update(ctx context.Context, id uuid.UUID, name, desc string, tech []string, url string) (*entity.Project, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// ProjectHandler handles HTTP requests for the projects resource.
type ProjectHandler struct {
	useCase ProjectUseCaseInterface
	logger  logger.Logger
}

// NewProjectHandler creates a new ProjectHandler.
func NewProjectHandler(uc ProjectUseCaseInterface, log logger.Logger) *ProjectHandler {
	return &ProjectHandler{useCase: uc, logger: log}
}

type createProjectRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tech        []string `json:"tech"`
	URL         string   `json:"url"`
}

// List handles GET /projects and returns all projects as a JSON array.
func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	projects, err := h.useCase.GetAll(r.Context())
	if err != nil {
		respondError(w, err)
		return
	}
	if projects == nil {
		projects = []*entity.Project{}
	}
	respondJSON(w, http.StatusOK, projects)
}

// GetByID handles GET /projects/{id}.
func (h *ProjectHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	rawID := r.PathValue("id")
	id, err := uuid.Parse(rawID)
	if err != nil {
		respondError(w, apperrors.NewBadRequestError("invalid project id", err))
		return
	}

	p, err := h.useCase.GetByID(r.Context(), id)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusOK, p)
}

// Create handles POST /projects.
func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createProjectRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, apperrors.NewBadRequestError("invalid request body", err))
		return
	}

	p, err := h.useCase.Create(r.Context(), req.Name, req.Description, req.Tech, req.URL)
	if err != nil {
		respondError(w, err)
		return
	}
	respondJSON(w, http.StatusCreated, p)
}

// respondError maps an error to an HTTP JSON response.
func respondError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	if appErr, ok := err.(*apperrors.AppError); ok {
		w.WriteHeader(appErr.Code)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": appErr.Message})
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
}

// decodeJSON decodes a JSON request body into dst.
func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}
