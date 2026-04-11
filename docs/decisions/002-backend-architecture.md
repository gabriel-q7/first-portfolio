# ADR 002 — Backend Architecture

## Status

Accepted

## Context

The portfolio required a backend to serve project data, expose health endpoints, and eventually support AI-generated insights. The architecture needed to be production-ready (logging, metrics, graceful shutdown) while remaining simple enough for a single-developer project and runnable without any external services.

## Decision

Use Go with the standard `net/http` package (Go 1.22+ method+path routing) and a clean architecture layering:

- **Domain** — entities and repository interfaces; no external imports.
- **Use case** — business logic; depends only on domain interfaces.
- **Infrastructure** — concrete implementations (Postgres, Redis, AI client, queue).
- **Interfaces/HTTP** — handlers and router; depends on use case interfaces.

Both Postgres and Redis are optional. If their environment variables are unset, the server starts with in-memory fallbacks. This makes local development and testing require zero infrastructure.

The HTTP middleware chain is assembled in `internal/interfaces/http/router.go` and applies: request ID, timeout, body size limit, request logging, CORS, security headers, and rate limiting.

## Consequences

- The server starts with no external dependencies; useful for local development and CI.
- Adding a new persistence backend requires only implementing the `ProjectRepository` or `CacheRepository` interface.
- The clean architecture layering increases file count but makes each layer independently testable.
- `net/http` without a third-party router keeps the dependency surface small but requires careful route ordering when patterns overlap.

## Alternatives Considered

- **Framework (Gin, Echo, Chi)**: Rejected to keep the dependency surface minimal; Go 1.22 native routing covers the project's needs.
- **Mandatory Postgres on startup**: Rejected because it complicates local development and CI without meaningful benefit at this scale.
- **Monolithic `main.go`**: Rejected because it makes unit testing individual layers impossible and conflates concerns.
