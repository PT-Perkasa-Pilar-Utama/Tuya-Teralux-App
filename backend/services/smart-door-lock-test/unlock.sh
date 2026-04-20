#!/bin/bash
# Smart Door Lock - Reply Unlock Request
# Usage: ./unlock.sh true   (or)   ./unlock.sh false

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="$SCRIPT_DIR/.env"

# Check argument
if [ -z "$1" ]; then
    echo "Usage: $0 <true|false>"
    echo "Example: $0 true"
    exit 1
fi

VALUE="$1"

# Validate value
if [ "$VALUE" != "true" ] && [ "$VALUE" != "false" ]; then
    echo "❌ Error: Value must be 'true' or 'false'"
    echo "Usage: $0 <true|false>"
    exit 1
fi

# Check .env file
if [ ! -f "$ENV_FILE" ]; then
    echo "❌ Error: .env file not found at $ENV_FILE"
    exit 1
fi

# Load environment variables
export $(grep -v '^#' "$ENV_FILE" | xargs)

# Check required vars
if [ -z "$TUYA_CLIENT_ID" ] || [ -z "$TUYA_ACCESS_SECRET" ] || [ -z "$TUYA_DEVICE_ID" ]; then
    echo "❌ Error: Missing required credentials in .env"
    echo "Required: TUYA_CLIENT_ID, TUYA_ACCESS_SECRET, TUYA_DEVICE_ID"
    exit 1
fi

echo "=============================================================="
echo "  SMART DOOR LOCK - REPLY UNLOCK REQUEST"
echo "=============================================================="
echo ""
echo "  Device: $TUYA_DEVICE_ID"
echo "  Value:  $VALUE"
echo ""

# Use Python for proper signature handling
python3 - "$TUYA_CLIENT_ID" "$TUYA_ACCESS_SECRET" "$TUYA_DEVICE_ID" "$TUYA_BASE_URL" "$VALUE" <<'PYTHON_SCRIPT'
import sys
import requests
import hmac
import hashlib
import time
import uuid
import json

CLIENT_ID = sys.argv[1]
SECRET = sys.argv[2]
DEVICE_ID = sys.argv[3]
BASE_URL = sys.argv[4] if len(sys.argv) > 4 and sys.argv[4] else "https://openapi-sg.iotbing.com"
VALUE = sys.argv[5].lower() == 'true'

EMPTY_SHA256 = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

def get_token():
    t = int(time.time() * 1000)
    nonce = str(uuid.uuid4())
    path = "/v1.0/token?grant_type=1"
    sign_str = f"{CLIENT_ID}{t}{nonce}GET\n{EMPTY_SHA256}\n\n{path}"
    sign = hmac.new(SECRET.encode(), sign_str.encode(), hashlib.sha256).hexdigest().upper()
    
    headers = {
        "sign": sign,
        "sign_method": "HMAC-SHA256",
        "t": str(t),
        "nonce": nonce,
        "client_id": CLIENT_ID
    }
    
    resp = requests.get(f"{BASE_URL}{path}", headers=headers)
    result = resp.json()
    
    if result.get("success"):
        return result["result"]["access_token"]
    else:
        print(f"❌ Failed to get token: {result.get('msg')} (code: {result.get('code')})")
        return None

def send_command(access_token, cmd_value):
    t = int(time.time() * 1000)
    nonce = str(uuid.uuid4())
    path = f"/v1.0/devices/{DEVICE_ID}/commands"
    
    body = {"commands": [{"code": "reply_unlock_request", "value": cmd_value}]}
    body_json = json.dumps(body, separators=(',', ':'))
    body_hash = hashlib.sha256(body_json.encode()).hexdigest()
    
    sign_str = f"{CLIENT_ID}{access_token}{t}{nonce}POST\n{body_hash}\n\n{path}"
    sign = hmac.new(SECRET.encode(), sign_str.encode(), hashlib.sha256).hexdigest().upper()
    
    headers = {
        "sign": sign,
        "sign_method": "HMAC-SHA256",
        "t": str(t),
        "nonce": nonce,
        "client_id": CLIENT_ID,
        "access_token": access_token,
        "Content-Type": "application/json"
    }
    
    resp = requests.post(f"{BASE_URL}{path}", headers=headers, data=body_json)
    return resp.json()

# Main
print("=== Step 1: Get Access Token ===")
access_token = get_token()

if not access_token:
    sys.exit(1)

print(f"✅ Access Token: {access_token[:30]}...")
print("")

# Step 2: Send Command
print("=== Step 2: Send Command ===")
result = send_command(access_token, VALUE)

print(f"Response: {json.dumps(result, indent=2)}")
print("")

print("==============================================================")
print("  RESULT")
print("==============================================================")

if result.get("success"):
    print("✅ Command sent successfully!")
    print("")
    print(f"   Reply Unlock Request: {str(VALUE).lower()}")
    print("")
    print("💡 Note: This command acknowledges an unlock request.")
    print("   It may not actually unlock the door depending on device.")
    sys.exit(0)
else:
    print("❌ Command failed!")
    print("")
    print(f"   Code: {result.get('code', 'N/A')}")
    print(f"   Message: {result.get('msg', 'N/A')}")
    
    code = result.get("code")
    if code == 2008:
        print("")
        print("💡 Error 2008: Command or value not supported")
        print("   This device may not support reply_unlock_request")
    elif code == 1106:
        print("")
        print("💡 Error 1106: Permission denied")
        print("   Check if device is linked to your Tuya cloud project")
    elif code == 2001:
        print("")
        print("💡 Error 2001: Device is offline")
        print("   Check WiFi connection")
    
    sys.exit(1)
PYTHON_SCRIPT
