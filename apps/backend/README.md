# Backend

Go `net/http` API with SQLite persistence and a WebSocket terminal.

```bash
SQLITE_PATH=/tmp/portfolio.db go run ./cmd/api
go test -race ./...
go vet ./...
```

Production configuration is environment-only:

| Variable | Default |
|---|---|
| `SERVER_PORT` | `8080` |
| `SQLITE_PATH` | `/data/portfolio.db` |
| `SQLITE_BUSY_TIMEOUT` | `5s` |
| `SQLITE_MAX_OPEN_CONNS` | `2` |
| `TRUSTED_PROXY_CIDR` | empty |
| `AUTH_API_KEYS` | empty (writes denied) |
| `AI_API_KEY` | empty (chat disabled) |
| `LOG_LEVEL` | `info` |

The container is statically compiled, distroless, non-root, read-only, and
health-checked by the application binary itself.
