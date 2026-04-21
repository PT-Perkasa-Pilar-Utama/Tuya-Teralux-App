#!/usr/bin/env python3
"""
Tuya Smart Lock - Create Temporary Password (5 minutes)
Generate encrypted temporary passwords for Tuya smart door locks
Loads credentials from .env file
"""
import requests
import hmac
import hashlib
import time
import uuid
from Crypto.Cipher import AES
from Crypto.Util.Padding import pad
import binascii
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
EMPTY_JSON_SHA256 = "44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a"

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

def get_password_ticket(access_token):
    """Step 2: Get Password Ticket"""
    t = int(time.time() * 1000)
    nonce = str(uuid.uuid4())
    path = f"/v1.0/devices/{DEVICE_ID}/door-lock/password-ticket"
    string_to_sign = f"POST\n{EMPTY_JSON_SHA256}\n\n{path}"
    sign_str = f"{CLIENT_ID}{access_token}{t}{nonce}{string_to_sign}"
    sign = hmac.new(SECRET.encode(), sign_str.encode(), hashlib.sha256).hexdigest().upper()

    headers = {"sign": sign, "sign_method": "HMAC-SHA256", "t": str(t), "nonce": nonce,
               "client_id": CLIENT_ID, "access_token": access_token}
    response = requests.post(f"{BASE_URL}{path}", headers=headers, json={})
    result = response.json()

    if result.get('success'):
        return result['result']['ticket_id'], result['result']['ticket_key']
    else:
        raise Exception(f"Failed to get password ticket: {result.get('msg')}")

def decrypt_ticket_key(ticket_key, access_key):
    """Step 3: Decrypt Ticket Key (AES-256-ECB)"""
    ticket_key_bytes = binascii.unhexlify(ticket_key)
    key_bytes = access_key.encode('utf-8').ljust(32, b'\x00')
    cipher = AES.new(key_bytes, AES.MODE_ECB)
    decrypted = cipher.decrypt(ticket_key_bytes)
    padding_len = decrypted[-1]
    if padding_len <= 16:
        decrypted = decrypted[:-padding_len]
    return binascii.hexlify(decrypted).decode().upper()

def encrypt_password(password, decrypted_key):
    """Step 4: Encrypt Password (AES-128-ECB + PKCS7Padding)"""
    key_bytes = binascii.unhexlify(decrypted_key[:32])  # First 16 bytes
    password_bytes = password.encode('utf-8')
    padded_password = pad(password_bytes, AES.block_size)
    cipher = AES.new(key_bytes, AES.MODE_ECB)
    encrypted = cipher.encrypt(padded_password)
    return binascii.hexlify(encrypted).decode().upper()

def create_temp_password(access_token, ticket_id, encrypted_password, password, valid_minutes=5):
    """Step 5: Create Temporary Password"""
    t = int(time.time() * 1000)
    nonce = str(uuid.uuid4())
    path = f"/v1.0/devices/{DEVICE_ID}/door-lock/temp-password"

    now_sec = int(time.time())
    body = {
        "password": encrypted_password,
        "password_type": "ticket",
        "ticket_id": ticket_id,
        "effective_time": now_sec,
        "invalid_time": now_sec + (valid_minutes * 60),
        "time_zone": "+07:00",
        "name": f"password_{password}"
    }

    body_json = json.dumps(body, separators=(',', ':'))
    body_sha256 = hashlib.sha256(body_json.encode()).hexdigest()

    string_to_sign = f"POST\n{body_sha256}\n\n{path}"
    sign_str = f"{CLIENT_ID}{access_token}{t}{nonce}{string_to_sign}"
    sign = hmac.new(SECRET.encode(), sign_str.encode(), hashlib.sha256).hexdigest().upper()

    headers = {"sign": sign, "sign_method": "HMAC-SHA256", "t": str(t), "nonce": nonce,
               "client_id": CLIENT_ID, "access_token": access_token, "Content-Type": "application/json"}

    response = requests.post(f"{BASE_URL}{path}", headers=headers, data=body_json)
    result = response.json()

    return result, body

def generate_password(password="7654321", valid_minutes=5):
    """Main function to generate a temporary password"""
    print("==========================================")
    print(f"  TUYA SMART LOCK - CREATE PASSWORD")
    print(f"  Password: {password}")
    print(f"  Duration: {valid_minutes} minutes")
    print("==========================================\n")

    # Step 1: Get Access Token
    print("=== Step 1: Get Access Token ===")
    access_token = get_access_token()
    print(f"✅ Access Token: {access_token}\n")

    # Step 2: Get Password Ticket
    print("=== Step 2: Get Password Ticket ===")
    ticket_id, ticket_key = get_password_ticket(access_token)
    print(f"✅ Ticket ID: {ticket_id}")
    print(f"✅ Ticket Key: {ticket_key}\n")

    # Step 3: Decrypt Ticket Key
    print("=== Step 3: Decrypt Ticket Key (AES-256-ECB) ===")
    decrypted_key = decrypt_ticket_key(ticket_key, SECRET)
    print(f"✅ Decrypted Key: {decrypted_key}\n")

    # Step 4: Encrypt Password
    print("=== Step 4: Encrypt Password (AES-128-ECB) ===")
    print(f"Plain Password: {password}")
    encrypted_password = encrypt_password(password, decrypted_key)
    print(f"✅ Encrypted Password: {encrypted_password}\n")

    # Step 5: Create Temporary Password
    print("=== Step 5: Create Temporary Password ===")
    result, body = create_temp_password(access_token, ticket_id, encrypted_password, password, valid_minutes)

    print(f"Request Body: {json.dumps(body, indent=2)}")
    print(f"\nResponse: {json.dumps(result, indent=2)}")

    if result.get('success'):
        password_id = result['result']['id']
        effective_time = time.strftime('%Y-%m-%d %H:%M:%S', time.localtime(body['effective_time']))
        invalid_time = time.strftime('%Y-%m-%d %H:%M:%S', time.localtime(body['invalid_time']))
        print(f"\n{'='*50}")
        print(f"✅ SUCCESS! Temporary Password Created!")
        print(f"{'='*50}")
        print(f"   Password: {password}")
        print(f"   Password ID: {password_id}")
        print(f"   Valid From: {effective_time}")
        print(f"   Valid Until: {invalid_time}")
        print(f"   Duration: {valid_minutes} minutes")
        print(f"   Device: {DEVICE_ID}")
        print(f"{'='*50}\n")
        return True, password_id
    else:
        print(f"\n❌ FAILED: {result.get('msg')} (code: {result.get('code')})")
        return False, None

if __name__ == "__main__":
    import sys

    # Default password or from command line
    password = sys.argv[1] if len(sys.argv) > 1 else "7654321"
    valid_minutes = int(sys.argv[2]) if len(sys.argv) > 2 else 5

    success, password_id = generate_password(password, valid_minutes)
    exit(0 if success else 1)
