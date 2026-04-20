#!/usr/bin/env python3
"""
Smart Life Traffic Interceptor
Capture and analyze Smart Life app API requests

Usage:
    python3 interceptor.py
    
This will start a proxy server on port 8080.
Configure your phone to use this proxy, then open Smart Life app.
"""

import json
import os
import sys
from datetime import datetime
from pathlib import Path

# Check if mitmproxy is installed
try:
    from mitmproxy import ctx, http
except ImportError:
    print("❌ mitmproxy not installed!")
    print("\nInstall with:")
    print("  pip3 install mitmproxy")
    print("  or")
    print("  brew install mitmproxy")
    sys.exit(1)

# Configuration
CAPTURE_DIR = Path(__file__).parent / "captures"
CAPTURE_DIR.mkdir(exist_ok=True)

# Filter for Tuya/Smart Life domains
TARGET_DOMAINS = [
    'tuya.com',
    'iot.tuya.com', 
    'openapi',
    'smartlife.com',
    'tuyaus.com',
    'tuyacn.com',
    'tuyaeu.com'
]

# Filter for door lock related paths
DOOR_LOCK_PATHS = [
    'door-lock',
    'devices',
    'token',
    'password',
    'commands',
    'dynamic-password',
    'temp-password'
]


class SmartLifeInterceptor:
    """MITM Proxy addon for capturing Smart Life traffic"""
    
    def __init__(self):
        self.request_count = 0
        self.capture_file = None
        self.start_time = datetime.now()
        
    def start(self):
        """Called when proxy starts"""
        timestamp = self.start_time.strftime('%Y%m%d_%H%M%S')
        self.capture_file = CAPTURE_DIR / f"smartlife_{timestamp}.json"
        self.captures = []
        
        print("\n" + "=" * 70)
        print("  SMART LIFE TRAFFIC INTERCEPTOR")
        print("=" * 70)
        print(f"\n  📁 Capturing to: {self.capture_file}")
        print(f"  🌐 Proxy Port: 8080")
        print(f"  🎯 Target Domains: {', '.join(TARGET_DOMAINS)}")
        print("\n  📱 Configure your phone:")
        print("     1. WiFi Settings → Proxy → Manual")
        print(f"     2. Server: {self.get_local_ip()}")
        print("     3. Port: 8080")
        print("\n  🔐 Install certificate:")
        print("     1. Open browser on phone")
        print("     2. Go to: http://mitm.it")
        print("     3. Download and install certificate")
        print("\n" + "=" * 70 + "\n")
    
    def get_local_ip(self):
        """Get local IP address"""
        import socket
        try:
            s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
            s.connect(("8.8.8.8", 80))
            ip = s.getsockname()[0]
            s.close()
            return ip
        except Exception:
            return "127.0.0.1"
    
    def should_capture(self, flow: http.HTTPFlow) -> bool:
        """Check if this flow should be captured"""
        host = flow.request.host.lower()
        path = flow.request.path.lower()
        
        # Check domain
        if not any(domain in host for domain in TARGET_DOMAINS):
            return False
        
        return True
    
    def request(self, flow: http.HTTPFlow):
        """Called on HTTP request"""
        if not self.should_capture(flow):
            return
        
        self.request_count += 1
        
        # Log to console
        print(f"\n{'='*70}")
        print(f"  REQUEST #{self.request_count}")
        print(f"{'='*70}")
        print(f"  ⏰ Time: {datetime.now().strftime('%H:%M:%S')}")
        print(f"  🔹 Method: {flow.request.method}")
        print(f"  🔹 URL: {flow.request.url}")
        print(f"  🔹 Host: {flow.request.host}")
        print(f"  🔹 Path: {flow.request.path}")
        
        # Print headers
        if flow.request.headers:
            print(f"\n  📋 Headers:")
            for key, value in flow.request.headers.items():
                # Skip sensitive headers
                if key.lower() not in ['authorization', 'cookie', 'sign']:
                    print(f"     {key}: {value}")
                else:
                    print(f"     {key}: [REDACTED]")
        
        # Print body
        if flow.request.content:
            try:
                body = flow.request.content.decode('utf-8')
                print(f"\n  📦 Request Body:")
                try:
                    # Try to pretty print JSON
                    body_json = json.loads(body)
                    print(f"     {json.dumps(body_json, indent=2)[:1000]}")
                except json.JSONDecodeError:
                    # Not JSON, print as is
                    print(f"     {body[:500]}")
            except UnicodeDecodeError:
                print(f"     [Binary content, {len(flow.request.content)} bytes]")
    
    def response(self, flow: http.HTTPFlow):
        """Called on HTTP response"""
        if not self.should_capture(flow):
            return
        
        # Log to console
        print(f"\n{'='*70}")
        print(f"  RESPONSE #{self.request_count}")
        print(f"{'='*70}")
        print(f"  ⏰ Time: {datetime.now().strftime('%H:%M:%S')}")
        print(f"  🔸 Status: {flow.response.status_code}")
        print(f"  🔸 URL: {flow.request.url}")
        
        # Print headers
        if flow.response.headers:
            print(f"\n  📋 Headers:")
            for key, value in flow.response.headers.items():
                print(f"     {key}: {value}")
        
        # Print body
        if flow.response.content:
            try:
                body = flow.response.content.decode('utf-8')
                print(f"\n  📦 Response Body:")
                try:
                    # Try to pretty print JSON
                    body_json = json.loads(body)
                    print(f"     {json.dumps(body_json, indent=2)[:1000]}")
                    
                    # Highlight important fields
                    if 'result' in body_json:
                        print(f"\n  ✅ Result: {body_json['result']}")
                    if 'success' in body_json:
                        print(f"  {'✅' if body_json['success'] else '❌'} Success: {body_json['success']}")
                    if 'code' in body_json and body_json['code'] != 0:
                        print(f"  ⚠️  Code: {body_json['code']}")
                    if 'msg' in body_json:
                        print(f"  💬 Message: {body_json['msg']}")
                        
                except json.JSONDecodeError:
                    # Not JSON, print as is
                    print(f"     {body[:500]}")
            except UnicodeDecodeError:
                print(f"     [Binary content, {len(flow.response.content)} bytes]")
        
        # Save to captures list
        self.save_capture(flow)
    
    def save_capture(self, flow: http.HTTPFlow):
        """Save capture to file"""
        capture = {
            'timestamp': datetime.now().isoformat(),
            'request': {
                'method': flow.request.method,
                'url': flow.request.url,
                'host': flow.request.host,
                'path': flow.request.path,
                'headers': dict(flow.request.headers),
                'body': flow.request.get_text() if flow.request.content else None
            },
            'response': {
                'status_code': flow.response.status_code,
                'headers': dict(flow.response.headers),
                'body': flow.response.get_text() if flow.response.content else None
            }
        }
        
        self.captures.append(capture)
        
        # Save to file periodically
        if len(self.captures) % 10 == 0:
            self.save_to_file()
    
    def done(self):
        """Called when proxy stops"""
        self.save_to_file()
        
        print("\n" + "=" * 70)
        print("  CAPTURE COMPLETE")
        print("=" * 70)
        print(f"\n  📊 Total Requests Captured: {len(self.captures)}")
        print(f"  📁 Saved to: {self.capture_file}")
        print("\n  💡 To analyze:")
        print(f"     cat {self.capture_file} | jq .")
        print("\n" + "=" * 70 + "\n")
    
    def save_to_file(self):
        """Save captures to JSON file"""
        if not hasattr(self, 'captures') or not self.captures:
            return
        
        with open(self.capture_file, 'w') as f:
            json.dump(self.captures, f, indent=2, default=str)
        
        print(f"\n  💾 Saved {len(self.captures)} captures to {self.capture_file}")


