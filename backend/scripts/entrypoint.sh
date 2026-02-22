#!/bin/bash
set -e

# Logging colors
INFO='\033[0;34m'
SUCCESS='\033[0;32m'
NC='\033[0m' # No Color

echo -e "${INFO}[INFO] Starting Teralux Backend Container Initialization...${NC}"

# 1. Model Check
MODEL_DIR="/app/bin"
WHISPER_MODEL="${MODEL_DIR}/ggml-base.bin"
LLAMA_MODEL="${MODEL_DIR}/llama-model.gguf"

mkdir -p "$MODEL_DIR"

if [ ! -f "$WHISPER_MODEL" ]; then
    echo -e "${INFO}[INFO] Whisper model not found at ${WHISPER_MODEL}. Downloading...${NC}"
    curl -L "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.bin" -o "$WHISPER_MODEL"
    echo -e "${SUCCESS}[SUCCESS] Whisper model downloaded.${NC}"
fi

if [ ! -f "$LLAMA_MODEL" ]; then
    echo -e "${INFO}[INFO] Llama model not found at ${LLAMA_MODEL}. Downloading...${NC}"
    curl -L "https://huggingface.co/unsloth/Llama-3.2-1B-Instruct-GGUF/resolve/main/Llama-3.2-1B-Instruct-Q4_K_M.gguf" -o "$LLAMA_MODEL"
    echo -e "${SUCCESS}[SUCCESS] Llama model downloaded.${NC}"
fi

# 2. Binary Check & Build
WHISPER_CLI="/app/bin/whisper-cli"
LLAMA_CLI="/app/bin/llama-cli"

# Build whisper-cli if missing
if [ ! -f "$WHISPER_CLI" ]; then
    echo -e "${INFO}[INFO] whisper-cli not found. Building...${NC}"
    mkdir -p "/app/tmp/whisper_build"
    cd "/app/tmp/whisper_build"
    git clone --depth 1 https://github.com/ggerganov/whisper.cpp.git .
    cmake -B build && cmake --build build --config Release -t main
    cp build/bin/main "$WHISPER_CLI" && chmod +x "$WHISPER_CLI"
    cd "/app" && rm -rf "/app/tmp/whisper_build"
    echo -e "${SUCCESS}[SUCCESS] whisper-cli built.${NC}"
fi

# Build llama-cli if missing
if [ ! -f "$LLAMA_CLI" ]; then
    echo -e "${INFO}[INFO] llama-cli not found. Building...${NC}"
    mkdir -p "/app/tmp/llama_build"
    cd "/app/tmp/llama_build"
    git clone --depth 1 https://github.com/ggerganov/llama.cpp.git .
    cmake -B build && cmake --build build --config Release -t llama-cli
    cp build/bin/llama-cli "$LLAMA_CLI" && chmod +x "$LLAMA_CLI"
    cd "/app" && rm -rf "/app/tmp/llama_build"
    echo -e "${SUCCESS}[SUCCESS] llama-cli built.${NC}"
fi

# 3. Database Migration
echo -e "${INFO}[INFO] Running database migrations...${NC}"

# Construct DB connection for migrate tool based on DB_TYPE
MIGRATIONS_PATH="/app/migrations"

if [ "$DB_TYPE" = "mysql" ]; then
    MYSQL_DSN="mysql://${DB_USER}:${DB_PASSWORD}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}"
    migrate -path "$MIGRATIONS_PATH" -database "$MYSQL_DSN" up
elif [ "$DB_TYPE" = "sqlite" ]; then
    SQLITE_DSN="sqlite3://${DB_SQLITE_PATH}"
    migrate -path "$MIGRATIONS_PATH" -database "$SQLITE_DSN" up
else
    echo -e "${INFO}[INFO] Skipping migration: DB_TYPE '${DB_TYPE}' not recognized or handled for auto-migrate.${NC}"
fi

# 4. Start Application
echo -e "${SUCCESS}[SUCCESS] Initialization complete. Starting Teralux Backend...${NC}"
exec /app/app
