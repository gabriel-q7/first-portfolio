# Business Rule — Terminal Commands

## Purpose

Define which commands the terminal accepts, how they are resolved, and what output they produce.

## Description

The terminal is an xterm.js client that communicates with the backend via WebSocket (`/ws`). When the user submits input and presses Enter, the frontend sends a `command` message. The backend parses, routes, and executes it, streaming output back as `output` messages, and finally a `done` message.

The backend maintains a command registry (`internal/interfaces/terminal/router.go`). Commands are dispatched to registered `HandlerFunc` functions.

## Inputs

- Raw string typed by the user and submitted via Enter.
- Input is trimmed before parsing. The first token is the command name (case-insensitive), the second is the sub-command, and remaining tokens are args.

## WebSocket Protocol

| Message type | Direction | Description |
|---|---|---|
| `command` | client → server | Execute a terminal command |
| `cancel` | client → server | Cancel the running command (Ctrl+C) |
| `output` | server → client | Incremental command output |
| `done` | server → client | Command completed |
| `error` | server → client | Command failed or error output |
| `status` | server → client | Connection/session status |

## Registered Commands

| Command | Sub-command | Args | Description |
|---|---|---|---|
| `help` | — | — | Show all available commands |
| `about` | — | — | Portfolio introduction |
| `contact` | — | — | Contact information |
| `clear` | — | — | Clear the terminal |
| `projects` | `list` | — | List all projects from `ProjectService` |
| `projects` | `show` | `<id>` | Show details for a project by UUID |
| `skills` | `list` | — | Tech stack with proficiency bars |
| `experience` | `list` | — | Work history |
| `db` | `status` | — | Database connection and project count |
| `db` | `list projects` | — | Projects in SQL-style table output |
| `chat` | `ask` | `"<question>"` | AI assistant with streaming response |

## Constraints

- Only the commands listed above are valid. Any other input returns a "command not found" error.
- Command names are case-insensitive; the parser lowercases the name and sub-command.
- Empty input produces no output and does not send a message.
- The `clear` command sends a single `output` message with content `__CLEAR__`; the xterm.js client calls `term.clear()` on receipt.
- One command executes per WebSocket session at a time; sending a new command while one is running cancels the previous.

## Edge Cases

- **Partial match**: `hel` is not resolved; it produces a "command not found" error.
- **Extra whitespace**: `  help  ` resolves to `help` after trimming.
- **Mixed case**: `HELP`, `Help`, `hElP` all resolve to `help`.
- **Invalid UUID** for `projects show`: returns an error output message.
- **Ctrl+C**: sends a `cancel` message; the backend cancels the command context.

## Related Code

- `apps/backend/internal/interfaces/terminal/router.go` — command registry
- `apps/backend/internal/interfaces/terminal/parser.go` — input parser
- `apps/backend/internal/interfaces/terminal/handler.go` — WebSocket session handler
- `apps/backend/internal/interfaces/terminal/commands/` — command handlers
- `apps/backend/internal/domain/protocol/message.go` — WebSocket message types
- `apps/frontend/src/lib/components/XTerminal.svelte` — xterm.js terminal component
- `apps/frontend/src/lib/services/websocket.ts` — WebSocket client with auto-reconnect

## Related Decisions

- [ADR 003 — Command Whitelist](../decisions/003-command-whitelist.md)
