#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMMON_SCRIPT="$SCRIPT_DIR/../remote/common_arch.sh"

if [ ! -f "$COMMON_SCRIPT" ]; then
    echo "Error: Cannot find common script at $COMMON_SCRIPT"
    exit 1
fi

source "$COMMON_SCRIPT"

# Preflight checks (connectivity, local binaries)
preflight_check

REMOTE_APK_PATH="$REMOTE_REPO_DIR/sensio_app/app/build/outputs/apk/debug/app-debug.apk"
LOCAL_STAGING_DIR="$SCRIPT_DIR/../../sensio_app/.remote-build"
LOCAL_APK_PATH="$LOCAL_STAGING_DIR/app-debug.apk"

log_info "Pulling built APK from remote host ($REMOTE_HOST)..."
start_time_pull=$(date +%s)

mkdir -p "$LOCAL_STAGING_DIR"
run_rsync artifact "$REMOTE_HOST:$REMOTE_APK_PATH" "$LOCAL_APK_PATH"

end_time_pull=$(date +%s)
duration_pull=$((end_time_pull - start_time_pull))

if [ ! -s "$LOCAL_APK_PATH" ]; then
    log_error "Pulled APK is empty or doesn't exist: $LOCAL_APK_PATH"
    exit 1
fi
log_info "APK successfully pulled to: sensio_app/.remote-build/app-debug.apk (${duration_pull}s)"

# Install locally via ADB
# Using the package name from deploy_remote_arch.sh
adb_install_apk "$LOCAL_APK_PATH" "com.example.whisperandroid"

log_info "Remote APK installation finished successfully."
