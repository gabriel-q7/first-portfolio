package handler

import (
	"context"
	"fmt"
	"strings"

	projectUseCase "github.com/gabriel-q7/portfolio/backend/internal/usecase/project"
)

// DBHandler handles the "db" command family.
type DBHandler struct {
	useCase projectUseCase.ProjectUseCase
}

// NewDBHandler creates a DBHandler backed by the given use case.
func NewDBHandler(uc projectUseCase.ProjectUseCase) *DBHandler {
	return &DBHandler{useCase: uc}
}

func (h *DBHandler) Handle(ctx context.Context, args []string, w Writer) error {
	sub := ""
	if len(args) > 0 {
		sub = args[0]
	}

	switch sub {
	case "status":
		return h.status(ctx, w)
	case "list":
		// "db list projects"
		if len(args) >= 2 && args[1] == "projects" {
			return h.listProjects(ctx, w)
		}
		return w.Error("  usage: db list projects")
	default:
		return w.Error(fmt.Sprintf("  unknown subcommand '%s'. Try: db status | db list projects", sub))
	}
}

func (h *DBHandler) status(ctx context.Context, w Writer) error {
	if err := w.Status("  Checking database connectivity…"); err != nil {
		return err
	}

	_, err := h.useCase.GetAll(ctx)

	if err := w.Output(""); err != nil {
		return err
	}
	if err == nil {
		if err2 := w.Output("  Database : ✓ reachable"); err2 != nil {
			return err2
		}
	} else {
		if err2 := w.Output(fmt.Sprintf("  Database : ✗ error — %v", err)); err2 != nil {
			return err2
		}
	}
	if err := w.Output(""); err != nil {
		return err
	}
	return w.Done()
}

func (h *DBHandler) listProjects(ctx context.Context, w Writer) error {
	if err := w.Status("  Querying projects…"); err != nil {
		return err
	}

	projects, err := h.useCase.GetAll(ctx)
	if err != nil {
		return w.Error(fmt.Sprintf("  error querying projects: %v", err))
	}

	if err := w.Output(""); err != nil {
		return err
	}
	if err := w.Output(fmt.Sprintf("  %d project(s) in database:", len(projects))); err != nil {
		return err
	}
	if err := w.Output(""); err != nil {
		return err
	}

	for _, p := range projects {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		tech := "—"
		if len(p.Tech) > 0 {
			tech = strings.Join(p.Tech, ", ")
		}
		if err := w.Output(fmt.Sprintf("  %-36s  %s", p.ID, p.Name)); err != nil {
			return err
		}
		if err := w.Output(fmt.Sprintf("  %-36s  %s", "", tech)); err != nil {
			return err
		}
		if err := w.Output(""); err != nil {
			return err
		}
	}
	return w.Done()
}
