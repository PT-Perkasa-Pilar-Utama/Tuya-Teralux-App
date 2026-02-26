# syntax=docker/dockerfile:1.7

############################################
# Base Builder
############################################
FROM golang:1.25-bookworm AS base-builder

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
# AI Builder (whisper + llama)
############################################
FROM base-builder AS ai-builder

# Stable pinned versions
ARG WHISPER_TAG=v1.7.4
ARG LLAMA_TAG=b4765

# ---- Build whisper.cpp ----
RUN git clone --depth 1 --branch ${WHISPER_TAG} \
    https://github.com/ggerganov/whisper.cpp.git /whisper.cpp && \
    cmake -S /whisper.cpp -B /whisper.cpp/build -DGGML_NATIVE=OFF && \
    cmake --build /whisper.cpp/build --config Release -t main && \
    strip --strip-unneeded /whisper.cpp/build/bin/main && \
    cp /whisper.cpp/build/bin/main /usr/local/bin/whisper-cli

# ---- Build llama.cpp ----
RUN git clone --depth 1 --branch ${LLAMA_TAG} \
    https://github.com/ggerganov/llama.cpp.git /llama.cpp && \
    cmake -S /llama.cpp -B /llama.cpp/build -DGGML_NATIVE=OFF && \
    cmake --build /llama.cpp/build --config Release -t llama-cli && \
    strip --strip-unneeded /llama.cpp/build/bin/llama-cli && \
    cp /llama.cpp/build/bin/llama-cli /usr/local/bin/llama-cli


############################################
# Go Builder
############################################
FROM base-builder AS app-builder

# Install migrate
RUN curl -fsSL https://github.com/golang-migrate/migrate/releases/download/v4.18.1/migrate.linux-$TARGETARCH.tar.gz \
    | tar -xz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate && \
    strip --strip-unneeded /usr/local/bin/migrate

# Cache deps
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source
COPY . .

# Build app
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
    && rm -rf /var/lib/apt/lists/*

RUN useradd -m -u 1000 appuser && \
    mkdir -p /app/bin /app/models /app/logs /app/uploads && \
    chown -R appuser:appuser /app

# Copy binaries
COPY --from=app-builder /app/app /app/app
COPY --from=app-builder /usr/local/bin/migrate /usr/local/bin/migrate
COPY --from=ai-builder /usr/local/bin/whisper-cli /app/bin/whisper-cli
COPY --from=ai-builder /usr/local/bin/llama-cli /app/bin/llama-cli

# Copy assets
COPY --from=app-builder /app/domain/terminal/migrations /app/migrations
COPY --from=app-builder /app/docs /app/docs
COPY scripts/entrypoint.sh /app/scripts/entrypoint.sh

RUN chmod +x /app/scripts/entrypoint.sh \
    /app/bin/whisper-cli \
    /app/bin/llama-cli && \
    chown appuser:appuser /app/scripts/entrypoint.sh

ENV PORT=8080
ENV AUTO_MIGRATE=true
ENV MODEL_AUTO_DOWNLOAD=true
ENV WHISPER_CLI=/app/bin/whisper-cli
ENV LLAMA_CLI=/app/bin/llama-cli

VOLUME ["/app/models"]

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=60s --retries=3 \
  CMD ["curl", "-f", "http://localhost:8080/api/health"]

ENTRYPOINT ["/app/scripts/entrypoint.sh"]