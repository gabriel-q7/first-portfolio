# Production deployment: GHCR + Docker Compose

This runbook deploys the portfolio to one Ubuntu VPS with 1 vCPU and 1 GB RAM.
The server never builds application images. A push of a stable SemVer tag is
the only event that starts the production workflow.

## Architecture

```text
developer -> push vX.Y.Z tag -> GitHub Actions
                                   | build static Go image
                                   | build Svelte -> Nginx image
                                   v
                         ghcr.io/<owner>/{backend,nginx}:vX.Y.Z
                                   |
                                   | SSH
                                   v
Internet -> VPS :80/:443 -> Nginx -> internal network -> Go -> SQLite volume
```

Only Nginx is attached to the edge network and publishes ports. The backend is
attached only to the `internal: true` application network, has no host port,
and accepts forwarded client metadata only from Nginx at `172.28.0.10`. The
internal network deliberately prevents backend egress; the optional external AI
integration is therefore disabled in this production profile. Add a separately
controlled egress design if that integration becomes a production requirement.

## Repository layout

```text
.
├── .github/workflows/release.yml
├── .env.example
├── docker-compose.yml
├── apps
│   ├── backend
│   │   ├── Dockerfile
│   │   └── cmd/api/main.go
│   └── frontend
│       ├── Dockerfile
│       ├── svelte.config.js
│       └── src/
├── deploy/nginx
│   ├── nginx.conf
│   ├── default.conf
│   ├── security-headers.conf
│   └── start-nginx.sh
└── scripts
    ├── deploy.sh
    └── rollback.sh
```

The VPS contains only runtime material:

```text
/opt/myapp
├── docker-compose.yml
├── .env                         # mode 0600; never committed
├── .previous-version            # written after a successful upgrade
├── scripts
│   ├── deploy.sh
│   └── rollback.sh
└── tls
    ├── fullchain.pem
    └── privkey.pem

/var/lib/docker/volumes/portfolio_sqlite_data/_data
└── portfolio.db                 # Docker-managed; do not edit directly
```

## Image and runtime decisions

### Go backend image

- The builder uses a supported Go 1.26 Alpine toolchain; the final image is
  distroless and contains only CA certificates, runtime metadata, and the binary.
- Copying `go.mod`/`go.sum` before source code preserves Docker's dependency
  layer cache, while the backend-specific `.dockerignore` excludes local secrets,
  binaries, and SQLite files from the build context.
- Unit tests run in the disposable builder stage before compilation; test source
  and tooling never enter the runtime image.
- `CGO_ENABLED=0` works because `modernc.org/sqlite` is pure Go. `netgo` and
  `osusergo` keep name/user lookups independent of libc.
- `-trimpath`, stripped symbols, an empty build ID, and no VCS probing remove
  build-machine paths and unnecessary debug data. The exact Git tag is injected
  into the health response with `-X main.version=...`.
- Distroless has no shell or package manager and runs as UID/GID 65532. `/data`
  is initialized with that ownership so SQLite can create its database, WAL, and
  shared-memory files in a named volume.
- The same binary performs the health check, so curl is not added to the image.
- Go handles SIGTERM/SIGINT, stops accepting work, cancels long-lived handlers,
  waits up to the configured shutdown timeout, and checkpoints the SQLite WAL.
- `http.Server` has explicit header/read/write/idle timeouts and a 16 KiB header
  limit. Middleware caps request bodies at 1 MiB. Logs are structured JSON on
  stdout and never write application log files inside the container.

### Svelte/Nginx image

- Node exists only in the first stage. `npm ci` uses the lock file and produces a
  deterministic production static build through SvelteKit's static adapter.
  Svelte diagnostics must pass before the image can be produced.
- Nginx 1.30 Alpine is the only runtime. No Node server, source tree, npm cache,
  compiler, or TLS-generation utility is copied into production.
- Svelte emits precompressed assets; Nginx enables gzip and `gzip_static`.
  Content-hashed `/_app/immutable/` assets cache for one year, while HTML is
  `no-store`, allowing releases to update without stale entry documents.
- The container runs as UID/GID 101 on unprivileged ports 8080/8443. Startup
  fails if a real certificate/key is absent instead of silently creating an
  untrusted certificate.
