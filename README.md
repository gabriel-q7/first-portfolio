# First Portfolio

A small production-oriented portfolio for a single VPS. The runtime has exactly
two containers:

- Nginx serves the compiled Svelte application and is the only public entry point.
- Go serves REST and WebSocket endpoints and persists projects in SQLite.

There is no Node production process and no external database or cache service.

## Run

```bash
cp .env.example .env
chmod 600 .env
# Add a real TLS certificate under deploy/tls before production use.
docker compose up -d --build
```

Open `https://localhost` for a local smoke test (the automatic fallback
certificate is self-signed). Only host ports 80 and 443 are published.

## Development checks

```bash
make backend-test
cd apps/frontend && npm run check && npm run build
docker compose config
```

See [the production deployment guide](docs/production-deployment.md) for the
architecture, configuration, backup/rollback procedure, memory budget, SQLite
trade-offs, and production checklist.
