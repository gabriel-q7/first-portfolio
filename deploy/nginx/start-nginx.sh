#!/bin/sh
set -eu

umask 077
mkdir -p /tmp/tls
mkdir -p \
    /tmp/nginx/client_temp \
    /tmp/nginx/proxy_temp \
    /tmp/nginx/fastcgi_temp \
    /tmp/nginx/uwsgi_temp \
    /tmp/nginx/scgi_temp

certificate="${TLS_CERT_FILE:-/etc/nginx/tls/fullchain.pem}"
private_key="${TLS_KEY_FILE:-/etc/nginx/tls/privkey.pem}"

if [ -r "$certificate" ] && [ -r "$private_key" ]; then
    cp "$certificate" /tmp/tls/server.crt
    cp "$private_key" /tmp/tls/server.key
else
    echo "WARNING: no TLS certificate found; generating a temporary self-signed certificate" >&2
    openssl req -x509 -newkey rsa:2048 -sha256 -nodes -days 7 \
        -subj "/CN=localhost" \
        -keyout /tmp/tls/server.key \
        -out /tmp/tls/server.crt >/dev/null 2>&1
fi

exec nginx -g "daemon off;"
