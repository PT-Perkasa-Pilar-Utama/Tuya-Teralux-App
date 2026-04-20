#!/usr/bin/env python3
"""
Test sending commands to door lock device
Based on device debugging output from Tuya IoT Platform
"""
import requests, hmac, hashlib, time, uuid, json, os
from pathlib import Path

env_path = Path('.env')
if env_path.exists():
    with open(env_path) as f:
        for line in f:
            line = line.strip()
            if line and not line.startswith('#') and '=' in line:
                key, value = line.split('=', 1)
                os.environ[key.strip()] = value.strip()

CLIENT_ID = os.environ.get('TUYA_CLIENT_ID')
SECRET = os.environ.get('TUYA_ACCESS_SECRET')
BASE_URL = 'https://openapi-sg.iotbing.com'
DEVICE_ID = os.environ.get('TUYA_DEVICE_ID')

EMPTY_SHA256 = 'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855'

def get_token():
    t = int(time.time() * 1000)
    nonce = str(uuid.uuid4())
    path = '/v1.0/token?grant_type=1'
    sign_str = f'{CLIENT_ID}{t}{nonce}GET\n{EMPTY_SHA256}\n\n{path}'
    sign = hmac.new(SECRET.encode(), sign_str.encode(), hashlib.sha256).hexdigest().upper()
    headers = {'sign': sign, 'sign_method': 'HMAC-SHA256', 't': str(t), 'nonce': nonce, 'client_id': CLIENT_ID}
    resp = requests.get(f'{BASE_URL}{path}', headers=headers)
    result = resp.json()
    if result.get('success'):
        return result['result']['access_token']
    return None

def send_command(access_token, commands):
    """Send commands to device"""
    t = int(time.time() * 1000)
    nonce = str(uuid.uuid4())
    path = f'/v1.0/devices/{DEVICE_ID}/commands'
    
    body = {"commands": commands}
    body_json = json.dumps(body, separators=(',', ':'))
    body_hash = hashlib.sha256(body_json.encode()).hexdigest()
    
    sign_str = f'{CLIENT_ID}{access_token}{t}{nonce}POST\n{body_hash}\n\n{path}'
    sign = hmac.new(SECRET.encode(), sign_str.encode(), hashlib.sha256).hexdigest().upper()
    
    headers = {
        'sign': sign, 'sign_method': 'HMAC-SHA256', 't': str(t),
        'nonce': nonce, 'client_id': CLIENT_ID, 'access_token': access_token,
        'Content-Type': 'application/json'
    }
    resp = requests.post(f'{BASE_URL}{path}', headers=headers, data=body_json)
    return resp.json()

def test_unlock_request(access_token):
    """Test unlock_request command (0-90 seconds)"""
    print("\n=== Test: Unlock Request (3 seconds) ===")
    result = send_command(access_token, [
        {"code": "unlock_request", "value": 3}
    ])
    print(f"Response: {json.dumps(result, indent=2)}")
    return result.get('success')

def test_reply_unlock(access_token):
    """Test reply_unlock_request command"""
    print("\n=== Test: Reply Unlock Request ===")
    result = send_command(access_token, [
        {"code": "reply_unlock_request", "value": True}
    ])
    print(f"Response: {json.dumps(result, indent=2)}")
    return result.get('success')

def test_update_password_list(access_token):
    """Test update_all_password command"""
    print("\n=== Test: Update All Password List ===")
    # This is a list type, try sending password data
    result = send_command(access_token, [
        {"code": "update_all_password", "value": []}
    ])
    print(f"Response: {json.dumps(result, indent=2)}")
    return result.get('success')

# Main
token = get_token()
if not token:
    print("Failed to get token")
    exit(1)

print(f"Device ID: {DEVICE_ID}")
print(f"Token: {token[:20]}...\n")

# Test commands
print("=" * 60)
print("TESTING DEVICE COMMANDS")
print("=" * 60)

results = {}
results['unlock_request'] = test_unlock_request(token)
results['reply_unlock'] = test_reply_unlock(token)
results['update_password_list'] = test_update_password_list(token)

print("\n" + "=" * 60)
print("SUMMARY")
print("=" * 60)
for test, success in results.items():
    status = "✅" if success else "❌"
    print(f"{status} {test}: {'PASS' if success else 'FAIL'}")
