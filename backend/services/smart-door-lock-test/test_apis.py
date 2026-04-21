#!/usr/bin/env python3
"""
Tuya Smart Lock - Test All Password APIs
Try different API endpoints to find what works
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
env_path = Path(__file__).parent.parent / ".env"
if env_path.exists():
    with open(env_path) as f:
        for line in f:
            line = line.strip()
            if line and not line.startswith('#') and '=' in line:
                key, value = line.split('=', 1)
                os.environ[key.strip()] = value.strip()

# Configuration
CLIENT_ID = os.environ.get("TUYA_CLIENT_ID")
SECRET = os.environ.get("TUYA_ACCESS_SECRET")
BASE_URL = "https://openapi-sg.iotbing.com"
DEVICE_ID = os.environ.get("TUYA_DEVICE_ID")
USER_ID = os.environ.get("TUYA_USER_ID")

print(f"Client ID: {CLIENT_ID}")
print(f"Device ID: {DEVICE_ID}")
print(f"User ID: {USER_ID}")
print(f"Base URL: {BASE_URL}\n")

EMPTY_SHA256 = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"

def get_token():
    t = int(time.time() * 1000)
    nonce = str(uuid.uuid4())
    path = "/v1.0/token?grant_type=1"
    sign_str = f"{CLIENT_ID}{t}{nonce}GET\n{EMPTY_SHA256}\n\n{path}"
    sign = hmac.new(SECRET.encode(), sign_str.encode(), hashlib.sha256).hexdigest().upper()
    
    headers = {
        "sign": sign, "sign_method": "HMAC-SHA256", "t": str(t), 
        "nonce": nonce, "client_id": CLIENT_ID
    }
    resp = requests.get(f"{BASE_URL}{path}", headers=headers)
    result = resp.json()
    
    if result.get('success'):
        return result['result']['access_token']
    else:
        print(f"❌ Token error: {result.get('msg')}")
        return None

def test_device_info(access_token):
    """Test 1: Get device info"""
    print("\n=== Test 1: Get Device Info ===")
    t = int(time.time() * 1000)
    nonce = str(uuid.uuid4())
    path = f"/v1.0/devices/{DEVICE_ID}"
    sign_str = f"{CLIENT_ID}{access_token}{t}{nonce}GET\n{EMPTY_SHA256}\n\n{path}"
    sign = hmac.new(SECRET.encode(), sign_str.encode(), hashlib.sha256).hexdigest().upper()
    
    headers = {
        "sign": sign, "sign_method": "HMAC-SHA256", "t": str(t),
        "nonce": nonce, "client_id": CLIENT_ID, "access_token": access_token
    }
    resp = requests.get(f"{BASE_URL}{path}", headers=headers)
    result = resp.json()
    print(f"Response: {json.dumps(result, indent=2)}")
    return result.get('success')

def test_user_devices(access_token):
    """Test 2: Get user's devices"""
    print("\n=== Test 2: Get User Devices ===")
    t = int(time.time() * 1000)
    nonce = str(uuid.uuid4())
    path = f"/v1.0/users/{USER_ID}/devices"
    sign_str = f"{CLIENT_ID}{access_token}{t}{nonce}GET\n{EMPTY_SHA256}\n\n{path}"
    sign = hmac.new(SECRET.encode(), sign_str.encode(), hashlib.sha256).hexdigest().upper()
    
    headers = {
        "sign": sign, "sign_method": "HMAC-SHA256", "t": str(t),
        "nonce": nonce, "client_id": CLIENT_ID, "access_token": access_token
    }
    resp = requests.get(f"{BASE_URL}{path}", headers=headers)
    result = resp.json()
    print(f"Response: {json.dumps(result, indent=2)}")
    
    if result.get('success'):
        devices = result.get('result', [])
        target = next((d for d in devices if d.get('id') == DEVICE_ID), None)
        if target:
            print(f"\n✅ Found device: {target.get('name')}")
            print(f"   Category: {target.get('category')}")
            print(f"   Online: {target.get('online')}")
            return True
        else:
            print(f"\n❌ Device {DEVICE_ID} not found in user's devices")
    return False

def test_temp_password_v1(access_token):
    """Test 3: Create temp password (API v1)"""
    print("\n=== Test 3: Temp Password (v1 API) ===")
    t = int(time.time() * 1000)
    nonce = str(uuid.uuid4())
    path = f"/v1.0/devices/{DEVICE_ID}/door-lock/temp-password"
    
    body = {"password_type": 2, "valid_time": 5}
    body_json = json.dumps(body, separators=(',', ':'))
    body_hash = hashlib.sha256(body_json.encode()).hexdigest()
    
    sign_str = f"{CLIENT_ID}{access_token}{t}{nonce}POST\n{body_hash}\n\n{path}"
    sign = hmac.new(SECRET.encode(), sign_str.encode(), hashlib.sha256).hexdigest().upper()
    
    headers = {
        "sign": sign, "sign_method": "HMAC-SHA256", "t": str(t),
        "nonce": nonce, "client_id": CLIENT_ID, "access_token": access_token,
        "Content-Type": "application/json"
    }
    resp = requests.post(f"{BASE_URL}{path}", headers=headers, data=body_json)
    result = resp.json()
    print(f"Response: {json.dumps(result, indent=2)}")
    return result.get('success')

