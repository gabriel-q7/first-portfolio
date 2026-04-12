# Portfolio Backend

## Purpose

Exposes a REST API that serves portfolio data (projects, health) to the frontend and any other consumers. Designed for production readiness from the start: structured logging, metrics, rate limiting, graceful shutdown, and pluggable infrastructure.

## Responsibilities

- Serve and manage `Project` resources (list, get by ID, create)
- Expose health and readiness endpoints for container orchestration
- Expose a Prometheus-compatible metrics endpoint
- Authenticate write operations via API key
- Apply rate limiting, CORS, request IDs, and secure headers to all requests
- Persist projects in PostgreSQL (falls back to in-memory store if unconfigured)
- Cache responses in Redis (falls back to in-memory cache if unconfigured)
- Process background jobs via an internal worker queue
- Optionally integrate with an AI provider (OpenAI-compatible) for project insights

## Stack

| Technology | Role |
|---|---|
| Go 1.23 | Language |
| `net/http` (stdlib) | HTTP server and routing |
| PostgreSQL 16 | Persistent project storage |
| Redis 7 | Response cache |
| Prometheus | Metrics exposition |
| `log/slog` | Structured logging |
| Docker | Container packaging |

## Project Structure

```
apps/backend/
├── cmd/api/
│   └── main.go                    # Entrypoint: wires dependencies, starts server
├── configs/
│   └── config.go                  # Configuration loaded from environment variables
├── internal/
│   ├── domain/
│   │   ├── entity/
│   │   │   └── project.go         # Project and AIInsight domain types
│   │   └── repository/
│   │       ├── project.go         # ProjectRepository interface
│   │       └── cache.go           # CacheRepository interface
│   ├── infrastructure/
│   │   ├── clients/
│   │   │   ├── ai/client.go       # OpenAI-compatible AI client
│   │   │   └── external_apis/client.go  # Generic external API client
│   │   ├── persistence/
│   │   │   ├── nosql/redis_cache.go     # Redis CacheRepository implementation
│   │   │   └── sql/project_repo.go      # PostgreSQL ProjectRepository implementation
│   │   └── queue/worker.go        # In-process background job queue
│   ├── interfaces/http/
│   │   ├── router.go              # Route registration and middleware chain
│   │   └── handler/
│   │       ├── health.go          # Health/liveness/readiness handlers
│   │       └── project.go         # Project CRUD handlers
│   ├── middleware/
│   │   ├── auth.go                # API key and JWT authentication
│   │   ├── cors.go                # CORS policy
│   │   ├── logger.go              # Request logging
│   │   ├── ratelimit.go           # Token-bucket rate limiter
│   │   ├── requestid.go           # X-Request-ID injection
│   │   ├── secureheaders.go       # Security response headers
│   │   └── timeout.go             # Per-request timeout
│   └── usecase/project/
│       └── service.go             # Project business logic (cache-aside pattern)
└── pkg/
    ├── errors/errors.go           # Typed application errors (AppError)
    ├── logger/logger.go           # Logger interface and slog implementation
    └── metrics/metrics.go         # Prometheus metrics registry
```

## Getting Started

```bash
cd apps/backend
cp .env.example .env   # fill in credentials as needed
make run               # go run ./cmd/api (http://localhost:8080)
```

To run with the full infrastructure stack (Postgres + Redis):

```bash
# From the repo root
make up
```

## Available Scripts

| Command | Description |
|---|---|
| `make build` | Compile binary to `bin/server` |
| `make run` | Run the server directly with `go run` |
| `make test` | Run all tests with race detector and coverage |
| `make lint` | Run `golangci-lint` |
| `make fmt` | Format all Go files with `gofmt` |
| `make vet` | Run `go vet` |
| `make tidy` | Tidy `go.mod` / `go.sum` |
| `make docker-build` | Build Docker image |
| `make docker-run` | Run Docker image locally |

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `APP_ENV` | `development` | Environment name (`development`, `production`) |
| `SERVER_PORT` | `8080` | HTTP listen port |
| `SERVER_READ_TIMEOUT` | `15s` | HTTP read timeout |
| `SERVER_WRITE_TIMEOUT` | `15s` | HTTP write timeout |
| `SERVER_IDLE_TIMEOUT` | `60s` | HTTP idle connection timeout |
| `POSTGRES_DSN` | _(empty)_ | PostgreSQL connection string; omit to use in-memory store |
| `POSTGRES_MAX_OPEN_CONNS` | `25` | Max open DB connections |
| `POSTGRES_MAX_IDLE_CONNS` | `5` | Max idle DB connections |
| `POSTGRES_CONN_MAX_LIFETIME` | `5m` | Max connection lifetime |
| `REDIS_ADDR` | _(empty)_ | Redis address (e.g. `localhost:6379`); omit to use in-memory cache |
| `REDIS_PASSWORD` | _(empty)_ | Redis password |
| `REDIS_DB` | `0` | Redis logical database index |
| `REDIS_POOL_SIZE` | `10` | Redis connection pool size |
| `AI_PROVIDER` | `openai` | AI provider name |
| `AI_API_KEY` | _(empty)_ | AI API key; omit to use no-op AI client |
| `AI_BASE_URL` | `https://api.openai.com/v1` | AI API base URL |
| `AI_RATE_LIMIT` | `10` | AI requests per second |
| `AUTH_API_KEYS` | _(empty)_ | Comma-separated list of valid API keys for write routes |
| `JWT_SECRET` | _(empty)_ | Secret for JWT validation (future use) |
| `RATE_LIMIT_RPS` | `100` | Global rate limit (requests per second) |
| `RATE_LIMIT_BURST` | `200` | Rate limit burst size |
| `LOG_LEVEL` | `info` | Log verbosity (`debug`, `info`, `warn`, `error`) |
| `METRICS_ENABLED` | `true` | Enable Prometheus metrics endpoint |

## Integrations

- **Frontend** — `GET /api/v1/projects` and `GET /health` are consumed by `apps/frontend/src/lib/services/api.ts`
- **PostgreSQL** — stores Projects; the schema is auto-initialised on startup via `sqlRepo.InitSchema`
- **Redis** — caches project list responses; keys are flushed when projects are mutated
- **OpenAI API** — optional; used for AI-generated project insights via the `AIProvider` interface

## Related Documentation

- [Architecture — Backend](../../docs/architecture/backend.md)
- [API Endpoints](../../docs/api/endpoints.md)
- [Business Rules — Error Handling](../../docs/business-rules/error-handling.md)
- [ADR 002 — Backend Architecture](../../docs/decisions/002-backend-architecture.md)
