#!/usr/bin/env python3
"""
Tuya Smart Lock - Create Temporary Password (Simple API - No Ticket)
Generate temporary passwords for Tuya smart door locks using direct API
Loads credentials from .env file
"""
import requests
import hmac
import hashlib
import time
import uuid
import json
import os
from pathlib import Path

# Load .env file
env_path = Path(__file__).parent / ".env"
if env_path.exists():
    with open(env_path) as f:
        for line in f:
            line = line.strip()
            if line and not line.startswith('#') and '=' in line:
                key, value = line.split('=', 1)
                os.environ[key.strip()] = value.strip()

# ============= CONFIGURATION =============
CLIENT_ID = os.environ.get("TUYA_CLIENT_ID")
SECRET = os.environ.get("TUYA_ACCESS_SECRET")
BASE_URL = "https://openapi-sg.iotbing.com"
DEVICE_ID = os.environ.get("TUYA_DEVICE_ID")

if not CLIENT_ID or not SECRET or not DEVICE_ID:
    print("❌ Missing required environment variables:")
    print("   TUYA_CLIENT_ID, TUYA_ACCESS_SECRET, TUYA_DEVICE_ID")
    exit(1)

# SHA256 constants
EMPTY_BODY_SHA256 = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

def get_access_token():
    """Step 1: Get Access Token"""
    t = int(time.time() * 1000)
    nonce = str(uuid.uuid4())
    path = "/v1.0/token?grant_type=1"
    string_to_sign = f"GET\n{EMPTY_BODY_SHA256}\n\n{path}"
    sign_str = f"{CLIENT_ID}{t}{nonce}{string_to_sign}"
    sign = hmac.new(SECRET.encode(), sign_str.encode(), hashlib.sha256).hexdigest().upper()

    headers = {"sign": sign, "sign_method": "HMAC-SHA256", "t": str(t), "nonce": nonce, "client_id": CLIENT_ID}
    response = requests.get(f"{BASE_URL}{path}", headers=headers)
    result = response.json()

    if result.get('success'):
        return result['result']['access_token']
    else:
        raise Exception(f"Failed to get access token: {result.get('msg')}")

def create_temp_password_simple(access_token, password, valid_minutes=5):
    """
    Create Temporary Password using simple API (password_type=2)
    This is the direct API without ticket encryption
    """
    t = int(time.time() * 1000)
    nonce = str(uuid.uuid4())
    path = f"/v1.0/devices/{DEVICE_ID}/door-lock/temp-password"

    body = {
        "password_type": 2,  # 2 = temporary password
        "valid_time": valid_minutes,
        "password": password  # Optional: custom password
    }

    body_json = json.dumps(body, separators=(',', ':'))
    body_sha256 = hashlib.sha256(body_json.encode()).hexdigest()

    string_to_sign = f"POST\n{body_sha256}\n\n{path}"
    sign_str = f"{CLIENT_ID}{access_token}{t}{nonce}{string_to_sign}"
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

    print(f"Request URL: {BASE_URL}{path}")
    print(f"Request Body: {json.dumps(body, indent=2)}")
    
    response = requests.post(f"{BASE_URL}{path}", headers=headers, data=body_json)
    result = response.json()

    return result, body

def generate_password(password="7654321", valid_minutes=5):
    """Main function to generate a temporary password"""
    print("==========================================")
    print(f"  TUYA SMART LOCK - CREATE PASSWORD (Simple API)")
    print(f"  Password: {password}")
    print(f"  Duration: {valid_minutes} minutes")
    print("==========================================\n")

    # Step 1: Get Access Token
    print("=== Step 1: Get Access Token ===")
    access_token = get_access_token()
    print(f"✅ Access Token: {access_token}\n")

    # Step 2: Create Temporary Password
    print("=== Step 2: Create Temporary Password ===")
    result, body = create_temp_password_simple(access_token, password, valid_minutes)

    print(f"\nResponse: {json.dumps(result, indent=2)}")

    if result.get('success'):
        result_data = result.get('result', {})
        generated_password = result_data.get('password', 'N/A')
        expire_time = result_data.get('expire_time', 0)
        expire_str = time.strftime('%Y-%m-%d %H:%M:%S', time.localtime(expire_time / 1000)) if expire_time else 'N/A'
        
        print(f"\n{'='*50}")
        print(f"✅ SUCCESS! Temporary Password Created!")
        print(f"{'='*50}")
        print(f"   Password: {generated_password}")
        print(f"   Valid Until: {expire_str}")
        print(f"   Duration: {valid_minutes} minutes")
        print(f"   Device: {DEVICE_ID}")
        print(f"{'='*50}\n")
        return True, generated_password
    else:
        print(f"\n❌ FAILED: {result.get('msg')} (code: {result.get('code')})")
        return False, None

if __name__ == "__main__":
    import sys

    # Default password or from command line
    password = sys.argv[1] if len(sys.argv) > 1 else "7654321"
    valid_minutes = int(sys.argv[2]) if len(sys.argv) > 2 else 5

    success, generated_password = generate_password(password, valid_minutes)
    exit(0 if success else 1)
