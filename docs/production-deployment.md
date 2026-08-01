# Single-VPS production deployment

This deployment is sized for one vCPU and 1 GB RAM. It deliberately uses one
application process and one edge process; everything else is a file or a Docker
primitive.

## 1. Architecture

```text
Internet
   │ TCP 80/443 only
   ▼
┌──────────────────────────────────────────────────────┐
│ Nginx container (non-root, 64 MiB limit)             │
│  ├─ /            compiled Svelte static files       │
│  ├─ /api/*       reverse proxy to backend:8080      │
│  ├─ /api/ws      WebSocket proxy                    │
│  └─ TLS, gzip, cache, headers, limits, access logs  │
└──────────────────────┬───────────────────────────────┘
                       │ private Docker bridge
                       ▼
┌──────────────────────────────────────────────────────┐
│ Go backend container (non-root, 256 MiB limit)       │
│  ├─ REST API and terminal WebSocket                  │
│  ├─ bounded in-process response cache                │
│  └─ database/sql → SQLite WAL                        │
└──────────────────────┬───────────────────────────────┘
                       │ /data (read/write)
                       ▼
                 Docker named volume
                 portfolio_sqlite_data
```

The backend has no host port. Nginx overwrites forwarded headers and the backend
trusts only Nginx's fixed `172.28.0.10/32` address. The Docker bridge is named
`app_internal`; it remains egress-capable so the optional AI client can make
outbound HTTPS calls. If AI is permanently disabled, setting the network to
`internal: true` provides stricter egress isolation.

## 2. Relevant folder structure

```text
.
├── .dockerignore
├── .env.example
├── docker-compose.yml
├── apps
│   ├── backend
│   │   ├── Dockerfile
│   │   ├── cmd/api/main.go
│   │   ├── configs/config.go
│   │   └── internal/infrastructure
│   │       ├── cache/memory.go
│   │       └── persistence/sqlite
│   │           ├── database.go
│   │           ├── project_repo.go
│   │           └── project_repo_test.go
│   └── frontend
│       ├── Dockerfile
│       ├── svelte.config.js
│       ├── src/
│       └── static/
└── deploy
    ├── nginx
    │   ├── nginx.conf
    │   ├── default.conf
    │   ├── security-headers.conf
    │   └── start-nginx.sh
    └── tls/
```

## 3. Architectural decisions and optimizations

### SQLite persistence

The domain `ProjectRepository` interface is unchanged. Only its infrastructure
implementation changed. Projects use UUID text keys, technology lists are JSON,
and every query is parameterized.

Startup is fail-closed: if the database directory, file, pragmas, or schema
cannot be initialized, the backend exits instead of silently losing writes in
memory. The initial schema is created in a transaction and marked with
`PRAGMA user_version = 1`.

The connection applies:

| Setting | Value | Reason |
|---|---:|---|
| `journal_mode` | `WAL` | Readers do not block the writer; online backup is safer. |
| `busy_timeout` | 5000 ms | Brief write contention waits instead of failing immediately. |
| `synchronous` | `NORMAL` | Good WAL durability with fewer fsyncs than `FULL`. |
| `foreign_keys` | `ON` | Enforces future relational constraints. |
| `cache_size` | `-2000` | Caps SQLite's page cache near 2 MiB. |
| `journal_size_limit` | 8 MiB | Prevents an idle WAL from remaining unnecessarily large. |
| `temp_store` | memory | Avoids temp files for this small data set. |
| open connections | 2 | Allows one reader around a writer without a large pool. |

The named volume is initialized from `/data`, which is owned by UID/GID 65532.
The backend then enforces directory mode `0700` and database mode `0600`. WAL and
shared-memory files live beside the database on the same volume.

Single SQL statements are already atomic. Schema changes use explicit
transactions. Any future multi-statement business operation must use
`BeginTx`, pass the transaction through repository methods, keep the transaction
short, commit once, and always defer rollback. Do not perform network calls while
holding a SQLite write transaction.

SQLite limitations:

- There is one writer at a time. WAL improves read concurrency, not write
  concurrency.
- It is appropriate for one application instance and modest write volume, not
  horizontal backend replicas.
- The volume must be local block storage; do not place the database on NFS or a
  shared network filesystem.
- Long transactions and slow disk can still produce busy-timeout failures.
- Backups must include SQLite consistency. Stop the app for a file-level backup,
  or use SQLite's online backup/VACUUM mechanism.
- If sustained concurrent writes become material, move to a client/server
  database in a separately planned migration.

### Go backend

- A multi-stage build compiles with `CGO_ENABLED=0`, `-trimpath`, stripped debug
  symbols, no build ID, and a version injected at link time.
- `modernc.org/sqlite` is pure Go, preserving a static binary and allowing a
  distroless final image.
- The distroless image contains no shell or package manager and runs as UID
  65532.
- The root filesystem is read-only. Only the SQLite volume and an 8 MiB `/tmp`
  tmpfs are writable.
- The HTTP server sets header/read/write/idle timeouts and a 16 KiB maximum
  header size. Requests are capped at 1 MiB in both Nginx and Go.
- `/health/live` checks the process; `/health/ready` pings SQLite. The same
  binary implements the container health check, avoiding curl/wget in the image.
- SIGTERM/SIGINT trigger bounded graceful shutdown followed by a WAL checkpoint.
- Logs are structured JSON to stdout. Docker's local log driver rotates them.
- The response cache is bounded to 256 entries and cleans up opportunistically,
  so there is no cache daemon or cleanup goroutine.
- The visitor limiter is bounded to 2,048 clients and also cleans up
  opportunistically.
- The unused metrics registry, worker pool, external client, and fallback
  persistence implementations were removed.

### Svelte and Nginx

