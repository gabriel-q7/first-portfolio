package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/gabriel-q7/portfolio/backend/internal/domain/entity"
	projectUseCase "github.com/gabriel-q7/portfolio/backend/internal/usecase/project"
	"github.com/google/uuid"
)

// ProjectsHandler handles the "projects" command family.
type ProjectsHandler struct {
	useCase projectUseCase.ProjectUseCase
}

// NewProjectsHandler creates a ProjectsHandler backed by the given use case.
func NewProjectsHandler(uc projectUseCase.ProjectUseCase) *ProjectsHandler {
	return &ProjectsHandler{useCase: uc}
}

func (h *ProjectsHandler) Handle(ctx context.Context, args []string, w Writer) error {
	sub := ""
	if len(args) > 0 {
		sub = args[0]
	}

	switch sub {
	case "", "list":
		return h.list(ctx, w)
	case "show":
		if len(args) < 2 {
			return w.Error("  usage: projects show <id>")
		}
		return h.show(ctx, args[1], w)
	default:
		return w.Error(fmt.Sprintf("  unknown subcommand '%s'. Try: projects list | projects show <id>", sub))
	}
}

func (h *ProjectsHandler) list(ctx context.Context, w Writer) error {
	projects, err := h.useCase.GetAll(ctx)
	if err != nil {
		return w.Error(fmt.Sprintf("  error fetching projects: %v", err))
	}

	if err := w.Output(""); err != nil {
		return err
	}
	if err := w.Output("  ┌─────────────────────────────────────────┐"); err != nil {
		return err
	}
	if err := w.Output("  │  PROJECTS                               │"); err != nil {
		return err
	}
	if err := w.Output("  └─────────────────────────────────────────┘"); err != nil {
		return err
	}
	if err := w.Output(""); err != nil {
		return err
	}

	if len(projects) == 0 {
		if err := w.Output("  No projects found."); err != nil {
			return err
		}
		if err := w.Output(""); err != nil {
			return err
		}
		return w.Done()
	}

	for i, p := range projects {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		if err := writeProjectSummary(w, i+1, p); err != nil {
			return err
		}
	}
	if err := w.Output("  Tip: use 'projects show <id>' for details."); err != nil {
		return err
	}
	if err := w.Output(""); err != nil {
		return err
	}
	return w.Done()
}

func (h *ProjectsHandler) show(ctx context.Context, rawID string, w Writer) error {
	id, err := uuid.Parse(rawID)
	if err != nil {
		return w.Error(fmt.Sprintf("  invalid project id: %s", rawID))
	}

	p, err := h.useCase.GetByID(ctx, id)
	if err != nil {
		return w.Error(fmt.Sprintf("  project not found: %s", rawID))
	}

	lines := []string{
		"",
		fmt.Sprintf("  Name        : %s", p.Name),
		fmt.Sprintf("  Description : %s", p.Description),
		fmt.Sprintf("  Tech        : %s", strings.Join(p.Tech, " · ")),
	}
	if p.URL != "" {
		lines = append(lines, fmt.Sprintf("  URL         : %s", p.URL))
	}
	if p.Featured {
		lines = append(lines, "  Featured    : ✓")
	}
	lines = append(lines, "")

	for _, l := range lines {
		if err := w.Output(l); err != nil {
			return err
		}
	}
	return w.Done()
}

func writeProjectSummary(w Writer, n int, p *entity.Project) error {
	if err := w.Output(fmt.Sprintf("  [%02d] %s", n, p.Name)); err != nil {
		return err
	}
	if p.Description != "" {
		if err := w.Output(fmt.Sprintf("       %s", p.Description)); err != nil {
			return err
		}
	}
	if len(p.Tech) > 0 {
		if err := w.Output(fmt.Sprintf("       %s", strings.Join(p.Tech, " · "))); err != nil {
			return err
		}
	}
	return w.Output("")
}
