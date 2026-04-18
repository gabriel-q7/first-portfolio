package commands

import (
	"context"
	"fmt"

	"github.com/gabriel-q7/portfolio/backend/internal/interfaces/terminal"
)

// RegisterPortfolioCommands registers the skills and experience handlers.
func RegisterPortfolioCommands(r *terminal.CommandRouter) {
	r.Register("skills", func(ctx context.Context, cmd terminal.ParsedCommand, out terminal.OutputStream) error {
		switch cmd.Sub {
		case "list", "":
			return handleSkillsList(ctx, out)
		default:
			return out.WriteError(fmt.Sprintf("  unknown sub-command: %s. Use 'skills list'.", cmd.Sub))
		}
	})
	r.Register("experience", func(ctx context.Context, cmd terminal.ParsedCommand, out terminal.OutputStream) error {
		switch cmd.Sub {
		case "list", "":
			return handleExperienceList(ctx, out)
		default:
			return out.WriteError(fmt.Sprintf("  unknown sub-command: %s. Use 'experience list'.", cmd.Sub))
		}
	})
}

func handleSkillsList(_ context.Context, out terminal.OutputStream) error {
	lines := []string{
		"",
		"  \x1b[32m┌─────────────────────────────────────────┐\x1b[0m",
		"  \x1b[32m│  TECH STACK                             │\x1b[0m",
		"  \x1b[32m└─────────────────────────────────────────┘\x1b[0m",
		"",
		"  \x1b[36mLanguages\x1b[0m",
		"    TypeScript   ████████████  Expert",
		"    Go           ██████████    Advanced",
		"    Python       ████████      Intermediate",
		"",
		"  \x1b[36mFrontend\x1b[0m",
		"    SvelteKit    ████████████  Expert",
		"    React        ████████      Intermediate",
		"    Tailwind CSS ████████████  Expert",
		"",
		"  \x1b[36mBackend\x1b[0m",
		"    Go net/http  ████████████  Expert",
		"    PostgreSQL   ██████████    Advanced",
		"    Redis        ████████      Advanced",
		"",
		"  \x1b[36mInfrastructure\x1b[0m",
		"    Docker       ████████████  Expert",
		"    Docker Compose ██████████  Advanced",
		"    Linux/Bash   ██████████    Advanced",
		"",
	}
	for _, l := range lines {
		if err := out.Write(l); err != nil {
			return err
		}
	}
	return nil
}

func handleExperienceList(_ context.Context, out terminal.OutputStream) error {
	lines := []string{
		"",
		"  \x1b[32m┌─────────────────────────────────────────┐\x1b[0m",
		"  \x1b[32m│  EXPERIENCE                             │\x1b[0m",
		"  \x1b[32m└─────────────────────────────────────────┘\x1b[0m",
		"",
		"  \x1b[36mSoftware Engineer\x1b[0m",
		"  Various Projects · 2020 — Present",
		"  · Building full-stack web applications",
		"  · Designing clean, maintainable architectures",
		"  · Working with Go and TypeScript ecosystems",
		"",
		"  \x1b[36mPersonal Projects\x1b[0m",
		"  Open Source · Ongoing",
		"  · first-portfolio — Terminal-style portfolio (this site)",
		"  · Exploring AI integration and streaming APIs",
		"",
	}
	for _, l := range lines {
		if err := out.Write(l); err != nil {
			return err
		}
	}
	return nil
}
