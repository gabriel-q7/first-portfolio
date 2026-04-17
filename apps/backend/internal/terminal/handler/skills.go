package handler

import "context"

// SkillsHandler lists technical skills.
type SkillsHandler struct{}

func (h *SkillsHandler) Handle(_ context.Context, args []string, w Writer) error {
	// Only subcommand is "list"
	sub := ""
	if len(args) > 0 {
		sub = args[0]
	}
	if sub != "" && sub != "list" {
		return w.Error("  unknown subcommand. Usage: skills list")
	}

	lines := []string{
		"",
		"  ┌─────────────────────────────────────────┐",
		"  │  SKILLS                                 │",
		"  └─────────────────────────────────────────┘",
		"",
		"  Languages   : Go · TypeScript · JavaScript · Python · SQL",
		"  Frontend    : SvelteKit · React · Tailwind CSS · HTML/CSS",
		"  Backend     : Go · Node.js · REST APIs · WebSockets",
		"  Databases   : PostgreSQL · Redis · SQLite",
		"  DevOps      : Docker · Docker Compose · CI/CD · Linux",
		"  Tools       : Git · VS Code · Neovim · Postman",
		"",
	}
	for _, l := range lines {
		if err := w.Output(l); err != nil {
			return err
		}
	}
	return w.Done()
}