SvelteKit uses `adapter-static`; the frontend build creates only HTML, CSS, JS,
and precompressed assets. Node and npm exist only in the builder stage. Nginx is
the runtime.

Nginx runs as its unprivileged `nginx` user on container ports 8080/8443, while
Docker maps host ports 80/443. It has all Linux capabilities dropped, a read-only
root filesystem, a 16 MiB `/tmp` tmpfs, one worker, 256 worker connections, and
eight keep-alive backend connections.

The edge configuration provides gzip, immutable one-year caching for hashed
assets, no-store HTML, SPA fallback, 30-second keep-alive, a 1 MiB body limit,
proxy timeouts, per-IP request/connection limits, WebSocket timeouts, TLS 1.2/1.3,
and security headers. The CSP permits scripts only from the same origin; inline
styles remain allowed because the compiled Svelte UI uses style attributes.
Remote fonts were removed.

CORS is intentionally disabled. The browser calls same-origin `/api` routes, so
cross-origin access is unnecessary. WebSocket upgrades require an `Origin` host
equal to the request host.

### Resource budget

| Component | Reservation | Hard limit | CPU limit |
|---|---:|---:|---:|
| Go backend | 64 MiB | 256 MiB | 0.75 CPU |
| Nginx | 16 MiB | 64 MiB | 0.25 CPU |
| Combined containers | 80 MiB | 320 MiB | 1 CPU |

`GOMAXPROCS=1` prevents excess scheduler threads. `GOMEMLIMIT=192MiB` makes the
collector react before the container's 256 MiB ceiling. `GOGC=100` is a balanced
starting point; lowering it trades CPU for memory and is usually counterproductive
on one vCPU. The remaining RAM is reserved for the kernel, Docker, filesystem
cache, SSH, and short deployment spikes. A 1–2 GB host swap file is recommended
as an emergency cushion, not as normal application memory.

## 4. Deployment

Prerequisites: a Linux VPS, Docker Engine with the Compose plugin, DNS pointing
to the VPS, and firewall rules allowing SSH plus TCP 80/443 only.

1. Clone the repository and enter it.
2. Create runtime settings:

   ```bash
   cp .env.example .env
   chmod 600 .env
   openssl rand -base64 48
   ```

   Put the generated value in `AUTH_API_KEYS`. Comma-separated keys are
   supported for rotation. `AI_API_KEY` is optional and is never built into an
   image.

3. Place the domain certificate chain in `deploy/tls/fullchain.pem` and the
   private key in `deploy/tls/privkey.pem`, then run:

   ```bash
   sudo chown 101:101 deploy/tls/privkey.pem
   chmod 600 deploy/tls/privkey.pem
   docker compose config
   docker compose up -d --build
   ```

   Without those files, startup creates a seven-day self-signed certificate for
   smoke testing. That fallback is not production TLS.

4. Verify:

   ```bash
   docker compose ps
   curl -fsS https://example.com/api/health
   curl -fsS https://example.com/api/v1/projects
   docker compose logs --tail=100
   docker stats --no-stream
   ```

The application creates a fresh `/data/portfolio.db` automatically. There is no
data migration because the previous database was empty.

## 5. Backup and rollback

Before every application/schema upgrade, make a cold volume backup:

```bash
mkdir -p backups
docker compose down
docker run --rm \
  -v portfolio_sqlite_data:/data:ro \
  -v "$PWD/backups:/backup" \
  alpine:3.22 \
  tar -czf /backup/sqlite-$(date +%Y%m%d-%H%M%S).tgz -C /data .
docker compose up -d
```

`docker compose down` does not delete the named volume. Never add `--volumes`
during normal rollback.

Application rollback:

1. Record the current Git commit and `APP_VERSION`.
2. Stop traffic with `docker compose down`.
3. Check out the last known-good release.
4. If its schema is compatible, run `docker compose up -d --build`.
5. If the schema is incompatible, restore the matching cold backup:

   ```bash
   docker run --rm \
     -v portfolio_sqlite_data:/data \
     -v "$PWD/backups:/backup:ro" \
     alpine:3.22 \
     sh -c 'find /data -mindepth 1 -maxdepth 1 -delete && tar -xzf /backup/BACKUP_FILE.tgz -C /data'
   docker compose up -d --build
   ```

6. Re-run health/API checks and inspect logs.

The restore command is destructive to the SQLite volume; verify `BACKUP_FILE`
and the exact volume name first. Keep at least one off-VPS encrypted backup and
periodically test restore on a disposable volume.

## 6. Production checklist

- [ ] DNS resolves to the VPS.
- [ ] Only SSH, 80, and 443 are allowed by the host/provider firewalls.
- [ ] Docker and the base images are on supported security-patched releases.
- [ ] `.env` is mode `0600`, excluded from Git, and contains a long unique API key.
- [ ] A trusted certificate is mounted; logs contain no self-signed warning.
- [ ] Certificate renewal updates the mounted files and restarts Nginx.
- [ ] HSTS behavior is accepted for the domain before serving real users.
- [ ] `docker compose config` contains only `backend` and `nginx`.
- [ ] Both containers are healthy and the backend has no published host port.
- [ ] UID is non-root, capabilities are empty, and root filesystems are read-only.
- [ ] `/api/health`, project REST endpoints, SPA refresh, and WebSocket terminal work.
- [ ] CORS responses are absent because all browser traffic is same-origin.
- [ ] Rate/body/timeout limits match actual expected usage.
- [ ] Idle and load memory remain below the configured limits.
- [ ] The SQLite volume is on local disk with adequate free space.
- [ ] Cold backup, off-host copy, and restore have been tested.
- [ ] A rollback commit/image version and its matching database backup are recorded.
- [ ] `go test -race ./...`, `go vet ./...`, frontend check/build, and dependency
      vulnerability scanning run before release.
