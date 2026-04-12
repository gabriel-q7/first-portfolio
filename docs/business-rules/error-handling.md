# Business Rule — Error Handling

## Purpose

Define how errors are classified, surfaced to callers, and presented to end users across the frontend and backend.

## Description

### Backend

Errors are represented as `AppError` structs (`pkg/errors/errors.go`). Each `AppError` carries an HTTP status code and a human-readable message. The handler layer calls `respondError(w, err)`, which either maps the `AppError` to the correct HTTP status or defaults to `500 Internal Server Error` for untyped errors.

| Error type | HTTP status | Trigger |
|---|---|---|
| `BadRequestError` | 400 | Invalid input (e.g. malformed UUID, unknown JSON field) |
| `NotFoundError` | 404 | Resource not found in repository |
| `UnauthorizedError` | 401 | Missing or invalid API key |
| `InternalError` | 500 | Unexpected infrastructure failure |

The response body is always:

```json
{ "error": "<message>" }
```

Sensitive details (stack traces, internal causes) are logged server-side but never included in the response body.

### Frontend

The API client (`src/lib/services/api.ts`) throws a JavaScript `Error` if the HTTP response status is not `ok`. Components that call API functions are responsible for catching these errors and deciding how to display them (e.g. an `error` type terminal line).

Terminal command errors (unknown command, empty input) are handled synchronously by `executeCommand()` and produce a line of type `'error'` in the history.

## Inputs

- Backend: any `error` returned by a use case or repository.
- Frontend API client: any non-2xx HTTP response.
- Frontend terminal: any unrecognised command string.

## Outputs

- Backend: JSON `{"error": "..."}` with appropriate status code.
- Frontend API: thrown `Error` with a message like `"HTTP 404: Not Found"`.
- Frontend terminal: a `TerminalLine` with `type: 'error'` and a descriptive message.

## Constraints

- Backend error messages must be user-safe; do not leak internal identifiers, query details, or stack traces.
- All errors are logged at `warn` or `error` level on the backend with full context.
- The frontend does not retry failed API requests automatically; retry logic must be added explicitly if needed.

## Edge Cases

- **Backend unavailable (frontend)**: `fetch` rejects with a network error. This is not currently caught at the component level and will surface as an unhandled promise rejection.
- **JSON decode failure (backend)**: `decodeJSON` uses `DisallowUnknownFields`; extra fields produce a `400 Bad Request`.
- **UUID parse failure (backend)**: Invalid `id` path parameter produces a `400 Bad Request` via `uuid.Parse`.

## Related Code

- `pkg/errors/errors.go` — `AppError` type and constructor functions
- `internal/interfaces/http/handler/project.go` — `respondError()`, `decodeJSON()`
- `src/lib/services/api.ts` — frontend HTTP client with status check
- `src/lib/commands/index.ts` — terminal unknown-command error line

## Related Decisions

- [ADR 002 — Backend Architecture](../decisions/002-backend-architecture.md)
