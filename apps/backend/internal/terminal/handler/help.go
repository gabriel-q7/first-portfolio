package handler

import "context"

// HelpHandler prints the list of available commands.
type HelpHandler struct{}

func (h *HelpHandler) Handle(_ context.Context, _ []string, w Writer) error {
	lines := []string{
		"",
		"  Available commands:",
		"",
		"  help                    — show this help message",
		"  about                   — about me",
		"  projects list           — list all projects",
		"  projects show <id>      — show a specific project",
		"  skills list             — list my skills",
		"  experience list         — list my work experience",
		"  contact                 — contact information",
		"  chat ask \"<question>\"   — ask the AI assistant",
		"  db status               — database connectivity status",
		"  db list projects        — list projects via DB layer",
		"  clear                   — clear the terminal screen",
		"",
	}
	for _, l := range lines {
		if err := w.Output(l); err != nil {
			return err
		}
	}
	return w.Done()
}
