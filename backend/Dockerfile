# syntax=docker/dockerfile:1.7

############################
# Builder
############################
FROM golang:1.25-bookworm AS builder

WORKDIR /app

ARG TARGETARCH

# Install build tools and binutils (for strip)
RUN apt-get update && apt-get install -y --no-install-recommends \
    cmake \
    build-essential \
    git \
    binutils \
    && rm -rf /var/lib/apt/lists/*

# Cache dependency layer for Go
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source
COPY . .

# Build static Go binary
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=$TARGETARCH \
    go build \
      -trimpath \
      -ldflags="-s -w" \
      -o app && \
    strip --strip-unneeded app

# Install migrate (builder only)
RUN curl -fsSL https://github.com/golang-migrate/migrate/releases/download/v4.18.1/migrate.linux-$TARGETARCH.tar.gz \
    | tar -xz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate && \
    strip --strip-unneeded /usr/local/bin/migrate

# Build whisper-cli (Pinned to v1.7.4)
RUN git clone --depth 1 --branch v1.7.4 https://github.com/ggerganov/whisper.cpp.git /whisper.cpp && \
    cd /whisper.cpp && \
    cmake -B build -DGGML_NATIVE=OFF && cmake --build build --config Release -t main && \
    cp build/bin/main /usr/local/bin/whisper-cli && \
    strip --strip-unneeded /usr/local/bin/whisper-cli

# Build llama-cli (Pinned to b4765)
RUN git clone --depth 1 --branch b4765 https://github.com/ggerganov/llama.cpp.git /llama.cpp && \
    cd /llama.cpp && \
    cmake -B build -DGGML_NATIVE=OFF && cmake --build build --config Release -t llama-cli && \
    cp build/bin/llama-cli /usr/local/bin/llama-cli && \
    strip --strip-unneeded /usr/local/bin/llama-cli

############################
# Runtime
############################
FROM debian:bookworm-slim

WORKDIR /app

# Install minimal runtime deps
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user and setup directories
RUN useradd -m -u 1000 appuser && \
    mkdir -p /app/bin /app/models /app/logs && \
    chown -R appuser:appuser /app

# Copy binaries from builder
COPY --from=builder /app/app /app/app
COPY --from=builder /usr/local/bin/migrate /usr/local/bin/migrate
COPY --from=builder /usr/local/bin/whisper-cli /app/bin/whisper-cli
COPY --from=builder /usr/local/bin/llama-cli /app/bin/llama-cli

# Copy application assets/data
COPY --from=builder /app/domain/teralux/migrations /app/migrations
COPY --from=builder /app/docs /app/docs
COPY scripts/entrypoint.sh /app/scripts/entrypoint.sh

RUN chmod +x /app/scripts/entrypoint.sh /app/bin/whisper-cli /app/bin/llama-cli && \
    chown appuser:appuser /app/scripts/entrypoint.sh

# Runtime env
ENV PORT=8080
ENV AUTO_MIGRATE=true
ENV MODEL_AUTO_DOWNLOAD=true
ENV WHISPER_CLI=/app/bin/whisper-cli
ENV LLAMA_CLI=/app/bin/llama-cli

# State model volume contract
VOLUME ["/app/models"]

USER appuser

EXPOSE 8080

# Healthcheck (Exec form)
HEALTHCHECK --interval=30s --timeout=5s --start-period=60s --retries=3 \
  CMD ["curl", "-f", "http://localhost:8080/api/health"]

ENTRYPOINT ["/app/scripts/entrypoint.sh"]
