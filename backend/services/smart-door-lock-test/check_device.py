#!/usr/bin/env python3
"""
Tuya Smart Lock - Check Device Status
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
    """Get Access Token"""
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

def get_device_info(access_token):
    """Get Device Info"""
    t = int(time.time() * 1000)
    nonce = str(uuid.uuid4())
    path = f"/v1.0/devices/{DEVICE_ID}"
    string_to_sign = f"GET\n{EMPTY_BODY_SHA256}\n\n{path}"
    sign_str = f"{CLIENT_ID}{access_token}{t}{nonce}{string_to_sign}"
    sign = hmac.new(SECRET.encode(), sign_str.encode(), hashlib.sha256).hexdigest().upper()

    headers = {"sign": sign, "sign_method": "HMAC-SHA256", "t": str(t), "nonce": nonce, "client_id": CLIENT_ID, "access_token": access_token}
    response = requests.get(f"{BASE_URL}{path}", headers=headers)
    return response.json()

def get_device_specifications(access_token):
    """Get Device Specifications/Functions"""
    t = int(time.time() * 1000)
    nonce = str(uuid.uuid4())
    path = f"/v1.0/devices/{DEVICE_ID}/specifications"
    string_to_sign = f"GET\n{EMPTY_BODY_SHA256}\n\n{path}"
    sign_str = f"{CLIENT_ID}{access_token}{t}{nonce}{string_to_sign}"
    sign = hmac.new(SECRET.encode(), sign_str.encode(), hashlib.sha256).hexdigest().upper()

    headers = {"sign": sign, "sign_method": "HMAC-SHA256", "t": str(t), "nonce": nonce, "client_id": CLIENT_ID, "access_token": access_token}
    response = requests.get(f"{BASE_URL}{path}", headers=headers)
    return response.json()

def main():
    print("==========================================")
    print(f"  TUYA SMART LOCK - DEVICE INFO")
    print(f"  Device ID: {DEVICE_ID}")
    print("==========================================\n")

    # Step 1: Get Access Token
    print("=== Step 1: Get Access Token ===")
    access_token = get_access_token()
    print(f"✅ Access Token: {access_token}\n")

    # Step 2: Get Device Info
    print("=== Step 2: Get Device Info ===")
    device_result = get_device_info(access_token)
    print(f"Response: {json.dumps(device_result, indent=2)}\n")

    if device_result.get('success'):
        result = device_result.get('result', {})
        print(f"Device Name: {result.get('name', 'N/A')}")
        print(f"Device Category: {result.get('category', 'N/A')}")
        print(f"Online: {result.get('online', False)}")
        print(f"Sub Device: {result.get('sub', False)}")
        print(f"UUID: {result.get('uuid', 'N/A')}")
        print()

    # Step 3: Get Device Specifications
    print("=== Step 3: Get Device Specifications ===")
    specs_result = get_device_info(access_token)
    print(f"Response: {json.dumps(specs_result, indent=2)}\n")

    if specs_result.get('success'):
        result = specs_result.get('result', {})
        functions = result.get('functions', [])
        statuses = result.get('status', [])
        
        print(f"Available Functions ({len(functions)}):")
        for fn in functions:
            print(f"  - {fn.get('code')}: {fn.get('type')} ({fn.get('name', '')})")
        
        print(f"\nCurrent Statuses ({len(statuses)}):")
        for status in statuses:
            print(f"  - {status.get('code')}: {status.get('value')}")

if __name__ == "__main__":
    main()
