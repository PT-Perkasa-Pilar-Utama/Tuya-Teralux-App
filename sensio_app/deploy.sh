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

# Check if ADB is available and prefer SDK version
if [ -f "/opt/android-sdk/platform-tools/adb" ]; then
    export PATH="/opt/android-sdk/platform-tools:$PATH"
fi

if ! command -v adb &> /dev/null; then
    echo "❌ ADB not found. Please install Android SDK platform tools."
    exit 1
fi

# Get connected devices
echo "📱 Checking connected devices..."
DEVICES=$(adb devices | grep -v "List" | grep "device$" | awk '{print $1}')
DEVICE_COUNT=$(echo "$DEVICES" | grep -v "^$" | wc -l | xargs)

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

# Run Linter
echo "🔍 Running Linter..."
chmod +x gradlew
set +e
./gradlew ktlintCheck --quiet
LINT_STATUS=$?
set -e

if [ $LINT_STATUS -ne 0 ]; then
    echo "⚠️  Lint issues found. Attempting to auto-fix..."
    set +e
    ./gradlew ktlintFormat --quiet
    LINT_FIX_STATUS=$?
    set -e
    
    if [ $LINT_FIX_STATUS -eq 0 ]; then
         echo "✅ Lint issues auto-fixed!"
    else
         echo "❌ Lint failed even after auto-fix. Please check manually."
         exit 1
    fi
else
    echo "✅ Code style is correct."
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
set +e
# Added -t flag for debug builds
INSTALL_OUTPUT=$(adb -s "$DEVICE_ID" install -t -r -g "$APK_PATH" 2>&1)
INSTALL_STATUS=$?
set -e

if [ $INSTALL_STATUS -ne 0 ]; then
    if echo "$INSTALL_OUTPUT" | grep -q "INSTALL_FAILED_UPDATE_INCOMPATIBLE"; then
        echo "⚠️  Signature mismatch detected. Uninstalling existing app..."
        set +e
        adb -s "$DEVICE_ID" uninstall "$PACKAGE_NAME"
        echo "🔄 Retrying installation..."
        adb -s "$DEVICE_ID" install -r "$APK_PATH"
        INSTALL_STATUS=$?
        set -e
    else
        echo "❌ Installation failed:"
        echo "$INSTALL_OUTPUT"
        exit 1
    fi
fi

if [ $INSTALL_STATUS -eq 0 ]; then
    echo "✅ Installation successful!"
else
    echo "❌ Installation failed"
    exit 1
fi

# echo ""
# echo "📝 Showing logs (Ctrl+C to stop)..."
# adb -s "$DEVICE_ID" logcat -v daily -s "SensioWakeWord:D" "MainActivity:D" "*:E"
