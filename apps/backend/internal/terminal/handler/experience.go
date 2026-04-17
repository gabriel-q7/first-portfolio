package handler

import "context"

// ExperienceHandler lists work experience.
type ExperienceHandler struct{}

func (h *ExperienceHandler) Handle(_ context.Context, args []string, w Writer) error {
	sub := ""
	if len(args) > 0 {
		sub = args[0]
	}
	if sub != "" && sub != "list" {
		return w.Error("  unknown subcommand. Usage: experience list")
	}

	lines := []string{
		"",
		"  ┌─────────────────────────────────────────┐",
		"  │  EXPERIENCE                             │",
		"  └─────────────────────────────────────────┘",
		"",
		"  Software Engineer                          2023 – present",
		"  ├─ Built full-stack web applications with Go and TypeScript",
		"  ├─ Designed and implemented RESTful and WebSocket APIs",
		"  └─ Containerised workloads with Docker and Docker Compose",
		"",
		"  Junior Developer                           2021 – 2023",
		"  ├─ Contributed to frontend features with React and CSS",
		"  ├─ Wrote automated tests and maintained CI pipelines",
		"  └─ Participated in agile ceremonies and code reviews",
		"",
	}
	for _, l := range lines {
		if err := w.Output(l); err != nil {
			return err
		}
	}
	return w.Done()
}
