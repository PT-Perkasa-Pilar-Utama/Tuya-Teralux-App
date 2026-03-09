#!/bin/bash

# Script to stream noise-free logs for the Sensio app.
# Filters for app-specific tags and critical system errors.

PACKAGE_NAME="com.example.whisperandroid"

# Try to find the PID
PID=$(adb shell pidof $PACKAGE_NAME)

if [ -z "$PID" ]; then
    echo "Warning: App $PACKAGE_NAME is not running. Showing global logs with filters."
    # Filter by tags if PID not found
    adb logcat -v brief SensioService:D SensioOverlay:D SensioBGAssistantCoord:D SensioWakeWord:I AndroidRuntime:E *:S
else
    echo "Filtering logs for PID $PID ($PACKAGE_NAME)..."
    # Filter by PID and tags
    adb logcat --pid=$PID -v brief SensioService:D SensioOverlay:D SensioBGAssistantCoord:D SensioWakeWord:I AndroidRuntime:E *:S
fi
