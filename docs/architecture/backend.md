# Backend Architecture

## Purpose

Exposes a JSON REST API for portfolio data. Designed with clean architecture layers (domain, use case, infrastructure, interfaces) and zero mandatory external dependencies at startup.

## Layers

```
cmd/api/main.go          ← Composition root (wires all dependencies)
        │
internal/interfaces/http ← HTTP handlers + router (input boundary)
        │
internal/usecase         ← Business logic, orchestration (application layer)
        │
internal/domain          ← Entities, repository interfaces (core domain)
        │
internal/infrastructure  ← Concrete implementations (Postgres, Redis, AI, queue)
        │
pkg/                     ← Shared utilities (logger, metrics, errors)
```

## Dependency Flow

Dependencies point inward. The domain layer has no external imports. Infrastructure implements domain interfaces; use cases depend only on domain interfaces, not on concrete infrastructure.

```
infrastructure ──► domain ◄── usecase ◄── interfaces/http
```

## Request Lifecycle

```
HTTP request
    │
middleware chain (RequestID → Timeout → MaxBodySize → RequestLogger → CORS → SecureHeaders → RateLimiter)
    │
mux.ServeMux (Go 1.22 method+path routing)
    │
handler (interfaces/http/handler)
    │
usecase.ProjectService
    │
    ├─ CacheRepository.Get (Redis / in-memory)
    │   hit  → return cached result
    │   miss ↓
    └─ ProjectRepository (Postgres / in-memory)
        │
        CacheRepository.Set
        │
    return to handler → JSON response
```

## Middleware Chain (outermost → innermost)

1. `RequestID` — injects a UUID as `X-Request-ID`
2. `Timeout` — 30-second context timeout per request
3. `MaxBodySize` — limits request body to 1 MB
4. `RequestLogger` — structured log + metrics per request
5. `CORS` — applies allowed-origins policy
6. `SecureHeaders` — adds `Strict-Transport-Security`, `X-Frame-Options`, etc.
7. `RateLimiter` — token bucket (100 RPS / 200 burst by default)

## Authentication

Write routes (`POST /api/v1/projects`) require an `X-API-Key` header matching one of the values in `AUTH_API_KEYS`. Read routes are unauthenticated.

## Persistence

### PostgreSQL

- Used when `POSTGRES_DSN` is set.
- Schema initialised automatically on startup (`sqlRepo.InitSchema`).
- Connection retried up to 5 times with a 2-second delay.
- Falls back to in-memory map if unavailable or unconfigured.

### Redis

- Used when `REDIS_ADDR` is set.
- Cache-aside pattern: read from cache first; on miss, read from DB and populate cache.
- Cache invalidated (pattern flush) on project mutations.
- Falls back to in-memory TTL cache if unavailable or unconfigured.

## Background Queue

`internal/infrastructure/queue/worker.go` provides an in-process job queue (buffered channel + goroutine pool, default 4 workers). Used for background tasks like AI insight generation. The queue is gracefully drained on shutdown.

## Error Model

`pkg/errors/errors.go` defines `AppError` with a typed HTTP status code and message. Handlers call `respondError(w, err)` which maps `AppError` to the correct HTTP status and a `{"error": "..."}` JSON body. Unknown errors become `500 Internal Server Error`.

## Observability

| Signal | Mechanism |
|---|---|
| Logs | `log/slog` JSON to stdout, level controlled by `LOG_LEVEL` |
| Metrics | Prometheus registry exposed at `GET /metrics` |
| Health | `GET /health`, `GET /health/live`, `GET /health/ready` |
| Tracing | Placeholder (`TRACING_ENABLED`); not yet wired |

## See Also

- [API Endpoints](../api/endpoints.md)
- [Business Rules — Error Handling](../business-rules/error-handling.md)
- [ADR 002 — Backend Architecture](../decisions/002-backend-architecture.md)
- [ADR 003 — Command Whitelist](../decisions/003-command-whitelist.md)
