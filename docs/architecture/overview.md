# Architecture Overview

## System Context

The portfolio is a full-stack web application presented as an interactive terminal. Visitors open the frontend in a browser, type commands, and see styled output. The backend serves project data and supports write operations via an authenticated REST API.

```
                        Browser
                           │
                    ┌──────┴──────┐
                    │  Frontend   │
                    │ SvelteKit 5 │
                    │  :3000      │
                    └──────┬──────┘
                           │ HTTP / JSON
                    ┌──────┴──────┐
                    │  Backend    │
                    │   Go API    │
                    │  :8080      │
                    └──┬──────┬───┘
                       │      │
              ┌────────┴─┐  ┌─┴────────┐
              │ Postgres │  │  Redis   │
              │  :5432   │  │  :6379   │
              └──────────┘  └──────────┘
```

## Deployment Topology

All services run as Docker containers orchestrated by `docker-compose.yml` at the repo root. Each service has its own `Dockerfile`. The frontend runs on port 3000, the backend on port 8080.

## Communication

- **Frontend → Backend**: HTTP/JSON over `VITE_API_URL`. No WebSockets or streaming in the current version.
- **Backend → Postgres**: standard `database/sql` with `lib/pq` driver. Connection retried up to 5 times on startup.
- **Backend → Redis**: `go-redis` client. Falls back to in-memory cache if `REDIS_ADDR` is unset.

## Fallback Behaviour

Both persistence layers are optional. If `POSTGRES_DSN` is unset, the backend uses an in-memory map as a project repository. If `REDIS_ADDR` is unset, the backend uses an in-memory TTL cache. This lets the backend start with zero external dependencies.

## Cross-Cutting Concerns

| Concern | Mechanism |
|---|---|
| Logging | Structured JSON via `pkg/logger` (wraps `log/slog`) |
| Metrics | Prometheus counters/histograms exposed at `GET /metrics` |
| Auth | API key header (`X-API-Key`) for write routes |
| Rate limiting | Token bucket per IP via `middleware/ratelimit.go` |
| CORS | Configurable origins in `middleware/cors.go` |
| Request tracing | UUID injected into each request as `X-Request-ID` |
| Security headers | `Strict-Transport-Security`, `X-Content-Type-Options`, etc. |
| Timeouts | 30-second per-request timeout; configurable server timeouts |
| Max body size | 1 MB request body limit |

## See Also

- [Frontend Architecture](frontend.md)
- [Backend Architecture](backend.md)
- [API Endpoints](../api/endpoints.md)
- [ADR 001 — Monorepo Documentation Structure](../decisions/001-monorepo-documentation-structure.md)
- [ADR 002 — Backend Architecture](../decisions/002-backend-architecture.md)
