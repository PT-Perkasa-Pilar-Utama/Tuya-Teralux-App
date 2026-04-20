#!/usr/bin/env python3
"""
Simple Smart Life Traffic Capture
Start proxy and capture Smart Life app traffic

Usage:
    python3 capture.py
    
Then configure phone proxy to your computer IP:8080
"""

import subprocess
import sys
import socket
import json
from pathlib import Path
from datetime import datetime

CAPTURE_DIR = Path(__file__).parent / "captures"
CAPTURE_DIR.mkdir(exist_ok=True)

def get_local_ip():
    """Get local IP address"""
    try:
        s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        s.connect(("8.8.8.8", 80))
        ip = s.getsockname()[0]
        s.close()
        return ip
    except Exception:
        return "127.0.0.1"

def check_mitmproxy():
    """Check if mitmproxy is installed"""
    try:
        import mitmproxy
        return True
    except ImportError:
        return False

def main():
    print("\n" + "=" * 70)
    print("  SMART LIFE TRAFFIC CAPTURE")
    print("=" * 70)
    
    # Check mitmproxy
    if not check_mitmproxy():
        print("\n❌ mitmproxy not installed!")
        print("\nInstall with:")
        print("  pip3 install mitmproxy")
        print("\nOr:")
        print("  brew install mitmproxy (macOS)")
        print("  sudo apt-get install mitmproxy (Linux)")
        sys.exit(1)
    
    local_ip = get_local_ip()
    timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
    save_file = CAPTURE_DIR / f"capture_{timestamp}.mitm"
    
    print(f"\n  💾 Save to: {save_file}")
    print(f"  🌐 Proxy: {local_ip}:8080")
    print(f"  📊 Web UI: http://localhost:8081")
    
    print("\n" + "=" * 70)
    print("  SETUP YOUR PHONE")
    print("=" * 70)
    print(f"""
  1. WiFi Settings → Proxy → Manual
  
  2. Server: {local_ip}
     Port:   8080
  
  3. Install certificate:
     - Open browser: http://mitm.it
     - Download & install certificate
  
  4. Open Smart Life app
    """)
    
    print("=" * 70)
    print("\n  Press Ctrl+C to stop\n")
    
    # Start mitmweb
    cmd = [
        "mitmweb",
        "--listen-port", "8080",
        "--web-port", "8081",
        "--save-stream-log", str(save_file),
        "--set", "console=false"
    ]
    
    try:
        subprocess.run(cmd)
    except KeyboardInterrupt:
        print("\n\n✅ Capture stopped!")
        print(f"\n📁 Saved to: {save_file}")
        print(f"\n💡 Analyze with:")
        print(f"   mitmproxy -r {save_file}")
        print(f"   or")
        print(f"   python3 analyze_capture.py {save_file}")

if __name__ == "__main__":
    main()
