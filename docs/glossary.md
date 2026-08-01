# Glossary

**API key** — Secret supplied through `AUTH_API_KEYS` and required by write
endpoints in the `X-API-Key` header.

**Busy timeout** — How long SQLite waits for a locked writer before returning a
busy error. Production default: five seconds.

**CSP** — Content Security Policy emitted by Nginx to constrain browser resource
loading and reduce XSS impact.

**Named volume** — Docker-managed persistent storage. `portfolio_sqlite_data`
holds the SQLite database, WAL, and shared-memory files.

**SPA fallback** — Nginx returns `index.html` for unknown non-API routes so
client-side routing works on refresh.

**WAL** — SQLite write-ahead logging mode, allowing readers to continue while a
writer commits.
