#!/bin/bash

# setup.sh - Teralux Backend Environment Setup Script
# Generates .env.prod with user-provided credentials and secure random keys.

set -e

# Logging colors
INFO='\033[0;34m'
SUCCESS='\033[0;32m'
NC='\033[0m' # No Color

echo -e "${INFO}[INFO] Starting Teralux environment setup...${NC}"

# 1. Prompt for Tuya Configuration
echo -e "${INFO}[INFO] Please enter your Tuya credentials:${NC}"
read -p "Tuya Client ID: " TUYA_CLIENT_ID
read -p "Tuya Access Secret: " TUYA_ACCESS_SECRET
read -p "Tuya Base URL (default: https://openapi-sg.iotbing.com): " TUYA_BASE_URL
TUYA_BASE_URL=${TUYA_BASE_URL:-https://openapi-sg.iotbing.com}
read -p "Tuya User ID: " TUYA_USER_ID

# 2. Prompt for MQTT Configuration
echo -e "${INFO}[INFO] Please enter your MQTT credentials:${NC}"
read -p "MQTT Broker (default: wss://ws.farismnrr.com:443/mqtt): " MQTT_BROKER
MQTT_BROKER=${MQTT_BROKER:-wss://ws.farismnrr.com:443/mqtt}
read -p "MQTT Username: " MQTT_USERNAME
read -s -p "MQTT Password: " MQTT_PASSWORD
echo ""
read -p "MQTT Topic (default: users/teralux/whisper): " MQTT_TOPIC
MQTT_TOPIC=${MQTT_TOPIC:-users/teralux/whisper}

# 3. Prompt for Database Configuration
echo -e "${INFO}[INFO] Please enter your Database credentials (MySQL):${NC}"
read -p "DB Host (default: teralux-db): " DB_HOST
DB_HOST=${DB_HOST:-teralux-db}
read -p "DB Port (default: 3306): " DB_PORT
DB_PORT=${DB_PORT:-3306}
read -p "DB User (default: teralux): " DB_USER
DB_USER=${DB_USER:-teralux}

# Optional Database Password Generation
DB_PASSWORD=""
read -p "Do you want to generate a random SHA-512 password for this database user? (y/n, default: n): " GEN_PASSWORD
GEN_PASSWORD=${GEN_PASSWORD:-n}

if [[ "$GEN_PASSWORD" =~ ^[Yy]$ ]]; then
    echo -e "${INFO}[INFO] Generating secure DB_PASSWORD...${NC}"
    DB_PASSWORD=$(head -c 64 /dev/urandom | sha512sum | cut -d ' ' -f 1)
else
    read -s -p "Enter DB Password: " DB_PASSWORD
    echo ""
fi

read -p "DB Name (default: teralux): " DB_NAME
DB_NAME=${DB_NAME:-teralux}

# 4. Prompt for LLM Configuration
echo -e "${INFO}[INFO] Please enter your LLM credentials:${NC}"
read -p "LLM Provider (e.g., antigravity, gemini, ollama): " LLM_PROVIDER
read -p "LLM Base URL (e.g., https://api.openai.com/v1): " LLM_BASE_URL
read -s -p "LLM API Key: " LLM_API_KEY
echo ""
read -p "LLM Model (e.g., gpt-4, gemini-1.5-pro): " LLM_MODEL

# 5. Generate Random Keys for API, JWT, and DB Root
echo -e "${INFO}[INFO] Generating secure API_KEY, JWT_SECRET, and MYSQL_ROOT_PASSWORD...${NC}"
RANDOM_API_KEY=$(head -c 64 /dev/urandom | sha512sum | cut -d ' ' -f 1)
RANDOM_JWT_SECRET=$(head -c 64 /dev/urandom | sha512sum | cut -d ' ' -f 1)
MYSQL_ROOT_PASSWORD=$(head -c 64 /dev/urandom | sha512sum | cut -d ' ' -f 1)

# 6. Create .env.prod file
ENV_FILE=".env.prod"
if [ -f "$ENV_FILE" ]; then
    echo -e "${INFO}[INFO] Existing ${ENV_FILE} found. Replacing it...${NC}"
else
    echo -e "${INFO}[INFO] Creating ${ENV_FILE}...${NC}"
fi

cat <<EOF > $ENV_FILE
# =============================================================================
# Tuya Configuration
# =============================================================================
TUYA_CLIENT_ID=$TUYA_CLIENT_ID
TUYA_ACCESS_SECRET=$TUYA_ACCESS_SECRET
TUYA_BASE_URL=$TUYA_BASE_URL
TUYA_USER_ID=$TUYA_USER_ID

# =============================================================================
# API Key Configuration
# =============================================================================
API_KEY=$RANDOM_API_KEY
JWT_SECRET=$RANDOM_JWT_SECRET

# =============================================================================
# Log Configuration
# =============================================================================
LOG_LEVEL=INFO

# =============================================================================
# Speech / RAG Configuration
# =============================================================================
LLM_PROVIDER=$LLM_PROVIDER
LLM_BASE_URL=$LLM_BASE_URL
LLM_API_KEY=$LLM_API_KEY
LLM_MODEL=$LLM_MODEL
WHISPER_MODEL_PATH=bin/ggml-base.bin
OUTSYSTEMS_TRANSCRIBE_URL=https://orion-transcribe.devoutsys.com:8443/whisper/transcribe
MAX_FILE_SIZE_MB=25
PORT=8080

# =============================================================================
# Database Configuration (Backend & Container)
# =============================================================================
DB_TYPE=mysql
DB_HOST=$DB_HOST
DB_PORT=$DB_PORT
DB_USER=$DB_USER
DB_PASSWORD=$DB_PASSWORD
DB_NAME=$DB_NAME

# Docker MySQL Container Specific
MYSQL_DATABASE=$DB_NAME
MYSQL_USER=$DB_USER
MYSQL_PASSWORD=$DB_PASSWORD
MYSQL_ROOT_PASSWORD=$MYSQL_ROOT_PASSWORD

# =============================================================================
# Responses & Cache Configuration
# =============================================================================
CACHE_TTL=1h

# =============================================================================
# MQTT Configuration
# =============================================================================
MQTT_BROKER=$MQTT_BROKER
MQTT_USERNAME=$MQTT_USERNAME
MQTT_PASSWORD=$MQTT_PASSWORD
MQTT_TOPIC=$MQTT_TOPIC
EOF

echo -e "${SUCCESS}[SUCCESS] ${ENV_FILE} has been generated successfully!${NC}"
echo -e "${INFO}[INFO] Environment setup complete.${NC}"