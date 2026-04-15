#!/usr/bin/env python3
"""
Tuya Smart Lock - Cleanup All Temporary Passwords
Usage: python3 cleanup_passwords.py

Credentials must be set via environment variables or .env file:
  TUYA_CLIENT_ID, TUYA_SECRET, TUYA_DEVICE_ID
"""
import requests
import hmac
import hashlib
import time
import uuid
import json
import os

# ============= CONFIGURATION =============
CLIENT_ID = os.environ.get("TUYA_CLIENT_ID")
SECRET = os.environ.get("TUYA_SECRET")
BASE_URL = "https://openapi-sg.iotbing.com"
DEVICE_ID = os.environ.get("TUYA_DEVICE_ID")

if not CLIENT_ID or not SECRET or not DEVICE_ID:
    print("❌ Missing required environment variables:")
    print("   TUYA_CLIENT_ID, TUYA_SECRET, TUYA_DEVICE_ID")
    print("\nSet them via: export TUYA_CLIENT_ID=xxx")
    print("Or create a .env file (see .env.example)")
    exit(1)

EMPTY_BODY_SHA256 = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
# =========================================

def get_access_token():
    t = int(time.time() * 1000)
    nonce = str(uuid.uuid4())
    path = "/v1.0/token?grant_type=1"
    string_to_sign = f"GET\n{EMPTY_BODY_SHA256}\n\n{path}"
    sign_str = f"{CLIENT_ID}{t}{nonce}{string_to_sign}"
    sign = hmac.new(SECRET.encode(), sign_str.encode(), hashlib.sha256).hexdigest().upper()
    
    headers = {"sign": sign, "sign_method": "HMAC-SHA256", "t": str(t), "nonce": nonce, "client_id": CLIENT_ID}
    response = requests.get(f"{BASE_URL}{path}", headers=headers)
    return response.json()['result']['access_token']

def get_temp_passwords(access_token):
    t = int(time.time() * 1000)
    nonce = str(uuid.uuid4())
    path = f"/v1.0/devices/{DEVICE_ID}/door-lock/temp-passwords"
    string_to_sign = f"GET\n{EMPTY_BODY_SHA256}\n\n{path}"
    sign_str = f"{CLIENT_ID}{access_token}{t}{nonce}{string_to_sign}"
    sign = hmac.new(SECRET.encode(), sign_str.encode(), hashlib.sha256).hexdigest().upper()
    
    headers = {"sign": sign, "sign_method": "HMAC-SHA256", "t": str(t), "nonce": nonce,
               "client_id": CLIENT_ID, "access_token": access_token}
    
    response = requests.get(f"{BASE_URL}{path}", headers=headers)
    return response.json()

def delete_temp_password(access_token, password_id):
    t = int(time.time() * 1000)
    nonce = str(uuid.uuid4())
    path = f"/v1.0/devices/{DEVICE_ID}/door-lock/temp-passwords/{password_id}"
    string_to_sign = f"DELETE\n{EMPTY_BODY_SHA256}\n\n{path}"
    sign_str = f"{CLIENT_ID}{access_token}{t}{nonce}{string_to_sign}"
    sign = hmac.new(SECRET.encode(), sign_str.encode(), hashlib.sha256).hexdigest().upper()
    
    headers = {"sign": sign, "sign_method": "HMAC-SHA256", "t": str(t), "nonce": nonce,
               "client_id": CLIENT_ID, "access_token": access_token}
    
    response = requests.request("DELETE", f"{BASE_URL}{path}", headers=headers)
    return response.json()

if __name__ == "__main__":
    print("==========================================")
    print("  TUYA SMART LOCK - CLEANUP PASSWORDS")
    print("==========================================\n")
    
    access_token = get_access_token()
    print(f"✅ Access Token: {access_token[:20]}...\n")
    
    print("=== Getting Temporary Passwords List ===")
    result = get_temp_passwords(access_token)
    
    if not result.get('success'):
        print(f"❌ Failed: {result.get('msg')}")
        exit(1)
    
    passwords = result.get('result', [])
    if isinstance(passwords, dict):
        passwords = passwords.get('list', [])
    
    print(f"Found {len(passwords)} password(s)\n")
    
    if not passwords:
        print("✅ No passwords to delete!")
        exit(0)
    
    # Print passwords
    print("Current passwords:")
    for i, pwd in enumerate(passwords, 1):
        pwd_id = pwd.get('id', 'N/A')
        name = pwd.get('name', 'N/A')
        valid = pwd.get('valid', False)
        status = "✅ Valid" if valid else "❌ Expired"
        print(f"  {i}. {name} (ID: {pwd_id}) - {status}")
    
    print()
    
    # Confirm deletion
    confirm = input(f"Delete all {len(passwords)} password(s)? (yes/no): ").strip().lower()
    
    if confirm != 'yes':
        print("❌ Cancelled by user")
        exit(0)
    
    print("\n=== Deleting All Passwords ===\n")
    
    deleted = 0
    failed = 0
    
    for pwd in passwords:
        pwd_id = pwd.get('id')
        name = pwd.get('name', 'N/A')
        print(f"Deleting {pwd_id} ({name})...", end=" ")
        
        result = delete_temp_password(access_token, pwd_id)
        
        if result.get('success'):
            print("✅")
            deleted += 1
        else:
            print(f"❌ {result.get('msg')}")
            failed += 1
    
    print("\n==========================================")
    print(f"   DELETED: {deleted} ✅")
    print(f"   FAILED: {failed} ❌")
    print("==========================================\n")
