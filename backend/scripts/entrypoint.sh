#!/bin/bash
set -euo pipefail

# Logging colors
INFO='\033[0;34m'
SUCCESS='\033[0;32m'
WARNING='\033[0;33m'
ERROR='\033[0;31m'
NC='\033[0m'

# Setup library path for llama shared libraries
export LD_LIBRARY_PATH=/app/lib/llama:${LD_LIBRARY_PATH:-}

echo -e "${INFO}[INFO] Starting Sensio Backend (10/10 Production Grade)...${NC}"

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

# 2. Chromium Auto-Installation
# Function to check if Chromium exists at any known path
find_chromium() {
    local paths=(
        "/usr/bin/chromium"
        "/usr/bin/chromium-browser"
        "/usr/bin/google-chrome"
        "/usr/bin/google-chrome-stable"
        "/data/data/com.termux/files/usr/bin/chromium"
        "/data/data/com.termux/files/usr/bin/chromium-browser"
    )
    for path in "${paths[@]}"; do
        if [ -f "$path" ] && [ -x "$path" ]; then
            echo "$path"
            return 0
        fi
    done
    return 1
}

# Function to auto-install Chromium
install_chromium() {
    echo -e "${WARNING}[WARNING] Chromium not found. Attempting auto-installation...${NC}"
    
    # Detect package manager and OS
    if command -v apt-get &>/dev/null; then
        echo -e "${INFO}[INFO] Using apt-get for installation...${NC}"
        if apt-get update -qq && apt-get install -y -qq --no-install-recommends \
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
            2>/dev/null; then
            echo -e "${SUCCESS}[SUCCESS] Chromium installed via apt-get.${NC}"
            return 0
        fi
    fi
    
    if command -v pkg &>/dev/null; then
        # Termux Android
        echo -e "${INFO}[INFO] Detected Termux. Using pkg for installation...${NC}"
        if pkg install -y chromium 2>/dev/null; then
            echo -e "${SUCCESS}[SUCCESS] Chromium installed via pkg (Termux).${NC}"
            return 0
        fi
    fi
    
    if command -v apt &>/dev/null; then
        echo -e "${INFO}[INFO] Using apt for installation...${NC}"
        if apt update -qq && apt install -y -qq --no-install-recommends chromium 2>/dev/null; then
            echo -e "${SUCCESS}[SUCCESS] Chromium installed via apt.${NC}"
            return 0
        fi
    fi
    
    echo -e "${ERROR}[ERROR] Failed to install Chromium. No supported package manager found or installation failed.${NC}"
    echo -e "${WARNING}[WARNING] PDF generation will be disabled. The app will continue but summary stage may fail.${NC}"
    return 1
}

# Check and auto-install Chromium if needed
CHROMIUM_PATH=""
if chromium_path=$(find_chromium); then
    CHROMIUM_PATH="$chromium_path"
    echo -e "${SUCCESS}[SUCCESS] Chromium found at: ${CHROMIUM_PATH}${NC}"
else
    echo -e "${WARNING}[WARNING] Chromium not found in standard paths.${NC}"
    if install_chromium; then
        # Try to find it again after installation
        if chromium_path=$(find_chromium); then
            CHROMIUM_PATH="$chromium_path"
            echo -e "${SUCCESS}[SUCCESS] Chromium installed and found at: ${CHROMIUM_PATH}${NC}"
        else
            echo -e "${ERROR}[ERROR] Chromium installation succeeded but binary not found in expected paths.${NC}"
            CHROMIUM_PATH="/usr/bin/chromium"  # Default fallback
        fi
    else
        echo -e "${WARNING}[WARNING] Continuing without Chromium. PDF generation will be disabled.${NC}"
        CHROMIUM_PATH=""
    fi
fi

# Export Chromium path for the application
export CHROMIUM_PATH="${CHROMIUM_PATH:-/usr/bin/chromium}"
echo -e "${INFO}[INFO] CHROMIUM_PATH set to: ${CHROMIUM_PATH}${NC}"

# 2. Binary Verification (excluding Chromium - handled separately)
echo -e "${INFO}[INFO] Verifying binaries...${NC}"
for bin in "$WHISPER_CLI" "$LLAMA_CLI" "/app/app" "/usr/local/bin/migrate" "/usr/bin/ffmpeg" "/usr/bin/ffprobe"; do
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
        if [ "$(find "$LOCK_DIR" -maxdepth 0 -mmin +2 2>/dev/null)" ]; then
            echo -e "${WARNING}[WARNING] Stale lock detected (older than 2m). Cleaning up...${NC}"
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
    DB_N="${MYSQL_DATABASE:-sensio}"

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

# 5. Browser Prewarm
if [ -n "$CHROMIUM_PATH" ] && [ -f "$CHROMIUM_PATH" ]; then
    echo -e "${INFO}[INFO] Prewarming Chromium at ${CHROMIUM_PATH}...${NC}"
    if "$CHROMIUM_PATH" --headless=new --no-sandbox --disable-dev-shm-usage --dump-dom about:blank >/dev/null 2>&1; then
        echo -e "${SUCCESS}[SUCCESS] Chromium prewarmed.${NC}"
    else
        echo -e "${ERROR}[ERROR] Chromium prewarm failed. PDF generation may not work.${NC}"
        # Don't exit - allow app to continue without PDF generation
    fi
else
    echo -e "${WARNING}[WARNING] Chromium not available. Skipping prewarm. PDF generation will be disabled.${NC}"
fi

# 6. Final Permissions Audit
mkdir -p /app/logs
mkdir -p /app/uploads/reports

# 7. Start Application
echo -e "${SUCCESS}[SUCCESS] Initialization complete. Starting Sensio Backend...${NC}"
exec /app/app
