# Glossary

Terms used across the codebase and documentation.

---

**AppError**
A typed error struct (`pkg/errors/errors.go`) that carries an HTTP status code and a user-safe message. Handlers map `AppError` to the appropriate HTTP response; unknown errors become `500`.

**Boot sequence**
The animated startup text that runs in the terminal when the page first loads. Implemented in `src/lib/utils/boot.ts`. The `isBooting` store flag is `true` during this phase; user input is disabled.

**Cache-aside**
The caching strategy used by the backend. On a read, the use case checks the cache first; on a miss, it reads from the database and populates the cache. On a write, the relevant cache keys are invalidated.

**Command registry**
The static map of supported terminal commands defined in `src/lib/commands/index.ts`. Each entry has a `name`, `description`, and `handler` function.

**Command whitelist**
The finite set of commands accepted by the terminal: `help`, `about`, `projects`, `contact`, `clear`. Any other input produces an error.

**CRT effect**
The visual style applied to the terminal: green phosphor color scheme, scanlines overlay, subtle glow. Implemented entirely in CSS.

**Featured project**
A `Project` entity with `Featured: true`. Intended to surface highlighted work via a dedicated API filter (not yet exposed as a separate endpoint).

**In-memory fallback**
Both the project repository and cache have in-memory implementations used automatically when `POSTGRES_DSN` or `REDIS_ADDR` are unset. Data is lost on process restart.

**`isBooting` / `isReady`**
Svelte writable stores (`src/lib/stores/terminal.ts`) that control terminal state. `isBooting` is `true` during the startup animation; `isReady` becomes `true` when the terminal is ready for user input.

**Project**
The core domain entity: a portfolio item with `id`, `name`, `description`, `tech[]`, `url`, `featured`, `created_at`, `updated_at`.

**`respondError`**
The backend helper (`internal/interfaces/http/handler/project.go`) that maps any error to a JSON HTTP response.

**Sentinel value (`__CLEAR__`)**
A special `content` string returned by the `clear` command handler. The component layer detects this and calls `terminalHistory.clear()` instead of rendering the line.

**Terminal history**
The ordered list of `TerminalLine` objects displayed in the terminal. Held in the `terminalHistory` Svelte store. Session-only; not persisted.

**`TerminalLine`**
A typed object representing one line in the terminal. Fields: `id`, `content`, `type` (`'input'` | `'output'` | `'system'` | `'error'`), `timestamp`.

**Token bucket**
The rate-limiting algorithm used by the backend middleware. Each request consumes one token; tokens refill at `RATE_LIMIT_RPS` per second up to a `RATE_LIMIT_BURST` maximum.

**`VITE_API_URL`**
Environment variable that sets the backend base URL for the frontend at build time. Defaults to `http://localhost:8080`.