def test_temp_password_v2(access_token):
    """Test 4: Create temp password (API v2 with ticket)"""
    print("\n=== Test 4: Temp Password (v2 API with ticket) ===")
    
    # Step 1: Get ticket
    t = int(time.time() * 1000)
    nonce = str(uuid.uuid4())
    path = f"/v1.0/devices/{DEVICE_ID}/door-lock/password-ticket"
    sign_str = f"{CLIENT_ID}{access_token}{t}{nonce}POST\n{EMPTY_SHA256}\n\n{path}"
    sign = hmac.new(SECRET.encode(), sign_str.encode(), hashlib.sha256).hexdigest().upper()
    
    headers = {
        "sign": sign, "sign_method": "HMAC-SHA256", "t": str(t),
        "nonce": nonce, "client_id": CLIENT_ID, "access_token": access_token
    }
    resp = requests.post(f"{BASE_URL}{path}", headers=headers, json={})
    result = resp.json()
    print(f"Ticket Response: {json.dumps(result, indent=2)}")
    
    if not result.get('success'):
        print("❌ Failed to get ticket")
        return False
    
    ticket_id = result['result']['ticket_id']
    ticket_key = result['result']['ticket_key']
    print(f"Ticket ID: {ticket_id}")
    
    # Step 2: Create temp password with ticket
    t = int(time.time() * 1000)
    nonce = str(uuid.uuid4())
    path = f"/v1.0/devices/{DEVICE_ID}/door-lock/temp-password"
    
    now = int(time.time())
    body = {
        "password": "7654321",
        "password_type": "ticket",
        "ticket_id": ticket_id,
        "effective_time": now,
        "invalid_time": now + 300,  # 5 minutes
        "time_zone": "+07:00",
        "name": "test_password"
    }
    body_json = json.dumps(body, separators=(',', ':'))
    body_hash = hashlib.sha256(body_json.encode()).hexdigest()
    
    sign_str = f"{CLIENT_ID}{access_token}{t}{nonce}POST\n{body_hash}\n\n{path}"
    sign = hmac.new(SECRET.encode(), sign_str.encode(), hashlib.sha256).hexdigest().upper()
    
    headers = {
        "sign": sign, "sign_method": "HMAC-SHA256", "t": str(t),
        "nonce": nonce, "client_id": CLIENT_ID, "access_token": access_token,
        "Content-Type": "application/json"
    }
    resp = requests.post(f"{BASE_URL}{path}", headers=headers, data=body_json)
    result = resp.json()
    print(f"Password Response: {json.dumps(result, indent=2)}")
    return result.get('success')

def test_dynamic_password(access_token):
    """Test 5: Dynamic password"""
    print("\n=== Test 5: Dynamic Password ===")
    t = int(time.time() * 1000)
    nonce = str(uuid.uuid4())
    path = f"/v1.0/devices/{DEVICE_ID}/door-lock/dynamic-password"
    sign_str = f"{CLIENT_ID}{access_token}{t}{nonce}GET\n{EMPTY_SHA256}\n\n{path}"
    sign = hmac.new(SECRET.encode(), sign_str.encode(), hashlib.sha256).hexdigest().upper()
    
    headers = {
        "sign": sign, "sign_method": "HMAC-SHA256", "t": str(t),
        "nonce": nonce, "client_id": CLIENT_ID, "access_token": access_token
    }
    resp = requests.get(f"{BASE_URL}{path}", headers=headers)
    result = resp.json()
    print(f"Response: {json.dumps(result, indent=2)}")
    return result.get('success')

def main():
    print("=" * 60)
    print("TUYA SMART LOCK - API TEST")
    print("=" * 60)
    
    access_token = get_token()
    if not access_token:
        print("❌ Failed to get access token")
        return
    
    print(f"\n✅ Access Token: {access_token[:20]}...")
    
    # Run tests
    results = {}
    results['device_info'] = test_device_info(access_token)
    results['user_devices'] = test_user_devices(access_token)
    results['temp_password_v1'] = test_temp_password_v1(access_token)
    results['temp_password_v2'] = test_temp_password_v2(access_token)
    results['dynamic_password'] = test_dynamic_password(access_token)
    
    print("\n" + "=" * 60)
    print("SUMMARY")
    print("=" * 60)
    for test, success in results.items():
        status = "✅" if success else "❌"
        print(f"{status} {test}: {'PASS' if success else 'FAIL'}")

if __name__ == "__main__":
    main()
