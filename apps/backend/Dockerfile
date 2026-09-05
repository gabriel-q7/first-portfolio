FROM golang:1.26-alpine AS builder

ARG VERSION=dev
ARG TARGETOS=linux
ARG TARGETARCH

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go test ./...
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build -trimpath -buildvcs=false \
    -tags="netgo,osusergo" \
    -ldflags="-s -w -buildid= -X main.version=${VERSION}" \
    -o /out/server ./cmd/api

# Docker initializes a new named volume from this owned directory, allowing the
# non-root runtime user to create SQLite, WAL, and shared-memory files.
RUN mkdir -p /runtime-data && \
    touch /runtime-data/.volume-init && \
    chown -R 65532:65532 /runtime-data

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=builder --chown=65532:65532 /out/server /app/server
COPY --from=builder --chown=65532:65532 /runtime-data /data

USER 65532:65532
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/server", "healthcheck"]
ENTRYPOINT ["/app/server"]
