#!/bin/bash
#
# Tuya Smart Lock - Set Password Script
# Usage: ./setPassword.sh <password> [valid_hours]
# Example: ./setPassword.sh 7654321 24
#

set -e

# ============= CONFIGURATION =============
CLIENT_ID="nnwar5dvq7fsdqpdtkjf"
SECRET="051708689fc7401f84aaee88bfce9dda"
BASE_URL="https://openapi-sg.iotbing.com"
DEVICE_ID="a3621a5ad61e644d91aaa2"

# SHA256 constants
EMPTY_BODY_SHA256="e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
EMPTY_JSON_SHA256="44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a"
# =========================================

# Check arguments
if [ -z "$1" ]; then
    echo "Usage: $0 <password> [valid_hours]"
    echo "Example: $0 7654321 24"
    exit 1
fi

PASSWORD="$1"
VALID_HOURS="${2:-1}"

echo "=========================================="
echo "  TUYA SMART LOCK - SET PASSWORD"
echo "  Password: $PASSWORD"
echo "  Valid: $VALID_HOURS hour(s)"
echo "=========================================="
echo ""

# Function to generate HMAC-SHA256 signature
generate_sign() {
    local sign_str="$1"
    echo -n "$sign_str" | openssl dgst -sha256 -hmac "$SECRET" | awk '{print toupper($2)}'
}

# Step 1: Get Access Token
echo "=== Step 1: Get Access Token ==="
T=$(date +%s%3N)
NONCE=$(cat /proc/sys/kernel/random/uuid)
TOKEN_PATH="/v1.0/token?grant_type=1"

STRING_TO_SIGN="GET
${EMPTY_BODY_SHA256}

${TOKEN_PATH}"

SIGN=$(generate_sign "${CLIENT_ID}${T}${NONCE}${STRING_TO_SIGN}")

TOKEN_RESPONSE=$(curl -s -X GET "${BASE_URL}${TOKEN_PATH}" \
  -H "sign: $SIGN" \
  -H "sign_method: HMAC-SHA256" \
  -H "t: $T" \
  -H "nonce: $NONCE" \
  -H "client_id: $CLIENT_ID")

ACCESS_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.result.access_token')

if [ -z "$ACCESS_TOKEN" ] || [ "$ACCESS_TOKEN" = "null" ]; then
    echo "❌ Failed to get access token!"
    echo "Response: $TOKEN_RESPONSE"
    exit 1
fi

echo "✅ Access Token: ${ACCESS_TOKEN:0:20}..."
echo ""

# Step 2: Get Password Ticket
echo "=== Step 2: Get Password Ticket ==="
T=$(date +%s%3N)
NONCE=$(cat /proc/sys/kernel/random/uuid)
TICKET_PATH="/v1.0/devices/${DEVICE_ID}/door-lock/password-ticket"

STRING_TO_SIGN="POST
${EMPTY_JSON_SHA256}

${TICKET_PATH}"

SIGN=$(generate_sign "${CLIENT_ID}${ACCESS_TOKEN}${T}${NONCE}${STRING_TO_SIGN}")

