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
# AI Builder (whisper + llama) - Native Target
############################################
FROM native-builder AS ai-builder

# Stable pinned versions
ARG WHISPER_TAG=v1.7.4
ARG LLAMA_TAG=b4765

# ---- Build whisper.cpp ----
COPY services/whisper.cpp /whisper.cpp
RUN echo "🔨 Building whisper.cpp for $TARGETARCH..." && \
    cmake -S /whisper.cpp -B /whisper.cpp/build -DGGML_NATIVE=OFF && \
    cmake --build /whisper.cpp/build --config Release -t main && \
    strip --strip-unneeded /whisper.cpp/build/bin/main && \
    cp /whisper.cpp/build/bin/main /usr/local/bin/whisper-cli

# ---- Build llama.cpp ----
COPY services/llama.cpp /llama.cpp
RUN echo "🔨 Configuring llama.cpp for $TARGETARCH..." && \
    cmake -S /llama.cpp -B /llama.cpp/build \
        -DGGML_NATIVE=OFF \
        -DLLAMA_BUILD_EXAMPLES=ON \
        -DLLAMA_BUILD_SERVER=OFF || \
    ( echo "❌ CMake configuration failed"; exit 1 ) && \
    echo "🔨 Building llama.cpp..." && \
    ( cmake --build /llama.cpp/build --config Release -t llama-cli || \
      cmake --build /llama.cpp/build --config Release -t main || \
      { echo "❌ Build failed. Diagnostics:"; \
        cmake --build /llama.cpp/build --target help | grep -E "llama-cli|main|server" || true; \
        ls -la /llama.cpp/build/bin || true; \
        exit 1; } ) && \
    if [ -f /llama.cpp/build/bin/llama-cli ]; then \
        BIN_TARGET=/llama.cpp/build/bin/llama-cli; \
    else \
        BIN_TARGET=/llama.cpp/build/bin/main; \
    fi && \
    test -x "$BIN_TARGET" && \
    strip --strip-unneeded "$BIN_TARGET" && \
    cp "$BIN_TARGET" /usr/local/bin/llama-cli && \
    echo "🔨 Copying llama shared libraries..." && \
    mkdir -p /usr/local/lib/llama && \
    (cp /llama.cpp/build/bin/libllama.so* /usr/local/lib/llama/ 2>/dev/null || true) && \
    (cp /llama.cpp/build/src/libllama.so* /usr/local/lib/llama/ 2>/dev/null || true) && \
    (cp /llama.cpp/build/ggml/src/libggml.so* /usr/local/lib/llama/ 2>/dev/null || true) && \
    (cp /llama.cpp/build/ggml/src/libggml-base.so* /usr/local/lib/llama/ 2>/dev/null || true) && \
    (cp /llama.cpp/build/ggml/src/libggml-cpu.so* /usr/local/lib/llama/ 2>/dev/null || true) && \
    ls -la /usr/local/lib/llama/ || true


############################################
# Go Builder - Native Target
############################################
FROM native-builder AS app-builder

# Install migrate
RUN echo "📥 Downloading migrate for $TARGETARCH..." && \
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
COPY --from=ai-builder /usr/local/bin/whisper-cli /app/bin/whisper-cli
COPY --from=ai-builder /usr/local/bin/llama-cli /app/bin/llama-cli
COPY --from=ai-builder /usr/local/lib/llama /app/lib/llama

# Copy assets
COPY --from=app-builder /app/migrations /app/migrations
COPY --from=app-builder /app/domain/models/rag/skills/definitions /app/domain/models/rag/skills/definitions
COPY --from=app-builder /app/assets /app/assets
COPY --from=app-builder /app/docs /app/docs
COPY scripts/entrypoint.sh /app/scripts/entrypoint.sh

RUN chmod +x /app/scripts/entrypoint.sh \
    /app/bin/whisper-cli \
    /app/bin/llama-cli && \
    chown appuser:appuser /app/scripts/entrypoint.sh

ENV PORT=8080
ARG AUTO_MIGRATE=true
ENV AUTO_MIGRATE=$AUTO_MIGRATE
ENV MODEL_AUTO_DOWNLOAD=true
ENV WHISPER_CLI=/app/bin/whisper-cli
ENV LLAMA_CLI=/app/bin/llama-cli
ENV LD_LIBRARY_PATH=/app/lib/llama:${LD_LIBRARY_PATH:-}

VOLUME ["/app/models"]

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=60s --retries=3 \
  CMD ["curl", "-f", "http://localhost:8080/api/health"]

ENTRYPOINT ["/app/scripts/entrypoint.sh"]
