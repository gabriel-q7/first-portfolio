# API Endpoints

Base URL is configured via `VITE_API_URL` on the frontend and defaults to `http://localhost:8080`.

---

## Health

### `GET /health`

General health check. Always returns `200 OK` if the process is running.

**Response**

```json
{
  "status": "ok",
  "version": "1.0.0"
}
```

---

### `GET /health/live`

Kubernetes-style liveness probe. Returns `200 OK` if the process is alive.

---

### `GET /health/ready`

Kubernetes-style readiness probe. Returns `200 OK` if the service is ready to accept traffic (dependencies checked).

---

## Metrics

### `GET /metrics`

Prometheus text-format metrics. No authentication required.

---

## Projects

### `GET /api/v1/projects`

List all projects. No authentication required.

**Response** `200 OK`

```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "first-portfolio",
    "description": "Hacker-style portfolio in a monorepo.",
    "tech": ["SvelteKit", "Go", "Docker"],
    "url": "https://github.com/gabriel-q7/first-portfolio",
    "featured": true,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

Returns an empty array `[]` if no projects exist.

---

### `GET /api/v1/projects/{id}`

Get a single project by UUID. No authentication required.

**Path parameter**: `id` — UUID v4

**Response** `200 OK` — project object (same shape as a single item in the list response)

**Error responses**

| Status | Condition |
|---|---|
| `400 Bad Request` | `id` is not a valid UUID |
| `404 Not Found` | No project with that ID |

---

### `POST /api/v1/projects`

Create a new project. **Requires authentication** via `X-API-Key` header.

**Request body**

```json
{
  "name": "project name",
  "description": "short description",
  "tech": ["Go", "Postgres"],
  "url": "https://github.com/example/repo"
}
```

**Response** `201 Created` — created project object

**Error responses**

| Status | Condition |
|---|---|
| `400 Bad Request` | Missing required field, unknown field, or invalid JSON |
| `401 Unauthorized` | Missing or invalid `X-API-Key` |

---

## Authentication

Write routes require the `X-API-Key` header:

```
X-API-Key: <key>
```

Valid keys are set via the `AUTH_API_KEYS` environment variable (comma-separated list).

---

## Error Format

All error responses share the same JSON shape:

```json
{ "error": "<human-readable message>" }
```

---

## Rate Limiting

All routes are subject to a token-bucket rate limiter. Defaults: 100 requests/second, burst of 200. On limit exceeded the response is `429 Too Many Requests`.
