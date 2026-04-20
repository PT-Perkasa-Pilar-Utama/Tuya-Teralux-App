# Smart Life App Traffic Interception Guide

## Method 1: Browser DevTools (Easiest)

If Smart Life has a web version, you can inspect it directly:

### Steps:
1. Open Smart Life web version (if available)
2. Press `F12` to open DevTools
3. Go to **Network** tab
4. Perform actions in the app (create password, unlock, etc.)
5. Inspect the API requests

---

## Method 2: MITM Proxy (Recommended for Mobile App)

### Prerequisites
```bash
# Install mitmproxy
pip3 install mitmproxy

# Or on macOS
brew install mitmproxy

# Or on Linux
sudo apt-get install mitmproxy
```

### Setup Steps

#### 1. Start MITM Proxy
```bash
# Start mitmproxy with transparent mode
mitmproxy --mode transparent --listen-port 8080

# Or start mitmweb (web interface)
mitmweb --mode transparent --listen-port 8080 --web-port 8081
```

#### 2. Configure Phone

**Android:**
1. Go to Settings → WiFi
2. Long press your WiFi network → Modify Network
3. Advanced Options → Proxy → Manual
4. Proxy hostname: Your computer IP (e.g., `192.168.1.100`)
5. Proxy port: `8080`
6. Save

**iOS:**
1. Go to Settings → WiFi
2. Tap your WiFi network → Configure Proxy → Manual
3. Server: Your computer IP
4. Port: `8080`
5. Save

#### 3. Install CA Certificate on Phone

**Android (requires root for full HTTPS interception):**
1. Open browser on phone
2. Go to `http://mitm.it`
3. Download Android certificate
4. Install certificate:
   - Settings → Security → Encryption & Credentials
   - Install a Certificate → CA Certificate
   - Select downloaded certificate

**iOS:**
1. Open browser on phone
2. Go to `http://mitm.it`
3. Download iOS certificate
4. Install:
   - Settings → Profile Downloaded → Install
   - Settings → General → About → Certificate Trust Settings
   - Enable full trust for mitmproxy certificate

#### 4. Install Certificate as System Certificate (Android, requires root)
```bash
# Pull certificate from phone
adb pull /sdcard/download/mitmproxy-ca-cert.cer

# Convert to PEM
openssl x509 -inform DER -in mitmproxy-ca-cert.cer -out mitmproxy-ca-cert.pem

# Push to system certificate store (requires root)
adb push mitmproxy-ca-cert.pem /system/etc/security/cacerts/
adb shell chmod 644 /system/etc/security/cacerts/mitmproxy-ca-cert.pem
```

---

## Method 3: Using Frida (Advanced)

For apps with SSL pinning:

### Install Frida
```bash
pip3 install frida-tools
```

### Create SSL Pinning Bypass Script
```javascript
// ssl_bypass.js
Java.perform(function() {
    // Bypass SSL pinning for Smart Life
    var CertificatePinner = Java.use('okhttp3.CertificatePinner');
    CertificatePinner.check.overload('java.lang.String', 'java.util.List').implementation = function() {
        console.log('SSL pinning bypassed for: ' + arguments[0]);
        return;
    };
    
    // Bypass for other libraries
    var TrustManager = Java.use('javax.net.ssl.TrustManager');
    TrustManager.checkServerTrusted.implementation = function() {
        console.log('TrustManager bypassed');
    };
});
```

### Run Frida
```bash
# Find Smart Life process
frida-ps -U

# Attach to Smart Life with bypass script
frida -U -f "com.tuya.smartlife" -l ssl_bypass.js
```

---

## Capture Traffic

### Start Capturing
```bash
# Save traffic to file
mitmproxy --mode transparent --listen-port 8080 --set console=false --save-stream-log=/path/to/smartlife.mitm

# Or use mitmdump
mitmdump -w /path/to/smartlife.mitm
```

### Filter Smart Life Traffic
```bash
# View saved traffic
mitmproxy -r /path/to/smartlife.mitm

# Filter by host
mitmproxy -r /path/to/smartlife.mitm --set "view_filter=~host tuya"
```

---

## Analyze Traffic

### Look for These Endpoints:
- `/v1.0/devices/*/door-lock/*` - Door lock commands
- `/v1.0/token` - Authentication
- `/v1.0/users/*/devices` - Device list

### Export as HAR File
```bash
# Export to HAR for Chrome DevTools
mitmproxy -r /path/to/smartlife.mitm --save-stream-log=/path/to/output.har
```

---

## Quick Script: Auto-Capture Smart Life Traffic

```python
#!/usr/bin/env python3
"""
Auto-capture Smart Life app traffic
"""
from mitmproxy import ctx, http

class SmartLifeLogger:
    def request(self, flow: http.HTTPFlow):
        # Filter for Tuya/Smart Life domains
        if any(domain in flow.request.host for domain in [
            'tuya.com', 'iot.tuya.com', 'smartlife.com', 'tuyaus.com'
        ]):
            ctx.log.info(f"REQUEST: {flow.request.method} {flow.request.url}")
            if flow.request.content:
                ctx.log.info(f"BODY: {flow.request.content.decode('utf-8', errors='ignore')[:500]}")
    
    def response(self, flow: http.HTTPFlow):
        if any(domain in flow.request.host for domain in [
            'tuya.com', 'iot.tuya.com', 'smartlife.com', 'tuyaus.com'
        ]):
            ctx.log.info(f"RESPONSE: {flow.response.status_code} {flow.request.url}")
            if flow.response.content:
                ctx.log.info(f"BODY: {flow.response.content.decode('utf-8', errors='ignore')[:500]}")

addons = [SmartLifeLogger()]
```

### Run:
```bash
mitmproxy -s smartlife_logger.py --listen-port 8080
```

---

## Troubleshooting

### "No Internet" on Phone
- Check MITM proxy is running
- Verify firewall allows port 8080
- Check phone proxy settings

### "Certificate Error" in App
- App has SSL pinning
- Use Frida method to bypass
- Or try older app version

### Can't Capture HTTPS Traffic
- Install certificate as system certificate (requires root)
- Use Android emulator with root access
- Try iOS simulator instead

---

## Alternative: Use Android Emulator

```bash
# Install Android Studio with emulator
# Create emulator with Google Play (root access)
# Install Smart Life from Play Store
# Configure emulator proxy to host machine
# Install mitmproxy certificate as system cert
```

---

**Last Updated:** April 17, 2026
