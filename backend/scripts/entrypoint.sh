#!/bin/bash
set -e

# Logging colors
INFO='\033[0;34m'
SUCCESS='\033[0;32m'
NC='\033[0m' # No Color

echo -e "${INFO}[INFO] Starting Teralux Backend Container Initialization...${NC}"

# 1. Model Check
MODEL_DIR="/app/bin/models"
MODEL_PATH="${MODEL_DIR}/ggml-base.bin"

if [ ! -f "$MODEL_PATH" ]; then
    echo -e "${INFO}[INFO] Whisper model not found at ${MODEL_PATH}. Downloading...${NC}"
    mkdir -p "$MODEL_DIR"
    curl -L "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.bin" -o "$MODEL_PATH"
    echo -e "${SUCCESS}[SUCCESS] Model downloaded successfully.${NC}"
else
    echo -e "${SUCCESS}[SUCCESS] Whisper model found.${NC}"
fi

# 2. Binary Check & Build
WHISPER_CLI="/app/bin/whisper-cli"

if [ ! -f "$WHISPER_CLI" ]; then
    echo -e "${INFO}[INFO] whisper-cli not found at ${WHISPER_CLI}. Building from source...${NC}"
    mkdir -p "/app/tmp/whisper_build"
    cd "/app/tmp/whisper_build"
    
    echo -e "${INFO}[INFO] Cloning whisper.cpp...${NC}"
    git clone --depth 1 https://github.com/ggerganov/whisper.cpp.git .
    
    echo -e "${INFO}[INFO] Configuring and building...${NC}"
    cmake -B build
    cmake --build build --config Release -t main
    
    echo -e "${INFO}[INFO] Moving binary to ${WHISPER_CLI}...${NC}"
    cp build/bin/main "$WHISPER_CLI"
    chmod +x "$WHISPER_CLI"
    
    echo -e "${INFO}[INFO] Cleaning up build files...${NC}"
    cd "/app"
    rm -rf "/app/tmp/whisper_build"
    echo -e "${SUCCESS}[SUCCESS] whisper-cli built successfully.${NC}"
else
    echo -e "${SUCCESS}[SUCCESS] whisper-cli found.${NC}"
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