- `/` serves static files with `/200.html` SPA fallback. `/api` and `/api/` proxy
  to Go; `/api/ws` uses explicit WebSocket upgrade settings.
- Nginx applies a 1 MiB request limit, bounded proxy timeouts/buffers, connection
  and API request limits, TLS 1.2/1.3, gzip, and JSON access logs.
- Security headers include CSP/frame protection, MIME sniffing protection,
  restrictive browser permissions, referrer policy, cross-origin isolation, and
  HSTS. HSTS is appropriate only after trusted TLS and renewal are verified.

### Compose hardening and sizing

- Compose has `image:` only—there are no `build:` directives. `docker compose
  pull` must succeed before replacement containers are started.
- Both services use `restart: unless-stopped`, read-only root filesystems,
  `no-new-privileges`, all capabilities dropped, PID/resource limits, and small
  writable tmpfs mounts. Only SQLite's named volume is persistently writable.
- Nginx is limited to 64 MiB/0.25 CPU and Go to 256 MiB/0.75 CPU. `GOMAXPROCS=1`
  and `GOMEMLIMIT=192MiB` fit the one-core, one-GB host while preserving RAM for
  Docker, the kernel, page cache, SSH, and deployment overlap.
- Docker's `local` log driver rotates each service at 10 MiB x 3 files.
- The fixed private subnet exists solely to make trusted-proxy validation exact.
  Change both Compose addresses and `TRUSTED_PROXY_CIDR` together if it conflicts
  with another VPS network.

## One-time VPS setup

These examples target supported 64-bit Ubuntu and use Docker's official apt
repository. Run administrative commands from an existing sudo-capable account.

### 1. Base packages, updates, swap, and time

```bash
sudo apt update
sudo apt upgrade -y
sudo apt install -y ca-certificates curl openssh-server ufw certbot
sudo timedatectl set-timezone UTC
```

For a 1 GB VPS, a 1 GB swap file is a useful OOM safety margin (not normal
working memory):

```bash
sudo fallocate -l 1G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
```

Skip creation if `swapon --show` already reports suitable swap.

### 2. Install Docker Engine and Compose plugin

```bash
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg \
  -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc

sudo tee /etc/apt/sources.list.d/docker.sources >/dev/null <<EOF
Types: deb
URIs: https://download.docker.com/linux/ubuntu
Suites: $(. /etc/os-release && echo "${UBUNTU_CODENAME:-$VERSION_CODENAME}")
Components: stable
Architectures: $(dpkg --print-architecture)
Signed-By: /etc/apt/keyrings/docker.asc
EOF

sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io \
  docker-buildx-plugin docker-compose-plugin
sudo systemctl enable --now docker
sudo docker version
sudo docker compose version
```

The convenience install script is intentionally not used for production.

### 3. Create the deployment account and directories

Use a dedicated account such as `deploy`. Membership in the Docker group is
effectively root-equivalent; protect this SSH key as privileged infrastructure
access and do not reuse it for developers.

```bash
sudo adduser --disabled-password --gecos '' deploy
sudo usermod -aG docker deploy
sudo install -d -o deploy -g deploy -m 0750 /opt/myapp
sudo install -d -o deploy -g deploy -m 0750 /opt/myapp/scripts
sudo install -d -o deploy -g deploy -m 0750 /opt/myapp/tls
```

Install a dedicated public key in `/home/deploy/.ssh/authorized_keys`, then open
a new SSH session so Docker group membership takes effect. Disable password SSH
authentication after confirming key access. Keeping SSH on port 22 is fine;
changing the port is not a substitute for keys and firewalling.

Copy the following initial files from a trusted checkout:

```bash
scp docker-compose.yml deploy@SERVER:/opt/myapp/docker-compose.yml
scp scripts/deploy.sh scripts/rollback.sh deploy@SERVER:/opt/myapp/scripts/
ssh deploy@SERVER 'chmod 755 /opt/myapp/scripts/*.sh'
```

### 4. Create runtime configuration

On the VPS:

```bash
cd /opt/myapp
cp /path/to/checked-out/.env.example .env
chmod 600 .env
openssl rand -base64 48
```

Edit `.env` and set:

```dotenv
GHCR_OWNER=lowercase-github-owner
APP_VERSION=v1.0.0
AUTH_API_KEYS=replace-with-the-generated-random-value
LOG_LEVEL=info
HTTP_PORT=80
HTTPS_PORT=443
TLS_DIR=./tls
```

