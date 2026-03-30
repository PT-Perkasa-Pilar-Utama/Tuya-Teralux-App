# syntax=docker/dockerfile:1.7

############################################
# Base Builder (Cross-compilation for Go)
############################################
FROM --platform=$BUILDPLATFORM golang:1.25-bookworm AS base-builder

WORKDIR /app
ARG TARGETOS
ARG TARGETARCH
ARG TARGETPLATFORM
ARG BUILDPLATFORM

RUN apt-get update && apt-get install -y --no-install-recommends \
    cmake \
    build-essential \
    git \
    binutils \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*


############################################
# Native Builder (For C++ compilation)
############################################
FROM golang:1.25-bookworm AS native-builder

WORKDIR /app
ARG TARGETARCH

RUN apt-get update && apt-get install -y --no-install-recommends \
    cmake \
    build-essential \
    git \
    binutils \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*



############################################
# Go Builder - Native Target
############################################
FROM native-builder AS app-builder

ARG TARGETARCH

# Install migrate
RUN set -e; \
    echo "📥 Downloading migrate for ${TARGETARCH:-unknown}..." && \
    curl -fsSL https://github.com/golang-migrate/migrate/releases/download/v4.18.1/migrate.linux-${TARGETARCH}.tar.gz \
    | tar -xz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate && \
    strip --strip-unneeded /usr/local/bin/migrate || true

# Cache deps
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source
COPY . .

# Build app
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    echo "🔨 Building Go app for $TARGETARCH (CGO_ENABLED=0)..." && \
    go env && \
    CGO_ENABLED=0 \
    GOOS=linux \
    go build \
      -trimpath \
      -ldflags="-s -w" \
      -o app && \
    strip --strip-unneeded app


############################################
# Runtime
############################################
FROM debian:bookworm-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    ca-certificates \
    chromium \
    fonts-liberation \
    libnss3 \
    libatk-bridge2.0-0 \
    libcups2 \
    libdrm2 \
    libxcomposite1 \
    libxdamage1 \
    libxrandr2 \
    libgbm1 \
    libasound2 \
    libpangocairo-1.0-0 \
    ffmpeg \
    && rm -rf /var/lib/apt/lists/*

RUN useradd -m -u 1000 appuser && \
    mkdir -p /app/bin /app/models /app/logs /app/uploads /app/uploads/reports && \
    chown -R appuser:appuser /app

# Copy binaries
COPY --from=app-builder /app/app /app/app
COPY --from=app-builder /usr/local/bin/migrate /usr/local/bin/migrate

# Copy assets
COPY --from=app-builder /app/migrations /app/migrations
COPY --from=app-builder /app/domain/models/rag/skills/definitions /app/domain/models/rag/skills/definitions
COPY --from=app-builder /app/assets /app/assets
COPY --from=app-builder /app/docs /app/docs
COPY scripts/entrypoint.sh /app/scripts/entrypoint.sh

RUN chmod +x /app/scripts/entrypoint.sh && \
    chown appuser:appuser /app/scripts/entrypoint.sh

ENV PORT=8080
ARG AUTO_MIGRATE=true
ENV AUTO_MIGRATE=$AUTO_MIGRATE
ENV MODEL_AUTO_DOWNLOAD=true
VOLUME ["/app/models"]

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=60s --retries=3 \
  CMD ["curl", "-f", "http://localhost:8080/api/health"]

ENTRYPOINT ["/app/scripts/entrypoint.sh"]
