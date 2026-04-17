package handler

import "context"

// ContactHandler prints contact information.
type ContactHandler struct{}

func (h *ContactHandler) Handle(_ context.Context, _ []string, w Writer) error {
	lines := []string{
		"",
		"  ┌─────────────────────────────────────────┐",
		"  │  CONTACT                                │",
		"  └─────────────────────────────────────────┘",
		"",
		"  GitHub  : github.com/gabriel-q7",
		"  Email   : hello@example.com",
		"",
		"  Feel free to reach out for collaborations or just to say hi.",
		"",
	}
	for _, l := range lines {
		if err := w.Output(l); err != nil {
			return err
		}
	}
	return w.Done()
}
