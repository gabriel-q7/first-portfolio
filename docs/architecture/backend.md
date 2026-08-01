# Backend architecture

The Go backend keeps domain, use-case, interface, and infrastructure layers.
`ProjectRepository` remains the persistence boundary; its production
implementation uses `database/sql` with the pure-Go SQLite driver.

Runtime dependencies are SQLite and a bounded in-process cache. Startup is
fail-closed when SQLite is unavailable. HTTP has explicit limits, structured
logging, API-key protection for writes, same-origin WebSocket validation, and
graceful shutdown.

The backend listens only on the private Docker bridge. Nginx is the trusted
proxy and the only source allowed to supply client forwarding data.
