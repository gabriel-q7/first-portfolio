#!/bin/sh
set -eu

APP_DIR=${APP_DIR:-/opt/myapp}
VERSION=${1:-}

if [ -z "$VERSION" ] && [ -f "$APP_DIR/.previous-version" ]; then
    VERSION=$(sed -n '1p' "$APP_DIR/.previous-version")
fi

if [ -z "$VERSION" ]; then
    echo "Usage: $0 v1.2.3" >&2
    exit 2
fi

owner=$(sed -n 's/^GHCR_OWNER=//p' "$APP_DIR/.env" | tail -n 1)
if [ -z "$owner" ]; then
    echo "ERROR: GHCR_OWNER is missing from $APP_DIR/.env" >&2
    exit 1
fi

exec "$APP_DIR/scripts/deploy.sh" "$VERSION" "$owner"
