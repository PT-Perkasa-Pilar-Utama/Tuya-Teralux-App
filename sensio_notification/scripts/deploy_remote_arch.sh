#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMMON_SCRIPT="$SCRIPT_DIR/../../scripts/remote/common_arch.sh"

if [ ! -f "$COMMON_SCRIPT" ]; then
    echo "Error: Cannot find common script at $COMMON_SCRIPT"
    exit 1
fi

source "$COMMON_SCRIPT"

preflight_check
sync_source
sync_remote_configs "sensio_notification"

log_info "Starting remote lint and build for Sensio Notification on $REMOTE_HOST..."
start_time=$(date +%s)

ssh_exec "cd $REMOTE_REPO_DIR/sensio_notification && chmod +x gradlew && ./gradlew --no-daemon ktlintCheck clean :composeApp:assembleDebug"

end_time=$(date +%s)
duration=$((end_time - start_time))
log_info "Sensio Notification remote lint and build completed in ${duration}s."

REMOTE_APK_PATH="$REMOTE_REPO_DIR/sensio_notification/composeApp/build/outputs/apk/debug/composeApp-debug.apk"
LOCAL_STAGING_DIR="$SCRIPT_DIR/../.remote-build"
LOCAL_APK_PATH="$LOCAL_STAGING_DIR/composeApp-debug.apk"

log_info "Pulling built APK from remote host..."
start_time_pull=$(date +%s)

mkdir -p "$LOCAL_STAGING_DIR"
run_rsync artifact "$REMOTE_HOST:$REMOTE_APK_PATH" "$LOCAL_APK_PATH"

end_time_pull=$(date +%s)
duration_pull=$((end_time_pull - start_time_pull))

if [ ! -s "$LOCAL_APK_PATH" ]; then
    log_error "Pulled APK is empty or doesn't exist: $LOCAL_APK_PATH"
    exit 1
fi
log_info "APK successfully pulled to: sensio_notification/.remote-build/composeApp-debug.apk (${duration_pull}s)"

# Android specific post-step: local adb install
adb_install_apk "$LOCAL_APK_PATH" "com.sensio.notification"

log_info "Sensio Notification deploy-remote pipeline finished successfully."
