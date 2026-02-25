#!/bin/bash
set -euo pipefail

# Logging colors
INFO='\033[0;34m'
SUCCESS='\033[0;32m'
WARNING='\033[0;33m'
ERROR='\033[0;31m'
NC='\033[0m'

echo -e "${INFO}[INFO] Starting Terminal Backend (10/10 Production Grade)...${NC}"

# 1. Environment & Path Setup
MODEL_DIR="/app/models"
LOCK_DIR="${MODEL_DIR}/.lock"
WHISPER_MODEL="${MODEL_DIR}/ggml-base.bin"
LLAMA_MODEL="${MODEL_DIR}/llama-model.gguf"

# Model Metadata (Pinned)
WHISPER_URL="https://huggingface.co/ggerganov/whisper.cpp/resolve/80da2d8bfee42b0e836fc3a9890373e5defc00a6/ggml-base.bin"
WHISPER_SHA="60ed5bc3dd14eea856493d334349b405782ddcaf0028d4b5df4088345fba2efe"

LLAMA_URL="https://huggingface.co/unsloth/Llama-3.2-1B-Instruct-GGUF/resolve/b9a1b2bae395064b27533b8a56b1306f4d2d2f5f/Llama-3.2-1B-Instruct-Q4_K_M.gguf"
LLAMA_SHA="3f5a22426976ab26cfe84dba63c1d08391717abb1af893e10f1b2968d862dcc1"

# Standard paths from environment or defaults
WHISPER_CLI="${WHISPER_CLI:-/app/bin/whisper-cli}"
LLAMA_CLI="${LLAMA_CLI:-/app/bin/llama-cli}"

mkdir -p "$MODEL_DIR"

# 2. Binary Verification
echo -e "${INFO}[INFO] Verifying binaries...${NC}"
for bin in "$WHISPER_CLI" "$LLAMA_CLI" "/app/app" "/usr/local/bin/migrate"; do
    if [ ! -f "$bin" ]; then
        echo -e "${ERROR}[ERROR] Required binary missing: ${bin}${NC}"
        exit 1
    fi
done
echo -e "${SUCCESS}[SUCCESS] Binaries verified.${NC}"

# 3. Model Management (Atomic, SHA-Verified & Race-Protected)
function acquire_lock() {
    local max_retries=30
    local count=0

    # Handling stale locks (older than 10 mins)
    if [ -d "$LOCK_DIR" ]; then
        # Check if lock directory is stale (using find for portability)
        # If the directory was created more than 10 minutes ago, remove it.
        if [ "$(find "$LOCK_DIR" -maxdepth 0 -mmin +10 2>/dev/null)" ]; then
            echo -e "${WARNING}[WARNING] Stale lock detected (older than 10m). Cleaning up...${NC}"
            rmdir "$LOCK_DIR" 2>/dev/null || true
        fi
    fi

    while ! mkdir "$LOCK_DIR" 2>/dev/null; do
        if [ $count -eq 0 ]; then
            echo -e "${WARNING}[WARNING] Another instance is managing models. Waiting...${NC}"
        fi
        sleep 2
        count=$((count + 1))
        if [ $count -ge $max_retries ]; then
            echo -e "${ERROR}[ERROR] Timeout waiting for model lock.${NC}"
            exit 1
        fi
    done
}

function release_lock() {
    rmdir "$LOCK_DIR" 2>/dev/null || true
}

function verify_checksum() {
    local file=$1
    local expected_sha=$2
    local name=$3

    echo -e "${INFO}[INFO] Verifying checksum for ${name}...${NC}"
    if echo "${expected_sha}  ${file}" | sha256sum -c --status; then
        echo -e "${SUCCESS}[SUCCESS] ${name} checksum verified.${NC}"
        return 0
    else
        echo -e "${ERROR}[ERROR] ${name} checksum mismatch!${NC}"
        return 1
    fi
}

