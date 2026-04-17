package commands

import (
	"context"

	"github.com/gabriel-q7/portfolio/backend/internal/interfaces/terminal"
)

// RegisterBasicCommands registers the help, about, clear, and contact handlers.
func RegisterBasicCommands(r *terminal.CommandRouter) {
	r.Register("help", handleHelp)
	r.Register("about", handleAbout)
	r.Register("clear", handleClear)
	r.Register("contact", handleContact)
}

func handleHelp(ctx context.Context, cmd terminal.ParsedCommand, out terminal.OutputStream) error {
	lines := []string{
		"",
		"  \x1b[32m┌─────────────────────────────────────────────────────┐\x1b[0m",
		"  \x1b[32m│  AVAILABLE COMMANDS                                 │\x1b[0m",
		"  \x1b[32m└─────────────────────────────────────────────────────┘\x1b[0m",
		"",
		"  \x1b[36mGeneral\x1b[0m",
		"    help                 — show this help message",
		"    about                — portfolio introduction",
		"    contact              — contact information",
		"    clear                — clear the terminal",
		"",
		"  \x1b[36mProjects\x1b[0m",
		"    projects list        — list all projects",
		"    projects show <id>   — show project details",
		"",
		"  \x1b[36mPortfolio\x1b[0m",
		"    skills list          — tech stack",
		"    experience list      — work history",
		"",
		"  \x1b[36mDatabase\x1b[0m",
		"    db status            — database connection and stats",
		"    db list projects     — list projects in SQL-style output",
		"",
		"  \x1b[36mAI\x1b[0m",
		"    chat ask \"<question>\" — ask the AI assistant",
		"",
	}
	for _, l := range lines {
		if err := out.Write(l); err != nil {
			return err
		}
	}
	return nil
}

func handleAbout(ctx context.Context, cmd terminal.ParsedCommand, out terminal.OutputStream) error {
	lines := []string{
		"",
		"  \x1b[32m┌─────────────────────────────────────────┐\x1b[0m",
		"  \x1b[32m│  PROFILE                                │\x1b[0m",
		"  \x1b[32m└─────────────────────────────────────────┘\x1b[0m",
		"",
		"  Name    : \x1b[36mGabriel\x1b[0m",
		"  Role    : \x1b[36mSoftware Engineer\x1b[0m",
		"  Stack   : TypeScript · Go · SvelteKit · Docker",
		"  Focus   : Clean code, scalable systems, great UX",
		"",
		"  I build things that are fast, minimal and maintainable.",
		"  Currently crafting this very portfolio. \x1b[33m🚀\x1b[0m",
		"",
	}
	for _, l := range lines {
		if err := out.Write(l); err != nil {
			return err
		}
	}
	return nil
}

func handleClear(ctx context.Context, cmd terminal.ParsedCommand, out terminal.OutputStream) error {
	return out.Write("__CLEAR__")
}

func handleContact(ctx context.Context, cmd terminal.ParsedCommand, out terminal.OutputStream) error {
	lines := []string{
		"",
		"  \x1b[32m┌─────────────────────────────────────────┐\x1b[0m",
		"  \x1b[32m│  CONTACT                                │\x1b[0m",
		"  \x1b[32m└─────────────────────────────────────────┘\x1b[0m",
		"",
		"  GitHub  : \x1b[36mgithub.com/gabriel-q7\x1b[0m",
		"  Email   : \x1b[36mhello@example.com\x1b[0m",
		"",
		"  Feel free to reach out for collaborations or just to say hi.",
		"",
	}
	for _, l := range lines {
		if err := out.Write(l); err != nil {
			return err
		}
	}
	return nil
}
