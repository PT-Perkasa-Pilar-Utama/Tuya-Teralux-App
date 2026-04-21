#!/bin/bash
# Smart Life Traffic Capture - Quick Setup
# This script helps you set up MITM proxy to capture Smart Life app traffic

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CAPTURE_DIR="$SCRIPT_DIR/captures"

echo "=============================================================="
echo "  SMART LIFE TRAFFIC CAPTURE - SETUP"
echo "=============================================================="

# Check if mitmproxy is installed
if ! command -v mitmweb &> /dev/null; then
    echo ""
    echo "❌ mitmproxy not found!"
    echo ""
    echo "Install mitmproxy:"
    echo "  macOS:   brew install mitmproxy"
    echo "  Linux:   sudo apt-get install mitmproxy"
    echo "  Pip:     pip3 install mitmproxy"
    echo ""
    exit 1
fi

echo "✅ mitmproxy found: $(which mitmweb)"

# Create capture directory
mkdir -p "$CAPTURE_DIR"
echo "✅ Capture directory: $CAPTURE_DIR"

# Get local IP
LOCAL_IP=$(hostname -I | awk '{print $1}')
echo "✅ Your IP: $LOCAL_IP"

echo ""
echo "=============================================================="
echo "  SETUP INSTRUCTIONS"
echo "=============================================================="
echo ""
echo "1. Start the proxy server:"
echo "   $ mitmweb --listen-port 8080"
echo ""
echo "2. Configure your phone WiFi:"
echo "   - Go to Settings → WiFi"
echo "   - Long press your network → Modify"
echo "   - Proxy: Manual"
echo "   - Server: $LOCAL_IP"
echo "   - Port: 8080"
echo ""
echo "3. Install certificate on phone:"
echo "   - Open browser on phone"
echo "   - Go to: http://mitm.it"
echo "   - Download and install certificate"
echo ""
echo "4. Open Smart Life app and perform actions"
echo ""
echo "5. Check mitmweb UI at: http://localhost:8081"
echo ""
echo "=============================================================="
echo ""
echo "Starting mitmweb..."
echo ""

# Start mitmweb
mitmweb --listen-port 8080 \
    --set confdir="$CAPTURE_DIR/.mitmproxy" \
    --set console=false \
    --set web_port=8081
