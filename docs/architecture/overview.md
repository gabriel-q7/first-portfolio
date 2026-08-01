# Architecture overview

The production runtime is a two-container modular monolith:

```text
Client → Nginx:80/443 → Go:8080 → SQLite:/data/portfolio.db
          └── compiled Svelte files
```

Nginx is the only public entry point. It serves the static SPA and proxies
same-origin `/api` requests. Go owns business logic and persistence. SQLite is a
local file in the `portfolio_sqlite_data` Docker volume.

See [Single-VPS production deployment](../production-deployment.md) for the
detailed diagram, boundaries, resource budget, and decisions.
