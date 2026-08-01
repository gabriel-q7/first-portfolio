# TLS certificate mount

Place the VPS certificate chain at `fullchain.pem` and its private key at
`privkey.pem`. The container runs as UID 101, so make the key readable only by
that UID with `sudo chown 101:101 privkey.pem && chmod 600 privkey.pem`. Never
commit it.

If these files are absent, the Nginx container generates a seven-day
self-signed certificate at startup for smoke testing only.
