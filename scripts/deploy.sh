#!/bin/sh
set -eu

APP_DIR=${APP_DIR:-/opt/myapp}
VERSION=${1:-}
GHCR_OWNER_VALUE=${2:-}

if ! printf '%s\n' "$VERSION" | grep -Eq '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$'; then
    echo "ERROR: version must be a stable semantic version such as v1.2.3" >&2
    exit 2
fi

if ! printf '%s\n' "$GHCR_OWNER_VALUE" | grep -Eq '^[a-z0-9]([a-z0-9-]{0,37}[a-z0-9])?$'; then
    echo "ERROR: GHCR owner must be a lowercase GitHub account or organization name" >&2
    exit 2
fi

cd "$APP_DIR"
umask 077

if [ ! -f .env ]; then
    echo "ERROR: $APP_DIR/.env does not exist; complete the VPS setup first" >&2
    exit 1
fi

previous_version=$(sed -n 's/^APP_VERSION=//p' .env | tail -n 1)
env_backup=$(mktemp "$APP_DIR/.env.backup.XXXXXX")
env_next=$(mktemp "$APP_DIR/.env.next.XXXXXX")
cp .env "$env_backup"

awk -v version="$VERSION" -v owner="$GHCR_OWNER_VALUE" '
    BEGIN { version_written = 0; owner_written = 0 }
    /^APP_VERSION=/ { print "APP_VERSION=" version; version_written = 1; next }
    /^GHCR_OWNER=/ { print "GHCR_OWNER=" owner; owner_written = 1; next }
    { print }
    END {
        if (!owner_written) print "GHCR_OWNER=" owner
        if (!version_written) print "APP_VERSION=" version
    }
' .env >"$env_next"
chmod 600 "$env_next"
mv "$env_next" .env

activated=0
rollback_on_error() {
    exit_code=$?
    trap - EXIT INT TERM
    echo "Deployment failed; restoring ${previous_version:-the previous configuration}" >&2
    cp "$env_backup" .env
    chmod 600 .env
    if [ "$activated" -eq 1 ]; then
        docker compose up -d --remove-orphans || true
    fi
    rm -f "$env_backup"
    exit "$exit_code"
}
trap rollback_on_error EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

docker compose config --quiet
docker compose pull
activated=1
docker compose up -d --remove-orphans

backend_id=$(docker compose ps -q backend)
nginx_id=$(docker compose ps -q nginx)
if [ -z "$backend_id" ] || [ -z "$nginx_id" ]; then
    echo "ERROR: one or more application containers did not start" >&2
    exit 1
fi

attempt=1
while [ "$attempt" -le 30 ]; do
    backend_health=$(docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}missing{{end}}' "$backend_id")
    nginx_health=$(docker inspect --format '{{if .State.Health}}{{.State.Health.Status}}{{else}}missing{{end}}' "$nginx_id")
    if [ "$backend_health" = healthy ] && [ "$nginx_health" = healthy ]; then
        break
    fi
    if [ "$backend_health" = unhealthy ] || [ "$nginx_health" = unhealthy ]; then
        echo "ERROR: unhealthy container (backend=$backend_health nginx=$nginx_health)" >&2
        exit 1
    fi
    if [ "$attempt" -eq 30 ]; then
        echo "ERROR: health checks timed out (backend=$backend_health nginx=$nginx_health)" >&2
        exit 1
    fi
    attempt=$((attempt + 1))
    sleep 2
done

if [ -n "$previous_version" ] && [ "$previous_version" != "$VERSION" ]; then
    printf '%s\n' "$previous_version" >.previous-version
    chmod 600 .previous-version
fi

rm -f "$env_backup"
trap - EXIT INT TERM
docker compose ps
echo "Deployed $VERSION successfully"
