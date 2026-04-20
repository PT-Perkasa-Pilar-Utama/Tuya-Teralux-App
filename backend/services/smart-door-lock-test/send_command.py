#!/usr/bin/env python3
"""
Tuya Smart Door Lock - Send Command Script
Send commands to your door lock device via Tuya Cloud API

Usage:
    python3 send_command.py

Edit the COMMANDS list below to send different commands to your device.
Available commands based on device specifications:

INTEGER Commands (0-999):
    - unlock_fingerprint
    - unlock_password
    - unlock_temporary
    - unlock_dynamic
    - unlock_card
    - unlock_face
    - unlock_key
    - unlock_request (0-90)
    - residual_electricity (0-100)
    - unlock_app

BOOLEAN Commands (true/false):
    - hijack
    - open_inside
    - doorbell
    - anti_lock_outside

ENUM Commands:
    - closed_opened: "unknown", "open", "closed"

LIST Commands:
    - update_all_finger: []
    - update_all_password: []
    - update_all_card: []
    - update_all_face: []
"""

import requests
import hmac
import hashlib
import time
import uuid
import json
import os
from pathlib import Path
from datetime import datetime

# =============================================================================
# CONFIGURATION - Edit this section
# =============================================================================

# Command(s) to send to the device
# Format: {"code": "<command_name>", "value": <value>}
# You can send multiple commands at once
COMMANDS = [
    # Example: Send unlock request for 3 seconds
    # {"code": "unlock_request", "value": 3},
    
    # Example: Set door status
    # {"code": "closed_opened", "value": "open"},
    
    # Example: Trigger doorbell
    # {"code": "doorbell", "value": True},
    
    # Example: Update password list (empty list to clear)
    # {"code": "update_all_password", "value": []},
    
    # ADD YOUR COMMAND HERE
    {"code": "reply_unlock_request", "value": True},
]

# =============================================================================
# END CONFIGURATION - No need to edit below this line
# =============================================================================

# Load credentials from .env file
env_path = Path(__file__).parent / ".env"
if not env_path.exists():
    print(f"❌ Error: .env file not found at {env_path}")
    print("\nCreate a .env file with:")
    print("  TUYA_CLIENT_ID=your_client_id")
    print("  TUYA_ACCESS_SECRET=your_access_secret")
    print("  TUYA_DEVICE_ID=your_device_id")
    exit(1)

# Parse .env file
with open(env_path) as f:
    for line in f:
        line = line.strip()
        if line and not line.startswith('#') and '=' in line:
            key, value = line.split('=', 1)
            os.environ[key.strip()] = value.strip()

# Get credentials from environment
CLIENT_ID = os.environ.get('TUYA_CLIENT_ID')
SECRET = os.environ.get('TUYA_ACCESS_SECRET')
BASE_URL = os.environ.get('TUYA_BASE_URL', 'https://openapi-sg.iotbing.com')
DEVICE_ID = os.environ.get('TUYA_DEVICE_ID')

# Validate credentials
if not CLIENT_ID or not SECRET or not DEVICE_ID:
    print("❌ Error: Missing required credentials")
    print("\nRequired environment variables:")
    print("  TUYA_CLIENT_ID - Your Tuya client ID")
    print("  TUYA_ACCESS_SECRET - Your Tuya access secret")
    print("  TUYA_DEVICE_ID - Your door lock device ID")
    exit(1)

# SHA256 constant for empty body
EMPTY_SHA256 = 'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855'


def get_access_token():
    """Get access token from Tuya API"""
    t = int(time.time() * 1000)
    nonce = str(uuid.uuid4())
    path = '/v1.0/token?grant_type=1'
    
    # Generate signature
    sign_str = f'{CLIENT_ID}{t}{nonce}GET\n{EMPTY_SHA256}\n\n{path}'
    sign = hmac.new(SECRET.encode(), sign_str.encode(), hashlib.sha256).hexdigest().upper()
    
    headers = {
        'sign': sign,
        'sign_method': 'HMAC-SHA256',
        't': str(t),
        'nonce': nonce,
        'client_id': CLIENT_ID
    }
    
    try:
        resp = requests.get(f'{BASE_URL}{path}', headers=headers, timeout=10)
        result = resp.json()
        
        if result.get('success'):
            return result['result']['access_token']
        else:
            print(f"❌ Failed to get access token: {result.get('msg')} (code: {result.get('code')})")
            return None
    except Exception as e:
        print(f"❌ Error getting access token: {e}")
        return None