`AUTH_API_KEYS` may contain comma-separated keys for rotation. Secrets are
runtime environment values only; they are not Docker build arguments, image
layers, repository files, or workflow output.

### 5. Install TLS certificate files

Point DNS at the VPS, stop anything currently using port 80, then request a
certificate. Replace the sample domain everywhere:

```bash
sudo certbot certonly --standalone -d portfolio.example.com
sudo install -o 101 -g 101 -m 0644 \
  /etc/letsencrypt/live/portfolio.example.com/fullchain.pem \
  /opt/myapp/tls/fullchain.pem
sudo install -o 101 -g 101 -m 0600 \
  /etc/letsencrypt/live/portfolio.example.com/privkey.pem \
  /opt/myapp/tls/privkey.pem
```

Add a root-owned Certbot deploy hook that repeats both `install` operations and
runs `docker compose -f /opt/myapp/docker-compose.yml --env-file
/opt/myapp/.env restart nginx`. Test renewal with `sudo certbot renew --dry-run`.
Nginx refuses to start if either mounted file is missing or unreadable.

### 6. Firewall

At the cloud-provider firewall, allow TCP 80/443 from the internet and SSH only
from a trusted administrative IP range where practical. On the host:

```bash
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow from ADMIN_IP_OR_CIDR to any port 22 proto tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
sudo ufw status verbose
```

Docker-published ports can bypass normal UFW input processing. This stack
publishes only 80/443, but verify externally with a port scan and enforce the
same allowlist at the provider firewall. For more advanced filtering, place
rules in Docker's `DOCKER-USER` chain rather than disabling Docker's iptables
management.

## GitHub configuration

Create a `production` environment in GitHub. Optional environment reviewers add
a human approval gate without allowing branch pushes to deploy.

Repository secrets:

| Name | Purpose |
|---|---|
| `VPS_HOST` | VPS hostname or IP |
| `VPS_USERNAME` | Dedicated deployment user, e.g. `deploy` |
| `VPS_SSH_PRIVATE_KEY` | Private half of its dedicated SSH key |
| `VPS_KNOWN_HOSTS` | Pinned host-key line from a separately verified source |
| `GHCR_USERNAME` | GitHub user used by the VPS to pull private packages |
| `GHCR_TOKEN` | Classic PAT with only `read:packages` for VPS pulls |

Optional repository variable `VPS_SSH_PORT` defaults to `22`.

The workflow publishes with the repository-scoped `GITHUB_TOKEN` and grants it
only `contents: read` and `packages: write`. Do not create a separate write PAT
for CI. GHCR packages are private by default; ensure both `backend` and `nginx`
are linked to the repository and the pull identity can read them.

Obtain `VPS_KNOWN_HOSTS` from a trusted console or verify its fingerprint out of
band. Do not replace it with unchecked `ssh-keyscan` during every deployment,
because that would accept an attacker's key on the deployment path.

## Release process

Regular pushes and pull requests do not run `.github/workflows/release.yml`.
The YAML declares only `push.tags: v*`, and its first step further rejects
anything except stable `vMAJOR.MINOR.PATCH` with no leading zeroes.

```bash
git checkout main
git pull --ff-only
make backend-test
cd apps/frontend && npm ci && npm run check && npm run build && cd ../..
git tag v1.0.0
git push origin v1.0.0
```

For `v1.0.0`, Actions:

1. checks out exactly that tag;
2. builds `ghcr.io/<owner>/backend:v1.0.0`;
3. compiles Svelte and builds `ghcr.io/<owner>/nginx:v1.0.0`;
4. pushes only those two tags—never `latest`;
5. pins SSH host identity and logs the VPS into GHCR;
6. installs the tagged Compose/deploy files under `/opt/myapp`;
7. runs `deploy.sh v1.0.0 <owner>`;
8. executes `docker compose pull`, then `docker compose up -d`;
9. waits for both container health checks and fails on timeout/unhealthy state.

The workflow uses a production concurrency group so two releases cannot mutate
the VPS simultaneously. Every shell step uses fail-fast settings, and the job
has a 30-minute timeout.

