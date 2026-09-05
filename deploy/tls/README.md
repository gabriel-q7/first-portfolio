# TLS certificate mount

Place the VPS certificate chain at `fullchain.pem` and its private key at
`privkey.pem`. The container runs as UID 101, so make the key readable only by
that UID with `sudo chown 101:101 privkey.pem && chmod 600 privkey.pem`. Never
commit it. Nginx fails closed when either file is absent or unreadable.
