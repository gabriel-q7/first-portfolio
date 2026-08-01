# ADR 003 — Terminal command whitelist

## Status

Accepted

## Context

The browser presents an xterm-based portfolio terminal connected to the Go
backend over a same-origin WebSocket.

## Decision

The backend parses input and dispatches only commands registered in its
`CommandRouter`. It never invokes a shell. Unknown commands return a controlled
error. Command messages are capped at 4 KiB and each session can cancel its
active command.

## Consequences

The UI remains interactive without exposing arbitrary process execution. Adding
a command requires a reviewed code change and deployment.

## Alternatives considered

Arbitrary server-side shell execution was rejected as unsafe. A client-only
command list was replaced because project output now uses backend use cases and
SQLite data.
