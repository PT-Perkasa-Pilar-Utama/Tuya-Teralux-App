#!/usr/bin/env python3
"""
Analyze Smart Life Traffic Capture
Parse and display captured traffic in a readable format

Usage:
    python3 analyze_capture.py [capture_file]
    
If no file specified, uses the latest capture.
"""

import json
import sys
from pathlib import Path
from datetime import datetime

CAPTURE_DIR = Path(__file__).parent / "captures"

# Keywords for filtering
DOOR_LOCK_KEYWORDS = [
    'door-lock',
    'password',
    'temp-password',
    'dynamic-password',
    'unlock',
    'lock',
    'commands',
    '/devices/'
]


def load_capture(filepath):
    """Load capture file"""
    print(f"📂 Loading: {filepath}")
    
    # Try different formats
    if filepath.suffix == '.mitm':
        # MITM flow format
        return load_mitm_flow(filepath)
    elif filepath.suffix == '.json':
        # JSON format
        with open(filepath) as f:
            return json.load(f)
    else:
        # Try as text
        return load_mitm_flow(filepath)


def load_mitm_flow(filepath):
    """Load MITM proxy flow file"""
    captures = []
    
    try:
        with open(filepath, 'r', encoding='utf-8', errors='ignore') as f:
            content = f.read()
            
        # Parse flow entries
        lines = content.split('\n')
        current_flow = None
        
        for line in lines:
            if line.startswith('{'):
                try:
                    flow = json.loads(line)
                    if 'request' in flow and 'response' in flow:
                        captures.append(flow)
                except json.JSONDecodeError:
                    pass
    
    except Exception as e:
        print(f"⚠️  Error loading file: {e}")
    
    return captures


def filter_door_lock(captures):
    """Filter for door lock related traffic"""
    filtered = []
    
    for cap in captures:
        if isinstance(cap, dict):
            request = cap.get('request', {})
            url = request.get('url', '')
            path = request.get('path', '')
            
            # Check if any keyword matches
            if any(kw in url.lower() or kw in path.lower() for kw in DOOR_LOCK_KEYWORDS):
                filtered.append(cap)
    
    return filtered


def format_body(body, max_length=500):
    """Format request/response body"""
    if not body:
        return None
    
    try:
        # Try to parse as JSON
        if isinstance(body, str):
            body_json = json.loads(body)
            return json.dumps(body_json, indent=2)[:max_length]
        elif isinstance(body, dict):
            return json.dumps(body, indent=2)[:max_length]
    except:
        pass
    
    # Return as string
    return str(body)[:max_length]


def analyze_capture(capture):
    """Analyze a single capture"""
    request = capture.get('request', {})
    response = capture.get('response', {})
    
    print("\n" + "=" * 70)
    
    # Request info
    method = request.get('method', 'UNKNOWN')
    url = request.get('url', '')
    path = request.get('path', '')
    host = request.get('host', '')
    
    print(f"🔹 REQUEST: {method} {path}")
    print(f"   Host: {host}")
    print(f"   URL: {url}")
    
    # Request headers
    headers = request.get('headers', {})
    if headers:
        print(f"\n   Headers:")
        for key, value in list(headers.items())[:5]:
            if key.lower() not in ['sign', 'authorization', 'cookie']:
                print(f"     {key}: {value}")
            else:
                print(f"     {key}: [REDACTED]")
    
    # Request body
    body = request.get('body')
    if body:
        print(f"\n   Request Body:")
        formatted = format_body(body)
        if formatted:
            for line in formatted.split('\n'):
                print(f"     {line}")
    
    # Response info
    status = response.get('status_code', 0)
    print(f"\n🔸 RESPONSE: {status}")
    
    # Response body
    body = response.get('body')
    if body:
        print(f"\n   Response Body:")
        formatted = format_body(body, max_length=800)
        if formatted:
            for line in formatted.split('\n'):
                print(f"     {line}")
            
            # Extract key fields
            try:
                if isinstance(body, str):
                    body_json = json.loads(body)
                else:
                    body_json = body
                
                if 'success' in body_json:
                    icon = "✅" if body_json['success'] else "❌"
                    print(f"\n   {icon} Success: {body_json['success']}")
                
                if 'result' in body_json:
                    result = body_json['result']
                    if isinstance(result, dict):
                        if 'password' in result:
                            print(f"   🔑 Password: {result['password']}")
                        if 'id' in result:
                            print(f"   📝 ID: {result['id']}")
                        if 'ticket_id' in result:
                            print(f"   🎫 Ticket ID: {result['ticket_id']}")
                        if 'access_token' in result:
                            print(f"   🔑 Access Token: {result['access_token'][:30]}...")
                
                if 'code' in body_json and body_json['code'] != 0:
                    print(f"   ⚠️  Code: {body_json['code']}")
                if 'msg' in body_json:
                    print(f"   💬 Message: {body_json['msg']}")
                    
            except:
                pass
    
    print()


def main():
    print("\n" + "=" * 70)
    print("  SMART LIFE TRAFFIC ANALYZER")
    print("=" * 70)
    
    # Get capture file
    if len(sys.argv) > 1:
        capture_file = Path(sys.argv[1])
    else:
        # Find latest capture
        if not CAPTURE_DIR.exists():
            print(f"❌ Capture directory not found: {CAPTURE_DIR}")
            sys.exit(1)
        
        captures = list(CAPTURE_DIR.glob("capture_*.mitm")) + \
                   list(CAPTURE_DIR.glob("smartlife_*.json"))
        
        if not captures:
            print("❌ No capture files found!")
            print("\nRun capture.py first to capture traffic.")
            sys.exit(1)
        
        capture_file = max(captures, key=lambda p: p.stat().st_mtime)
    
    # Load captures
    captures = load_capture(capture_file)
    
    if not captures:
        print("❌ No captures found in file!")
        sys.exit(1)
    
    print(f"\n📊 Total Captures: {len(captures)}")
    
    # Filter for door lock
    door_lock_captures = filter_door_lock(captures)
    print(f"🚪 Door Lock Related: {len(door_lock_captures)}")
    
    if not door_lock_captures:
        print("\n💡 No door lock traffic found.")
        print("   Try creating a temporary password in Smart Life app.")
        sys.exit(0)
    
    # Analyze each
    print("\n" + "=" * 70)
    print("  DOOR LOCK TRAFFIC")
    print("=" * 70)
    
    for i, cap in enumerate(door_lock_captures, 1):
        print(f"\n{'='*70}")
        print(f"  CAPTURE #{i}")
        analyze_capture(cap)
    
    # Summary
    print("\n" + "=" * 70)
    print("  SUMMARY")
    print("=" * 70)
    
    # Group by endpoint
    endpoints = {}
    for cap in door_lock_captures:
        path = cap.get('request', {}).get('path', 'unknown')
        if path not in endpoints:
            endpoints[path] = 0
        endpoints[path] += 1
    
    print("\n  Endpoints hit:")
    for path, count in sorted(endpoints.items(), key=lambda x: -x[1]):
        print(f"    {count}x  {path}")
    
    print("\n" + "=" * 70)
    print("\n💡 To export specific request:")
    print("   mitmproxy -r <capture_file> --save-stream-log=output.har")
    print("\n" + "=" * 70 + "\n")


if __name__ == "__main__":
    main()
