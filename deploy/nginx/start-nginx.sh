#!/bin/sh
set -eu

umask 077
mkdir -p \
    /tmp/nginx/client_temp \
    /tmp/nginx/proxy_temp \
    /tmp/nginx/fastcgi_temp \
    /tmp/nginx/uwsgi_temp \
    /tmp/nginx/scgi_temp

certificate="${TLS_CERT_FILE:-/etc/nginx/tls/fullchain.pem}"
private_key="${TLS_KEY_FILE:-/etc/nginx/tls/privkey.pem}"

if [ ! -r "$certificate" ] || [ ! -r "$private_key" ]; then
    echo "ERROR: TLS_CERT_FILE and TLS_KEY_FILE must reference readable certificate files" >&2
    exit 1
fi

exec nginx -g "daemon off;"
