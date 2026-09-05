FROM node:22-alpine AS frontend-builder

ARG VERSION=dev

WORKDIR /src
COPY apps/frontend/package.json apps/frontend/package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY apps/frontend/ ./
RUN npm run check && npm run build

FROM nginx:1.30-alpine

ARG VERSION=dev
LABEL org.opencontainers.image.title="portfolio-nginx" \
      org.opencontainers.image.version="${VERSION}"

COPY deploy/nginx/nginx.conf /etc/nginx/nginx.conf
COPY deploy/nginx/default.conf /etc/nginx/conf.d/default.conf
COPY deploy/nginx/security-headers.conf /etc/nginx/security-headers.conf
COPY deploy/nginx/start-nginx.sh /usr/local/bin/start-nginx
RUN chmod 0755 /usr/local/bin/start-nginx
COPY --from=frontend-builder --chown=101:101 /src/build/ /usr/share/nginx/html/

USER 101:101
EXPOSE 8080 8443
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -q -O /dev/null http://127.0.0.1:8080/nginx-health || exit 1
ENTRYPOINT ["/usr/local/bin/start-nginx"]
