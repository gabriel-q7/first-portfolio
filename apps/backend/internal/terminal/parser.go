package terminal

import (
	"strings"
	"unicode"
)

// ParsedCommand holds the top-level command name and its arguments.
type ParsedCommand struct {
	Name string
	Args []string
}

// Parse tokenizes a raw input string into a ParsedCommand.
// It handles single- and double-quoted argument groups so that
// chat ask "hello world" is passed as a single argument.
func Parse(input string) ParsedCommand {
	tokens := tokenize(input)
	if len(tokens) == 0 {
		return ParsedCommand{}
	}
	return ParsedCommand{
		Name: strings.ToLower(tokens[0]),
		Args: tokens[1:],
	}
}

// tokenize splits a string on whitespace while respecting quoted sections.
func tokenize(s string) []string {
	var tokens []string
	var cur strings.Builder
	inQuote := rune(0)

	for _, ch := range s {
		switch {
		case inQuote != 0:
			if ch == inQuote {
				inQuote = 0
			} else {
				cur.WriteRune(ch)
			}
		case ch == '"' || ch == '\'':
			inQuote = ch
		case unicode.IsSpace(ch):
			if cur.Len() > 0 {
				tokens = append(tokens, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(ch)
		}
	}
	if cur.Len() > 0 {
		tokens = append(tokens, cur.String())
	}
	return tokens
}