TICKET_RESPONSE=$(curl -s -X POST "${BASE_URL}${TICKET_PATH}" \
  -H "sign: $SIGN" \
  -H "sign_method: HMAC-SHA256" \
  -H "t: $T" \
  -H "nonce: $NONCE" \
  -H "client_id: $CLIENT_ID" \
  -H "access_token: $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}')

TICKET_ID=$(echo "$TICKET_RESPONSE" | jq -r '.result.ticket_id')
TICKET_KEY=$(echo "$TICKET_RESPONSE" | jq -r '.result.ticket_key')

if [ -z "$TICKET_ID" ] || [ "$TICKET_ID" = "null" ]; then
    echo "❌ Failed to get password ticket!"
    echo "Response: $TICKET_RESPONSE"
    exit 1
fi

echo "✅ Ticket ID: $TICKET_ID"
echo "✅ Ticket Key: $TICKET_KEY"
echo ""

# Step 3: Decrypt Ticket Key (AES-256-ECB)
echo "=== Step 3: Decrypt Ticket Key (AES-256-ECB) ==="

# Convert hex ticket_key to binary, decrypt with AES-256-ECB
DECRYPTED_KEY=$(echo -n "$TICKET_KEY" | xxd -r -p | \
  openssl enc -aes-256-ecb -d -K $(printf '%s' "$SECRET" | xxd -p | tr -d '\n' | head -c 64) 2>/dev/null | \
  xxd -p | tr -d '\n' | tr '[:lower:]' '[:upper:]')

# Remove PKCS7 padding (last byte indicates padding length)
PADDING_LEN=$((16#${DECRYPTED_KEY: -2}))
if [ "$PADDING_LEN" -le 16 ] && [ "$PADDING_LEN" -gt 0 ]; then
    DECRYPTED_KEY="${DECRYPTED_KEY:0:$((64 - PADDING_LEN * 2))}"
fi

echo "✅ Decrypted Key: $DECRYPTED_KEY"
echo ""

# Step 4: Encrypt Password (AES-128-ECB + PKCS7Padding)
echo "=== Step 4: Encrypt Password (AES-128-ECB) ==="
echo "Plain Password: $PASSWORD"

# Use first 32 hex chars (16 bytes) for AES-128 key
AES128_KEY="${DECRYPTED_KEY:0:32}"

# Encrypt password with PKCS7 padding
ENCRYPTED_PASSWORD=$(python3 -c "
from Crypto.Cipher import AES
from Crypto.Util.Padding import pad
import binascii

key = binascii.unhexlify('$AES128_KEY')
password = '$PASSWORD'.encode('utf-8')
padded = pad(password, AES.block_size)
cipher = AES.new(key, AES.MODE_ECB)
encrypted = cipher.encrypt(padded)
print(binascii.hexlify(encrypted).decode().upper())
")

echo "✅ Encrypted Password: $ENCRYPTED_PASSWORD"
echo ""

# Step 5: Create Temporary Password
echo "=== Step 5: Create Temporary Password ==="
T=$(date +%s%3N)
NONCE=$(cat /proc/sys/kernel/random/uuid)
TEMP_PATH="/v1.0/devices/${DEVICE_ID}/door-lock/temp-password"

NOW_SEC=$(date +%s)
INVALID_TIME=$((NOW_SEC + VALID_HOURS * 3600))

# Create JSON body
BODY=$(cat <<EOF
{"password":"${ENCRYPTED_PASSWORD}","password_type":"ticket","ticket_id":"${TICKET_ID}","effective_time":${NOW_SEC},"invalid_time":${INVALID_TIME},"time_zone":"+07:00","name":"password_${PASSWORD}"}
EOF
)

BODY_SHA256=$(echo -n "$BODY" | openssl dgst -sha256 | awk '{print $2}')

STRING_TO_SIGN="POST
${BODY_SHA256}

${TEMP_PATH}"

SIGN=$(generate_sign "${CLIENT_ID}${ACCESS_TOKEN}${T}${NONCE}${STRING_TO_SIGN}")

echo "Sending request..."

TEMP_RESPONSE=$(curl -s -X POST "${BASE_URL}${TEMP_PATH}" \
  -H "sign: $SIGN" \
  -H "sign_method: HMAC-SHA256" \
  -H "t: $T" \
  -H "nonce: $NONCE" \
  -H "client_id: $CLIENT_ID" \
  -H "access_token: $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d "$BODY")

echo "Response: $TEMP_RESPONSE"
echo ""

# Check result
SUCCESS=$(echo "$TEMP_RESPONSE" | jq -r '.success')
CODE=$(echo "$TEMP_RESPONSE" | jq -r '.code')

if [ "$SUCCESS" = "true" ]; then
    PASSWORD_ID=$(echo "$TEMP_RESPONSE" | jq -r '.result.id')
    echo "=========================================="
    echo "✅ SUCCESS! Password Created!"
    echo "=========================================="
    echo "   Password: $PASSWORD"
    echo "   Password ID: $PASSWORD_ID"
    echo "   Valid: $VALID_HOURS hour(s)"
    echo "   Device: $DEVICE_ID"
    echo "=========================================="
    exit 0
else
    MSG=$(echo "$TEMP_RESPONSE" | jq -r '.msg')
    echo "=========================================="
    echo "❌ FAILED!"
    echo "=========================================="
    echo "   Error: $MSG"
    echo "   Code: $CODE"
    echo "=========================================="
    exit 1
fi