# MITM Proxy addons
addons = [SmartLifeInterceptor()]


# Helper script to analyze captures
def analyze_captures():
    """Analyze captured traffic"""
    print("\n" + "=" * 70)
    print("  ANALYZING CAPTURES")
    print("=" * 70 + "\n")
    
    # Find latest capture file
    capture_files = list(CAPTURE_DIR.glob("smartlife_*.json"))
    if not capture_files:
        print("❌ No capture files found!")
        return
    
    latest = max(capture_files, key=lambda p: p.stat().st_mtime)
    print(f"📁 Reading: {latest}\n")
    
    with open(latest) as f:
        captures = json.load(f)
    
    # Filter for door lock related
    door_lock_captures = []
    for cap in captures:
        path = cap['request']['path'].lower()
        if any(kw in path for kw in DOOR_LOCK_PATHS):
            door_lock_captures.append(cap)
    
    print(f"📊 Total Captures: {len(captures)}")
    print(f"🚪 Door Lock Related: {len(door_lock_captures)}\n")
    
    if door_lock_captures:
        print("=" * 70)
        print("  DOOR LOCK REQUESTS")
        print("=" * 70)
        
        for i, cap in enumerate(door_lock_captures, 1):
            print(f"\n{i}. {cap['request']['method']} {cap['request']['path']}")
            
            if cap['request']['body']:
                try:
                    body = json.loads(cap['request']['body'])
                    print(f"   Request: {json.dumps(body, indent=2)[:300]}")
                except:
                    print(f"   Request: {cap['request']['body'][:200]}")
            
            if cap['response']['body']:
                try:
                    body = json.loads(cap['response']['body'])
                    print(f"   Response: {json.dumps(body, indent=2)[:300]}")
                    
                    # Extract key info
                    if 'success' in body:
                        print(f"   ✅ Success: {body['success']}")
                    if 'result' in body:
                        result = body['result']
                        if isinstance(result, dict):
                            if 'password' in result:
                                print(f"   🔑 Password: {result['password']}")
                            if 'id' in result:
                                print(f"   📝 ID: {result['id']}")
                except:
                    print(f"   Response: {cap['response']['body'][:200]}")


if __name__ == '__main__':
    if len(sys.argv) > 1 and sys.argv[1] == 'analyze':
        analyze_captures()
    else:
        print("Starting Smart Life Traffic Interceptor...")
        print("Press Ctrl+C to stop and save captures.")
        print("\nTip: Run 'python3 interceptor.py analyze' to analyze captures.")