function download_model() {
    local url=$1
    local output=$2
    local expected_sha=$3
    local name=$4

    # If file exists, verify it
    if [ -f "$output" ]; then
        if verify_checksum "$output" "$expected_sha" "$name"; then
            echo -e "${SUCCESS}[SUCCESS] ${name} model ready.${NC}"
            return
        else
            echo -e "${WARNING}[WARNING] Existing ${name} model is corrupt. Deleting and re-downloading...${NC}"
            rm -f "$output"
        fi
    fi

    if [ "${MODEL_AUTO_DOWNLOAD:-true}" != "true" ]; then
        echo -e "${ERROR}[ERROR] ${name} model missing/corrupt and MODEL_AUTO_DOWNLOAD is disabled.${NC}"
        exit 1
    fi

    acquire_lock
    # Double-check after acquiring lock in case another container finished it
    if [ -f "$output" ]; then
        if verify_checksum "$output" "$expected_sha" "$name"; then
            echo -e "${SUCCESS}[SUCCESS] ${name} model ready (provided by another instance).${NC}"
            release_lock
            return
        fi
        rm -f "$output"
    fi

    echo -e "${INFO}[INFO] ${name} model not found. Downloading to temporary file...${NC}"
    local tmp_output="${output}.tmp"
    local success=false
    local attempt=1
    local max_attempts=3

    while [ $attempt -le $max_attempts ]; do
        if curl -L --retry 3 --fail --output "$tmp_output" "$url"; then
            if verify_checksum "$tmp_output" "$expected_sha" "$name"; then
                mv "$tmp_output" "$output"
                echo -e "${SUCCESS}[SUCCESS] ${name} model downloaded and verified.${NC}"
                success=true
                break
            fi
        fi
        echo -e "${WARNING}[WARNING] Download attempt $attempt failed or verification failed for ${name}. Retrying...${NC}"
        rm -f "$tmp_output" 2>/dev/null || true
        sleep $((attempt * 5))
        attempt=$((attempt + 1))
    done

    release_lock

    if [ "$success" = false ]; then
        echo -e "${ERROR}[ERROR] Failed to download/verify ${name} model after $max_attempts attempts.${NC}"
        exit 1
    fi
}

download_model "$WHISPER_URL" "$WHISPER_MODEL" "$WHISPER_SHA" "Whisper"
download_model "$LLAMA_URL" "$LLAMA_MODEL" "$LLAMA_SHA" "Llama"

# 4. Database Readiness & Migration
if [ "${AUTO_MIGRATE:-true}" = "true" ] && [ "${DB_TYPE:-mysql}" = "mysql" ]; then
    # Use MYSQL_ vars as defined in .env.prod with safe fallbacks
    DB_H="${MYSQL_HOST:-localhost}"
    DB_P="${MYSQL_PORT:-3306}"
    DB_U="${MYSQL_USER:-root}"
    DB_PW="${MYSQL_PASSWORD:-root}"
    DB_N="${MYSQL_DATABASE:-terminal}"

    echo -e "${INFO}[INFO] Waiting for database (${DB_H}:${DB_P})...${NC}"
    
    timeout=60
    while ! (echo > /dev/tcp/"${DB_H}"/"${DB_P}") >/dev/null 2>&1; do
        sleep 1
        timeout=$((timeout - 1))
        if [ $timeout -le 0 ]; then
            echo -e "${ERROR}[ERROR] Database not ready after 60 seconds.${NC}"
            exit 1
        fi
    done
    echo -e "${SUCCESS}[SUCCESS] Database is up.${NC}"

    echo -e "${INFO}[INFO] Running migrations...${NC}"
    MIGRATIONS_PATH="/app/migrations"
    MYSQL_DSN="mysql://${DB_U}:${DB_PW}@tcp(${DB_H}:${DB_P})/${DB_N}"
    
    if migrate -path "$MIGRATIONS_PATH" -database "$MYSQL_DSN" up; then
        echo -e "${SUCCESS}[SUCCESS] Migrations complete.${NC}"
    else
        echo -e "${ERROR}[ERROR] Migration failed.${NC}"
        exit 1
    fi
fi

# 5. Final Permissions Audit
mkdir -p /app/logs

# 6. Start Application
echo -e "${SUCCESS}[SUCCESS] Initialization complete. Starting Terminal Backend...${NC}"
exec /app/app
