package terminal

import "strings"

// ParsedCommand holds the result of parsing raw terminal input.
type ParsedCommand struct {
	// Name is the primary command (e.g. "projects", "chat").
	Name string
	// Sub is an optional sub-command (e.g. "list", "show", "ask").
	Sub string
	// Args are positional arguments after the sub-command.
	Args []string
	// Raw is the original trimmed input.
	Raw string
}

// Parse splits raw input into a ParsedCommand.
// Format: <name> [sub] [args...]
func Parse(input string) ParsedCommand {
	trimmed := strings.TrimSpace(input)
	parts := strings.Fields(trimmed)

	if len(parts) == 0 {
		return ParsedCommand{Raw: trimmed}
	}

	cmd := ParsedCommand{
		Name: strings.ToLower(parts[0]),
		Raw:  trimmed,
	}

	if len(parts) > 1 {
		cmd.Sub = strings.ToLower(parts[1])
	}

	if len(parts) > 2 {
		cmd.Args = parts[2:]
	}

	return cmd
}
