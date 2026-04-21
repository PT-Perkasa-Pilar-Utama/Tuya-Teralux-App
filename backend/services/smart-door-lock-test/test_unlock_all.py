#!/usr/bin/env python3
"""
Try ALL possible ways to unlock the door via API
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

def send_instruction(access_token, instruction, value, test_name=""):
    """Send single instruction"""
    return send_command(access_token, [{"code": instruction, "value": value}], test_name)

# Main
token = get_token()
if not token:
    print("Failed to get token")
    exit(1)

print(f"Device ID: {DEVICE_ID}")
print(f"Token: {token[:20]}...\n")
print("=" * 60)
print("TRYING ALL UNLOCK METHODS")
print("=" * 60)

results = {}

# Method 1: unlock_request with various values
print("\n--- Method 1: unlock_request ---")
results['unlock_request_1'] = send_instruction(token, "unlock_request", 1, "unlock_request: 1 second")
results['unlock_request_5'] = send_instruction(token, "unlock_request", 5, "unlock_request: 5 seconds")
results['unlock_request_10'] = send_instruction(token, "unlock_request", 10, "unlock_request: 10 seconds")

# Method 2: reply_unlock_request
print("\n--- Method 2: reply_unlock_request ---")
results['reply_unlock_true'] = send_instruction(token, "reply_unlock_request", True, "reply_unlock_request: true")
results['reply_unlock_false'] = send_instruction(token, "reply_unlock_request", False, "reply_unlock_request: false")

# Method 3: Try direct unlock codes
print("\n--- Method 3: Direct unlock attempts ---")
results['unlock_direct'] = send_instruction(token, "unlock", True, "Direct unlock: true")
results['lock_direct'] = send_instruction(token, "lock", False, "Direct lock: false")

# Method 4: Try boolean variations
print("\n--- Method 4: Boolean variations ---")
results['unlock_request_true'] = send_instruction(token, "unlock_request", True, "unlock_request: true")
results['unlock_request_string'] = send_instruction(token, "unlock_request", "unlock", "unlock_request: 'unlock'")

# Method 5: Try with empty/zero values first then trigger
print("\n--- Method 5: Reset then unlock ---")
results['unlock_request_zero'] = send_instruction(token, "unlock_request", 0, "unlock_request: 0 (reset)")
time.sleep(1)
results['unlock_request_after'] = send_instruction(token, "unlock_request", 3, "unlock_request: 3 (after reset)")

print("\n" + "=" * 60)
print("SUMMARY")
print("=" * 60)
success_count = sum(1 for v in results.values() if v)
print(f"Successful: {success_count}/{len(results)}")
for test, success in results.items():
    status = "✅" if success else "❌"
    print(f"{status} {test}")
