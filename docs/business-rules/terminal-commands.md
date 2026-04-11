# Business Rule — Terminal Commands

## Purpose

Define which commands the frontend terminal accepts, how they are resolved, and what output they produce.

## Description

The frontend maintains a static command registry (`src/lib/commands/index.ts`). When the user submits input, `executeCommand()` normalises it and looks it up in the registry. The terminal does not execute real shell commands; all commands are pre-defined handlers that return arrays of styled output lines.

## Inputs

- Raw string typed by the user and submitted via Enter.
- Input is trimmed and converted to lowercase before lookup.

## Outputs

An array of `Omit<TerminalLine, 'id' | 'timestamp'>` objects, each with:
- `content`: string to render
- `type`: one of `'input'` | `'output'` | `'system'` | `'error'`

## Registered Commands

| Command | Description | Output type |
|---|---|---|
| `help` | Lists all available commands | `system` headers + `output` lines |
| `about` | Displays profile information | `system` headers + `output` lines |
| `projects` | Lists portfolio projects | `system` headers + `output` lines |
| `contact` | Displays contact information | `system` headers + `output` lines |
| `clear` | Clears the terminal history | Single `system` sentinel line |

## Constraints

- Only the commands listed above are valid. Any other input triggers the "command not found" error.
- Command names are case-insensitive; input is lowercased before lookup.
- Empty input (whitespace only) produces no output and does not add a line to history.
- The `clear` command returns a single sentinel line `{content: '__CLEAR__', type: 'system'}` which the consumer detects and uses to call `terminalHistory.clear()`.
- The first line of any successful command's output is always the echoed input line (type `'input'`), so the user sees what they typed.

## Edge Cases

- **Partial match**: `hel` is not resolved; it produces a "command not found" error.
- **Extra whitespace**: `  help  ` resolves to `help` after trimming.
- **Mixed case**: `HELP`, `Help`, `hElP` all resolve to `help`.

## Examples

**Input**: `help`
```
guest@portfolio:~$ help

  Available commands:

  help      — show this help message
  about     — about me
  projects  — list my projects
  contact   — get in touch
  clear     — clear the terminal
```

**Input**: `foo`
```
guest@portfolio:~$ foo
  command not found: foo. Type 'help' for available commands.
```

**Input**: `` (empty) → no output, no history entry.

## Related Code

- `src/lib/commands/index.ts` — command registry and `executeCommand()`
- `src/lib/types/terminal.ts` — `TerminalLine` and `Command` types
- `src/lib/components/TerminalInput.svelte` — captures input and calls `executeCommand()`

## Related Decisions

- [ADR 003 — Command Whitelist](../decisions/003-command-whitelist.md)
