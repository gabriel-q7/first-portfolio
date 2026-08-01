# API endpoints

Public browser traffic is same-origin through Nginx under `/api`. The backend's
port is not published.

## Health

`GET /api/health` returns general process status. Container-only probes use
`GET /health/live` and `GET /health/ready`; readiness also pings SQLite.

## Projects

### `GET /api/v1/projects`

Returns all projects, or `[]` when the fresh database is empty.

### `GET /api/v1/projects/{id}`

Returns one project by UUID. Invalid UUIDs return 400; missing rows return 404.

### `POST /api/v1/projects`

Creates a project and requires `X-API-Key`.

```json
{
  "name": "project name",
  "description": "short description",
  "tech": ["Go", "SQLite"],
  "url": "https://github.com/example/repo"
}
```

Unknown JSON fields and bodies over 1 MiB are rejected. Valid keys come from the
comma-separated `AUTH_API_KEYS` environment variable.

## WebSocket terminal

`GET /api/ws` upgrades to the terminal protocol. Nginx maps this public route to
the backend's internal `/ws` route. The backend accepts only a matching
same-origin `Origin` host and limits messages to 4 KiB.

Client command:

```json
{
  "type": "command",
  "request_id": "abc123",
  "command": "projects list",
  "timestamp": "2026-01-01T00:00:00Z"
}
```

Server messages use `status`, `output`, `error`, and `done` types.

## Errors and limits

JSON errors use `{ "error": "message" }`. Nginx applies 10 requests/second per
client with a burst of 20; Go adds a bounded 20 requests/second token bucket
with a burst of 40. Limit responses are HTTP 429.
