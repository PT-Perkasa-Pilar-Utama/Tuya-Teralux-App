#!/bin/bash

# Build and Deploy Whisper Android App via ADB
# Usage: ./deploy.sh [device-id (optional)]

set -e

DEVICE_ID=$1
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APK_PATH="$PROJECT_DIR/app/build/outputs/apk/debug/app-debug.apk"
PACKAGE_NAME="com.example.whisper_android"
LAUNCHER_ACTIVITY=".MainActivity"

echo "=== Whisper Android Build & Deploy ==="
echo ""

# Check if ADB is available
if ! command -v adb &> /dev/null; then
    echo "❌ ADB not found. Please install Android SDK platform tools."
    exit 1
fi

# Get connected devices
echo "📱 Checking connected devices..."
DEVICES=$(adb devices | grep -v "List" | grep "device$" | awk '{print $1}')
DEVICE_COUNT=$(echo "$DEVICES" | grep -c . || echo 0)

if [ "$DEVICE_COUNT" -eq 0 ]; then
    echo "❌ No ADB devices connected."
    exit 1
fi

# Determine target device
if [ -z "$DEVICE_ID" ]; then
    if [ "$DEVICE_COUNT" -eq 1 ]; then
        DEVICE_ID=$(echo "$DEVICES" | head -n 1)
        echo "✅ Found 1 device: $DEVICE_ID"
    else
        echo "Multiple devices found:"
        echo "$DEVICES"
        echo ""
        read -p "Enter device ID to deploy to: " DEVICE_ID
    fi
fi

echo ""
echo "🔨 Building APK (clean build)..."
cd "$PROJECT_DIR"
./gradlew clean assembleDebug --quiet

if [ ! -f "$APK_PATH" ]; then
    echo "❌ APK not found at $APK_PATH"
    exit 1
fi

echo "✅ Build complete: $APK_PATH"
echo ""

echo "📲 Installing APK on $DEVICE_ID..."
adb -s "$DEVICE_ID" install -r "$APK_PATH"

if [ $? -eq 0 ]; then
    echo "✅ Installation successful!"
    echo ""
    echo "🚀 Launching app..."
    adb -s "$DEVICE_ID" shell am start -n "$PACKAGE_NAME/$LAUNCHER_ACTIVITY"
    echo "✅ App launched!"
else
    echo "❌ Installation failed"
    exit 1
fi

echo ""
echo "📝 Showing logs (Ctrl+C to stop)..."
adb -s "$DEVICE_ID" logcat -v daily -s "SensioWakeWord:D" "MainActivity:D" "*:E"