Treat release tags as immutable. Protect `v*` tags/rules in GitHub and never
delete/recreate a published version. A tag points to source while the GHCR tag
selects its corresponding prebuilt images; keeping both immutable makes the
deployment reproducible.

## Rollback

Choose an already-published version; never retag it as `latest`:

```bash
ssh deploy@SERVER
cd /opt/myapp
./scripts/rollback.sh v1.0.0
```

With no argument, `rollback.sh` uses `.previous-version` if available. It calls
the same validation, pull, Compose update, and health checks as a forward
deployment. The current SQLite volume is preserved. If a release introduced a
backward-incompatible schema, restore the database backup taken for that
release before starting the older image; application rollback alone cannot undo
data/schema changes.

You may also rerun the successful GitHub Actions run associated with an earlier
tag if repository policy permits reruns. Do not move or push the old tag again.

## SQLite backup and restore

Before schema-changing releases, take a cold backup. This briefly stops both
containers but guarantees a self-consistent file set:

```bash
cd /opt/myapp
mkdir -p backups
docker compose down
docker run --rm \
  -v portfolio_sqlite_data:/data:ro \
  -v "$PWD/backups:/backup" \
  alpine:3.23 \
  tar -czf /backup/sqlite-v1.0.0.tgz -C /data .
docker compose up -d
```

`docker compose down` does not remove the named volume. Never add `--volumes`
to normal deploy, stop, or rollback commands. Copy backups encrypted to another
machine/location and test restoration on a disposable volume.

Restore is destructive and must be performed only after checking the archive
and exact volume name:

```bash
docker compose down
docker run --rm \
  -v portfolio_sqlite_data:/data \
  -v "$PWD/backups:/backup:ro" \
  alpine:3.23 \
  sh -c 'find /data -mindepth 1 -maxdepth 1 -delete && tar -xzf /backup/sqlite-v1.0.0.tgz -C /data'
docker compose up -d
```

## Operations

```bash
cd /opt/myapp
docker compose ps
docker compose logs --tail=100 nginx backend
docker stats --no-stream
curl -fsS https://portfolio.example.com/api/health
curl -fsS https://portfolio.example.com/api/v1/projects
```

Docker starts at boot through systemd, and `restart: unless-stopped` brings both
containers back. The backend waits for its volume and Nginx waits for backend
health. Apply Ubuntu/Docker security updates regularly and schedule deliberate
reboots to verify restart behavior.

## Production checklist

- [ ] `main` is reviewed and local backend/frontend checks pass.
- [ ] The release is a new immutable stable tag matching `vX.Y.Z`.
- [ ] GHCR contains exactly-versioned `backend` and `nginx` images; no process
      depends on `latest`.
- [ ] The production Compose file has no `build:` key.
- [ ] The VPS `.env` is mode 0600, excluded from Git, and contains a strong API key.
- [ ] The VPS GHCR PAT has `read:packages` only and the CI token has scoped permissions.
- [ ] The SSH host key is pinned; password authentication is disabled.
- [ ] Provider firewall and external scan show only SSH, 80, and 443.
- [ ] Trusted TLS and automated renewal have been tested before relying on HSTS.
- [ ] Both containers run non-root with empty capabilities and read-only roots.
- [ ] The backend has no published port and is only on the internal network.
- [ ] `/api/health`, REST, SPA refresh, and WebSocket flows work through Nginx.
- [ ] Container logs are structured/rotated and contain no tokens or credentials.
- [ ] Memory remains below limits during deployment overlap and swap is not routinely used.
- [ ] SQLite resides on local disk; free space and volume backups are monitored.
- [ ] A cold backup and off-VPS copy exist before schema changes.
- [ ] A rollback to an earlier exact tag and a database restore have been rehearsed.

## Authoritative operational references

- [Install Docker Engine on Ubuntu](https://docs.docker.com/engine/install/ubuntu/)
- [Install the Docker Compose plugin](https://docs.docker.com/compose/install/linux/)
- [Docker packet filtering and UFW behavior](https://docs.docker.com/engine/network/packet-filtering-firewalls/)
- [GitHub: publishing Docker images](https://docs.github.com/en/actions/tutorials/publish-packages/publish-docker-images)
- [GitHub Actions tag filters](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax#onpushbranchestagsbranches-ignoretags-ignore)
- [GitHub Container Registry authentication](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
