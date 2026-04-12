# ADR 003 — Command Whitelist (Frontend Terminal)

## Status

Accepted

## Context

The portfolio frontend presents a fake terminal. An early question was whether to allow arbitrary shell-like commands (piped to a real shell, processed server-side, or simulated client-side with a large library) or to restrict input to a fixed set of portfolio-specific commands.

## Decision

Maintain a static whitelist of commands defined entirely in `src/lib/commands/index.ts`. Any input not in the whitelist produces a "command not found" error. No server round-trip occurs for command execution; all output is generated synchronously on the client.

The current whitelist: `help`, `about`, `projects`, `contact`, `clear`.

## Consequences

- Command execution is instant (no network latency).
- No risk of unintended shell execution or injection.
- Adding a new command requires a code change and deployment; commands cannot be configured at runtime.
- The terminal feels intentionally constrained — which fits the portfolio aesthetic of a curated, minimal presentation.

## Alternatives Considered

- **Server-side command execution**: Rejected due to security risk, complexity, and latency.
- **Large terminal emulator library (xterm.js, etc.)**: Rejected as overkill for a portfolio whose terminal is a UI metaphor, not a real shell.
- **Dynamic command registry via API**: Rejected as unnecessary complexity; the command set is stable and content-driven.
