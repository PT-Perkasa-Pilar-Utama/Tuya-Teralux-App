#!/usr/bin/env python3
"""
Check device specifications and try different password API formats
"""
import requests, hmac, hashlib, time, uuid, json, os
from pathlib import Path

# Load .env
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

def get_device_specs(access_token):
    """Get device specifications"""
    t = int(time.time() * 1000)
    nonce = str(uuid.uuid4())
    path = f'/v1.0/devices/{DEVICE_ID}/specifications'
    sign_str = f'{CLIENT_ID}{access_token}{t}{nonce}GET\n{EMPTY_SHA256}\n\n{path}'
    sign = hmac.new(SECRET.encode(), sign_str.encode(), hashlib.sha256).hexdigest().upper()
    headers = {'sign': sign, 'sign_method': 'HMAC-SHA256', 't': str(t), 'nonce': nonce, 'client_id': CLIENT_ID, 'access_token': access_token}
    resp = requests.get(f'{BASE_URL}{path}', headers=headers)
    return resp.json()

def get_device_status(access_token):
    """Get device status"""
    t = int(time.time() * 1000)
    nonce = str(uuid.uuid4())
    path = f'/v1.0/devices/{DEVICE_ID}'
    sign_str = f'{CLIENT_ID}{access_token}{t}{nonce}GET\n{EMPTY_SHA256}\n\n{path}'
    sign = hmac.new(SECRET.encode(), sign_str.encode(), hashlib.sha256).hexdigest().upper()
    headers = {'sign': sign, 'sign_method': 'HMAC-SHA256', 't': str(t), 'nonce': nonce, 'client_id': CLIENT_ID, 'access_token': access_token}
    resp = requests.get(f'{BASE_URL}{path}', headers=headers)
    return resp.json()

def test_password_api(access_token, body, test_name):
    """Test password creation with different body formats"""
    print(f"\n=== {test_name} ===")
    t = int(time.time() * 1000)
    nonce = str(uuid.uuid4())
    path = f'/v1.0/devices/{DEVICE_ID}/door-lock/temp-password'
    
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
    result = resp.json()
    print(f"Body: {body}")
    print(f"Response: {json.dumps(result, indent=2)}")
    return result.get('success')

# Main
token = get_token()
if not token:
    print("Failed to get token")
    exit(1)

print(f"Device ID: {DEVICE_ID}")
print(f"Token: {token[:20]}...\n")

# Get device specs
print("=== Device Specifications ===")
specs = get_device_specs(token)
print(json.dumps(specs, indent=2))

# Get device status
print("\n=== Device Status ===")
status = get_device_status(token)
print(json.dumps(status, indent=2))

# Test different body formats
print("\n\n=== Testing Password APIs ===")

test_password_api(token, {"password_type": 2, "valid_time": 5}, "Test 1: password_type=2, valid_time")
test_password_api(token, {"password_type": "2", "valid_time": "5"}, "Test 2: password_type as string")
test_password_api(token, {"password_type": 2, "valid_time": 300}, "Test 3: valid_time in seconds")
test_password_api(token, {"type": 2, "valid_time": 5}, "Test 4: 'type' instead of 'password_type'")
test_password_api(token, {"password_type": 2, "minute": 5}, "Test 5: 'minute' parameter")
test_password_api(token, {"password_type": 2, "timeout": 5}, "Test 6: 'timeout' parameter")
