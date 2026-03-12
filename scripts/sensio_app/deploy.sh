#!/bin/bash

# Build and Deploy Whisper Android App via ADB
# Usage: ./deploy.sh [options] [device-id]
# Options:
#   --build-only    Only build the APK
#   --install-only  Only install the existing APK
#   --no-build      Skip building, just install

set -e

# Default behavior: build and install
DO_BUILD=true
DO_INSTALL=true
DEVICE_ID=""

# Source common helpers for interactive_select and colors
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMMON_SCRIPT="$SCRIPT_DIR/../remote/common_arch.sh"

if [ -f "$COMMON_SCRIPT" ]; then
    # We only want colors and interactive_select, not the whole preflight/remote logic
    # But sourcing it is the easiest way to get those functions.
    # We might need to be careful if it defines global variables or runs code.
    # common_arch.sh mostly defines functions except for some variable defaults.
    source "$COMMON_SCRIPT"
else
    # Fallback basic colors if common script not found
    GREEN='\033[0;32m'
    RED='\033[0;31m'
    YELLOW='\033[1;33m'
    NC='\033[0m'
    log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
    log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
fi

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --build-only)
            DO_INSTALL=false
            shift
            ;;
        --install-only)
            DO_BUILD=false
            shift
            ;;
        --no-build)
            DO_BUILD=false
            shift
            ;;
        -*)
            echo "Unknown option: $1"
            exit 1
            ;;
        *)
            DEVICE_ID="$1"
            shift
            ;;
    esac
done

PROJECT_DIR="$(cd "$SCRIPT_DIR/../../sensio_app" && pwd)"
# Check if we have a locally built APK or a remote-pulled one
APK_PATH="$PROJECT_DIR/app/build/outputs/apk/debug/app-debug.apk"
if [ ! -f "$APK_PATH" ] && [ -f "$PROJECT_DIR/.remote-build/app-debug.apk" ]; then
    APK_PATH="$PROJECT_DIR/.remote-build/app-debug.apk"
fi

PACKAGE_NAME="com.example.whisperandroid"

if [ "$DO_BUILD" = true ]; then
    echo "=== Whisper Android Build ==="
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
        set -e
    fi

    echo "🔨 Building APK (clean build)..."
    ./gradlew clean assembleDebug --quiet
    echo "✅ Build complete: $APK_PATH"
fi

if [ "$DO_INSTALL" = true ]; then
    echo "=== Whisper Android Deploy ==="
    
    if [ ! -f "$APK_PATH" ]; then
        echo "❌ APK not found at $APK_PATH. Please build first."
        exit 1
    fi

    # ADB Setup
    if [ -f "/opt/android-sdk/platform-tools/adb" ]; then
        export PATH="/opt/android-sdk/platform-tools:$PATH"
    fi

    if ! command -v adb &> /dev/null; then
        echo "❌ ADB not found. Please install Android SDK platform tools."
        exit 1
    fi

    # Device selection
    if [ -z "$DEVICE_ID" ]; then
        DEVICES=$(adb devices | grep -v "List" | grep "device$" | awk '{print $1}')
        DEVICE_COUNT=$(echo "$DEVICES" | grep -v "^$" | wc -l | xargs)

        if [ "$DEVICE_COUNT" -eq 0 ]; then
            echo "❌ No ADB devices connected."
            exit 1
        elif [ "$DEVICE_COUNT" -eq 1 ]; then
            DEVICE_ID=$(echo "$DEVICES" | head -n 1)
            echo "✅ Auto-selected device: $DEVICE_ID"
        else
            if command -v interactive_select &> /dev/null; then
                IFS=$'\n' read -rd '' -a device_arr <<<"$DEVICES" || true
                DEVICE_ID=$(interactive_select "Select device to deploy to:" "${device_arr[@]}")
            else
                echo "Multiple devices found:"
                echo "$DEVICES"
                read -p "Enter device ID: " DEVICE_ID
            fi
        fi
    fi

    echo "📲 Installing APK on $DEVICE_ID..."
    set +e
    INSTALL_OUTPUT=$(adb -s "$DEVICE_ID" install -t -r -g "$APK_PATH" 2>&1)
    INSTALL_STATUS=$?
    set -e

    if [ $INSTALL_STATUS -ne 0 ]; then
        if echo "$INSTALL_OUTPUT" | grep -q "INSTALL_FAILED_UPDATE_INCOMPATIBLE"; then
            echo "⚠️  Signature mismatch. Uninstalling existing app..."
            adb -s "$DEVICE_ID" uninstall "$PACKAGE_NAME" || true
            echo "🔄 Retrying installation..."
            adb -s "$DEVICE_ID" install -t -r -g "$APK_PATH"
            INSTALL_STATUS=$?
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
fi
