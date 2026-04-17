package handler

import "context"

// AboutHandler prints profile information.
type AboutHandler struct{}

func (h *AboutHandler) Handle(_ context.Context, _ []string, w Writer) error {
	lines := []string{
		"",
		"  ┌─────────────────────────────────────────┐",
		"  │  PROFILE                                │",
		"  └─────────────────────────────────────────┘",
		"",
		"  Name    : Gabriel",
		"  Role    : Software Engineer",
		"  Stack   : TypeScript · Go · SvelteKit · Docker",
		"  Focus   : Clean code, scalable systems, great UX",
		"",
		"  I build things that are fast, minimal and maintainable.",
		"  Currently crafting this very portfolio. 🚀",
		"",
	}
	for _, l := range lines {
		if err := w.Output(l); err != nil {
			return err
		}
	}
	return w.Done()
}
