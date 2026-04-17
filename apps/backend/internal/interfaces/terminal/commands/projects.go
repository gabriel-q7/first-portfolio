package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/gabriel-q7/portfolio/backend/internal/domain/entity"
	"github.com/gabriel-q7/portfolio/backend/internal/interfaces/terminal"
	projectUseCase "github.com/gabriel-q7/portfolio/backend/internal/usecase/project"
	"github.com/google/uuid"
)

// RegisterProjectCommands registers the projects list and show handlers.
func RegisterProjectCommands(r *terminal.CommandRouter, svc projectUseCase.ProjectUseCase) {
	r.Register("projects", func(ctx context.Context, cmd terminal.ParsedCommand, out terminal.OutputStream) error {
		switch cmd.Sub {
		case "list", "":
			return handleProjectsList(ctx, svc, out)
		case "show":
			if len(cmd.Args) == 0 {
				return out.WriteError("  usage: projects show <id>")
			}
			return handleProjectsShow(ctx, svc, cmd.Args[0], out)
		default:
			return out.WriteError(fmt.Sprintf("  unknown sub-command: %s. Use 'projects list' or 'projects show <id>'.", cmd.Sub))
		}
	})
}

func handleProjectsList(ctx context.Context, svc projectUseCase.ProjectUseCase, out terminal.OutputStream) error {
	projects, err := svc.GetAll(ctx)
	if err != nil {
		return out.WriteError(fmt.Sprintf("  error fetching projects: %s", err.Error()))
	}

	lines := []string{
		"",
		"  \x1b[32m┌─────────────────────────────────────────┐\x1b[0m",
		"  \x1b[32m│  PROJECTS                               │\x1b[0m",
		"  \x1b[32m└─────────────────────────────────────────┘\x1b[0m",
		"",
	}
	if len(projects) == 0 {
		lines = append(lines, "  No projects found.", "")
	} else {
		for i, p := range projects {
			lines = append(lines,
				fmt.Sprintf("  \x1b[36m[%02d]\x1b[0m %s", i+1, p.Name),
				fmt.Sprintf("       %s", p.Description),
				fmt.Sprintf("       \x1b[33m%s\x1b[0m", strings.Join(p.Tech, " · ")),
				fmt.Sprintf("       ID: \x1b[2m%s\x1b[0m", p.ID),
				"",
			)
		}
		lines = append(lines, fmt.Sprintf("  \x1b[2mTotal: %d project(s)\x1b[0m", len(projects)), "")
	}

	for _, l := range lines {
		if err := out.Write(l); err != nil {
			return err
		}
	}
	return nil
}

func handleProjectsShow(ctx context.Context, svc projectUseCase.ProjectUseCase, rawID string, out terminal.OutputStream) error {
	id, err := uuid.Parse(rawID)
	if err != nil {
		return out.WriteError(fmt.Sprintf("  invalid project ID: %s", rawID))
	}

	p, err := svc.GetByID(ctx, id)
	if err != nil {
		return out.WriteError(fmt.Sprintf("  project not found: %s", rawID))
	}

	lines := formatProject(p)
	for _, l := range lines {
		if err := out.Write(l); err != nil {
			return err
		}
	}
	return nil
}

func formatProject(p *entity.Project) []string {
	featured := "no"
	if p.Featured {
		featured = "\x1b[33myes\x1b[0m"
	}
	urlLine := "  —"
	if p.URL != "" {
		urlLine = fmt.Sprintf("  \x1b[36m%s\x1b[0m", p.URL)
	}

	return []string{
		"",
		"  \x1b[32m┌─────────────────────────────────────────┐\x1b[0m",
		fmt.Sprintf("  \x1b[32m│  %-41s│\x1b[0m", p.Name),
		"  \x1b[32m└─────────────────────────────────────────┘\x1b[0m",
		"",
		fmt.Sprintf("  ID          : \x1b[2m%s\x1b[0m", p.ID),
		fmt.Sprintf("  Description : %s", p.Description),
		fmt.Sprintf("  Tech        : \x1b[33m%s\x1b[0m", strings.Join(p.Tech, " · ")),
		fmt.Sprintf("  URL         : %s", urlLine),
		fmt.Sprintf("  Featured    : %s", featured),
		fmt.Sprintf("  Created     : %s", p.CreatedAt.Format("2006-01-02")),
		"",
	}
}
