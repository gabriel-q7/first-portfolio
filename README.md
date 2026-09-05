# First Portfolio

A production-oriented Svelte and Go portfolio for a small Ubuntu VPS. The
runtime contains only two containers:

- Nginx serves the statically compiled Svelte application and terminates TLS.
- Go serves the REST/WebSocket API and stores data in a persistent SQLite volume.

Node.js is build-only. The backend has no published port, and the production
Compose file contains no `build:` directives.

## Local verification

```bash
make backend-test
cd apps/frontend && npm ci && npm run check && npm run build
cd ../..
make build
```

The production stack expects published GHCR images, a `.env`, and real TLS
files. See [Production deployment](docs/production-deployment.md) for initial
VPS provisioning, GitHub secrets, releases, backups, and rollback.

## Production release

Normal branch pushes do not deploy. Merge and verify `main`, then push a stable
semantic version tag:

```bash
git tag v1.0.0
git push origin v1.0.0
```

GitHub Actions builds `ghcr.io/<owner>/backend:v1.0.0` and
`ghcr.io/<owner>/nginx:v1.0.0`, pushes only those exact tags, and deploys them.
No `latest` image is produced.
