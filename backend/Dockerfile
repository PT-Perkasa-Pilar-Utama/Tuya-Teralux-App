# syntax=docker/dockerfile:1.7

############################
# Builder
############################
FROM golang:1.25-bookworm AS builder

WORKDIR /app

# Cache dependency layer
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source
COPY . .

# Run Unit Tests
RUN --mount=type=cache,target=/root/.cache/go-build \
    go test -v ./...

# Build static binary
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build \
      -trimpath \
      -ldflags="-s -w" \
      -o app

# Install migrate (builder only)
RUN curl -fsSL https://github.com/golang-migrate/migrate/releases/download/v4.18.1/migrate.linux-amd64.tar.gz \
    | tar -xz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate

############################
# Runtime
############################
FROM gcr.io/distroless/base-debian12:nonroot

WORKDIR /app

# Copy binary
COPY --from=builder /app/app /app/app

# Copy migrate
COPY --from=builder /usr/local/bin/migrate /usr/local/bin/migrate

# Copy migrations
COPY --from=builder /app/domain/teralux/migrations /app/migrations

# Runtime env only
ENV AUTO_MIGRATE=false
ENV PORT=8080

EXPOSE 8080

# Healthcheck (distroless-safe)
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD ["/app/app", "healthcheck"]

ENTRYPOINT ["/app/app"]