def send_commands(access_token, commands):
    """Send commands to the device"""
    t = int(time.time() * 1000)
    nonce = str(uuid.uuid4())
    path = f'/v1.0/devices/{DEVICE_ID}/commands'
    
    # Prepare request body
    body = {"commands": commands}
    body_json = json.dumps(body, separators=(',', ':'))
    body_hash = hashlib.sha256(body_json.encode()).hexdigest()
    
    # Generate signature
    sign_str = f'{CLIENT_ID}{access_token}{t}{nonce}POST\n{body_hash}\n\n{path}'
    sign = hmac.new(SECRET.encode(), sign_str.encode(), hashlib.sha256).hexdigest().upper()
    
    headers = {
        'sign': sign,
        'sign_method': 'HMAC-SHA256',
        't': str(t),
        'nonce': nonce,
        'client_id': CLIENT_ID,
        'access_token': access_token,
        'Content-Type': 'application/json'
    }
    
    try:
        resp = requests.post(f'{BASE_URL}{path}', headers=headers, data=body_json, timeout=10)
        result = resp.json()
        return result
    except Exception as e:
        print(f"❌ Error sending command: {e}")
        return None


def get_device_status(access_token):
    """Get current device status"""
    t = int(time.time() * 1000)
    nonce = str(uuid.uuid4())
    path = f'/v1.0/devices/{DEVICE_ID}'
    
    sign_str = f'{CLIENT_ID}{access_token}{t}{nonce}GET\n{EMPTY_SHA256}\n\n{path}'
    sign = hmac.new(SECRET.encode(), sign_str.encode(), hashlib.sha256).hexdigest().upper()
    
    headers = {
        'sign': sign,
        'sign_method': 'HMAC-SHA256',
        't': str(t),
        'nonce': nonce,
        'client_id': CLIENT_ID,
        'access_token': access_token
    }
    
    try:
        resp = requests.get(f'{BASE_URL}{path}', headers=headers, timeout=10)
        return resp.json()
    except Exception as e:
        print(f"❌ Error getting device status: {e}")
        return None


def print_separator(title=""):
    """Print a visual separator"""
    width = 70
    print("\n" + "=" * width)
    if title:
        print(f"  {title}")
        print("=" * width)


def main():
    """Main function"""
    print_separator("TUYA SMART DOOR LOCK - SEND COMMAND")
    
    print(f"\n📅 Time: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print(f"🔧 Device ID: {DEVICE_ID}")
    print(f"🌐 Base URL: {BASE_URL}")
    
    # Show commands to be sent
    print_separator("COMMANDS TO SEND")
    for i, cmd in enumerate(COMMANDS, 1):
        print(f"  {i}. Code: {cmd['code']}, Value: {cmd['value']}")
    
    # Get access token
    print_separator("STEP 1: GET ACCESS TOKEN")
    access_token = get_access_token()
    
    if not access_token:
        print("\n❌ Failed to proceed without access token")
        exit(1)
    
    print(f"✅ Access Token: {access_token[:30]}...")
    
    # Get current device status (optional)
    print_separator("STEP 2: GET CURRENT DEVICE STATUS")
    status_result = get_device_status(access_token)
    
    if status_result and status_result.get('success'):
        device = status_result.get('result', {})
        print(f"  📱 Name: {device.get('name', 'N/A')}")
        print(f"  📶 Online: {device.get('online', False)}")
        print(f"  🔋 Category: {device.get('category', 'N/A')}")
        
        statuses = device.get('status', [])
        if statuses:
            print(f"\n  Current Status ({len(statuses)} items):")
            for s in statuses[:10]:  # Show first 10
                print(f"    • {s['code']}: {s['value']}")
            if len(statuses) > 10:
                print(f"    ... and {len(statuses) - 10} more")
    else:
        print(f"  ⚠️  Could not get device status")
    
    # Send commands
    print_separator("STEP 3: SEND COMMANDS")
    print("  Sending commands to device...")
    
    result = send_commands(access_token, COMMANDS)
    
    if result is None:
        print("\n❌ No response from API")
        exit(1)
    
    # Show result
    print_separator("RESULT")
    print(json.dumps(result, indent=2))
    
    if result.get('success'):
        print("\n✅ Commands sent successfully!")
        print("\n💡 Note: Some commands may take a few seconds to take effect.")
        print("   Check the Smart Life app or run this script again to verify.")
    else:
        error_code = result.get('code', 'N/A')
        error_msg = result.get('msg', 'Unknown error')
        print(f"\n❌ Command failed: {error_msg} (code: {error_code})")
        
        # Common error codes
        if error_code == 2008:
            print("\n💡 Error 2008: Command or value not supported by this device")
            print("   Check if the command code and value are valid for your device.")
        elif error_code == 1106:
            print("\n💡 Error 1106: Permission denied")
            print("   Make sure the device is linked to your Tuya cloud project.")
        elif error_code == 2001:
            print("\n💡 Error 2001: Device is offline")
            print("   Check if the device is connected to WiFi.")
    
    print_separator()


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        print("\n\n⚠️  Interrupted by user")
        exit(0)
    except Exception as e:
        print(f"\n❌ Unexpected error: {e}")
        exit(1)
