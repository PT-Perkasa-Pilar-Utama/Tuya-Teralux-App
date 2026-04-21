#!/usr/bin/env python3
"""
Try to ADD a password to the door lock device
So user can type the password on keypad to unlock
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

def send_command(access_token, commands, test_name=""):
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
    
    print(f"\n=== {test_name} ===")
    print(f"Commands: {commands}")
    resp = requests.post(f'{BASE_URL}{path}', headers=headers, data=body_json)
    result = resp.json()
    print(f"Response: {json.dumps(result, indent=2)}")
    return result.get('success')

# Main
token = get_token()
if not token:
    print("Failed to get token")
    exit(1)

print(f"Device ID: {DEVICE_ID}")
print(f"Token: {token[:20]}...\n")
print("=" * 60)
print("TRYING TO ADD PASSWORD TO DEVICE")
print("=" * 60)

results = {}

# Method 1: Try add_password command
print("\n--- Method 1: add_password ---")
results['add_password'] = send_command(token, [
    {"code": "add_password", "value": "123456"}
], "Add password: 123456")

# Method 2: Try with password object
print("\n--- Method 2: Password object ---")
results['password_obj'] = send_command(token, [
    {"code": "add_password", "value": {"password": "888888", "name": "test"}}
], "Add password object")

# Method 3: Try update_all_password with list
print("\n--- Method 3: update_all_password list ---")
results['update_password'] = send_command(token, [
    {"code": "update_all_password", "value": [{"password": "999999", "status": "valid"}]}
], "Update password list")

# Method 4: Try set_password
print("\n--- Method 4: set_password ---")
results['set_password'] = send_command(token, [
    {"code": "set_password", "value": "777777"}
], "Set password: 777777")

# Method 5: Try password_mgmt
print("\n--- Method 5: password_mgmt ---")
results['password_mgmt'] = send_command(token, [
    {"code": "password_mgmt", "value": {"op": "add", "password": "666666"}}
], "Password management add")

# Method 6: Try dp_send with password
print("\n--- Method 6: dp_send ---")
results['dp_send'] = send_command(token, [
    {"code": "dp_send", "value": {"unlock_password": "555555"}}
], "DP send password")

print("\n" + "=" * 60)
print("SUMMARY")
print("=" * 60)
success_count = sum(1 for v in results.values() if v)
print(f"Successful: {success_count}/{len(results)}")
for test, success in results.items():
    status = "✅" if success else "❌"
    print(f"{status} {test}")
