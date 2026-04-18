package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/gabriel-q7/portfolio/backend/internal/interfaces/terminal"
	projectUseCase "github.com/gabriel-q7/portfolio/backend/internal/usecase/project"
)

// RegisterDBCommands registers the db status and db list commands.
func RegisterDBCommands(r *terminal.CommandRouter, svc projectUseCase.ProjectUseCase) {
	r.Register("db", func(ctx context.Context, cmd terminal.ParsedCommand, out terminal.OutputStream) error {
		switch cmd.Sub {
		case "status":
			return handleDBStatus(ctx, svc, out)
		case "list":
			if len(cmd.Args) > 0 && strings.ToLower(cmd.Args[0]) == "projects" {
				return handleDBListProjects(ctx, svc, out)
			}
			return out.WriteError("  usage: db list projects")
		default:
			if cmd.Sub == "" {
				return out.WriteError("  usage: db <status|list projects>")
			}
			return out.WriteError(fmt.Sprintf("  unknown db sub-command: %s. Use 'db status' or 'db list projects'.", cmd.Sub))
		}
	})
}

func handleDBStatus(ctx context.Context, svc projectUseCase.ProjectUseCase, out terminal.OutputStream) error {
	projects, err := svc.GetAll(ctx)
	status := "\x1b[32mONLINE\x1b[0m"
	projectCount := 0
	dbType := "in-memory"
	if err == nil {
		projectCount = len(projects)
	} else {
		status = "\x1b[31mDEGRADED\x1b[0m"
	}

	lines := []string{
		"",
		"  \x1b[32m┌─────────────────────────────────────────┐\x1b[0m",
		"  \x1b[32m│  DATABASE STATUS                        │\x1b[0m",
		"  \x1b[32m└─────────────────────────────────────────┘\x1b[0m",
		"",
		fmt.Sprintf("  Status        : %s", status),
		fmt.Sprintf("  Backend       : %s", dbType),
		fmt.Sprintf("  Projects      : %d record(s)", projectCount),
		"",
	}
	for _, l := range lines {
		if err := out.Write(l); err != nil {
			return err
		}
	}
	return nil
}

func handleDBListProjects(ctx context.Context, svc projectUseCase.ProjectUseCase, out terminal.OutputStream) error {
	projects, err := svc.GetAll(ctx)
	if err != nil {
		return out.WriteError(fmt.Sprintf("  error: %s", err.Error()))
	}

	header := []string{
		"",
		"  SELECT id, name, featured, created_at FROM projects;",
		"",
		"  \x1b[32m+--------------------------------------+------------------------------+----------+------------+\x1b[0m",
		"  \x1b[32m| id                                   | name                         | featured | created_at |\x1b[0m",
		"  \x1b[32m+--------------------------------------+------------------------------+----------+------------+\x1b[0m",
	}
	for _, l := range header {
		if err := out.Write(l); err != nil {
			return err
		}
	}

	if len(projects) == 0 {
		if err := out.Write("  \x1b[2m(no rows)\x1b[0m"); err != nil {
			return err
		}
	} else {
		for _, p := range projects {
			featStr := "false"
			if p.Featured {
				featStr = "true "
			}
			row := fmt.Sprintf("  | %-36s | %-28s | %-8s | %s |",
				p.ID.String(),
				truncate(p.Name, 28),
				featStr,
				p.CreatedAt.Format("2006-01-02"),
			)
			if err := out.Write(row); err != nil {
				return err
			}
		}
	}

	footer := []string{
		"  \x1b[32m+--------------------------------------+------------------------------+----------+------------+\x1b[0m",
		fmt.Sprintf("  \x1b[2m(%d row(s))\x1b[0m", len(projects)),
		"",
	}
	for _, l := range footer {
		if err := out.Write(l); err != nil {
			return err
		}
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 3 {
		return s[:n]
	}
	return s[:n-3] + "..."
}
