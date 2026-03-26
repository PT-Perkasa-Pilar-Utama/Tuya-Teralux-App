#!/bin/bash
set -euo pipefail

# Logging colors
INFO='\033[0;34m'
SUCCESS='\033[0;32m'
WARNING='\033[0;33m'
ERROR='\033[0;31m'
NC='\033[0m'

echo -e "${INFO}[INFO] Starting Sensio Backend (10/10 Production Grade)...${NC}"

# 1. Environment & Path Setup
mkdir -p /app/models

# 2. Binary Verification
echo -e "${INFO}[INFO] Verifying binaries...${NC}"
for bin in "/app/app" "/usr/local/bin/migrate" "/usr/bin/chromium" "/usr/bin/ffmpeg" "/usr/bin/ffprobe"; do
    if [ ! -f "$bin" ]; then
        echo -e "${ERROR}[ERROR] Required binary missing: ${bin}${NC}"
        exit 1
    fi
done
echo -e "${SUCCESS}[SUCCESS] Binaries verified.${NC}"

# 3. Database Readiness & Migration
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
echo -e "${INFO}[INFO] Prewarming Chromium...${NC}"
if /usr/bin/chromium --headless=new --no-sandbox --disable-dev-shm-usage --dump-dom about:blank >/dev/null 2>&1; then
    echo -e "${SUCCESS}[SUCCESS] Chromium prewarmed.${NC}"
else
    echo -e "${ERROR}[ERROR] Chromium prewarm failed. PDF generation may not work.${NC}"
    exit 1
fi

# 6. Final Permissions Audit
mkdir -p /app/logs
mkdir -p /app/uploads/reports

# 7. Start Application
echo -e "${SUCCESS}[SUCCESS] Initialization complete. Starting Sensio Backend...${NC}"
exec /app/app
