#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMMON_SCRIPT="$SCRIPT_DIR/../../scripts/remote/common_arch.sh"

if [ ! -f "$COMMON_SCRIPT" ]; then
    echo "Error: Cannot find common script at $COMMON_SCRIPT"
    exit 1
fi

source "$COMMON_SCRIPT"

REMOTE_BINARY_NAME=${REMOTE_BINARY_NAME:-"main"}

preflight_check
sync_source
sync_remote_configs "backend"

log_info "Starting remote lint and build for Backend on $REMOTE_HOST..."
start_time=$(date +%s)

ssh_exec "cd $REMOTE_REPO_DIR/backend && make lint-strict && make vet && make build"

end_time=$(date +%s)
duration=$((end_time - start_time))
log_info "Backend remote lint and build completed in ${duration}s."

REMOTE_BINARY_PATH="$REMOTE_REPO_DIR/backend/main"
LOCAL_STAGING_DIR="$SCRIPT_DIR/../.remote-build"
LOCAL_BINARY_PATH="$LOCAL_STAGING_DIR/$REMOTE_BINARY_NAME"

log_info "Pulling built binary from remote host..."
start_time_pull=$(date +%s)

mkdir -p "$LOCAL_STAGING_DIR"
run_rsync artifact "$REMOTE_HOST:$REMOTE_BINARY_PATH" "$LOCAL_BINARY_PATH"

end_time_pull=$(date +%s)
duration_pull=$((end_time_pull - start_time_pull))

if [ ! -s "$LOCAL_BINARY_PATH" ]; then
    log_error "Pulled binary is empty or doesn't exist: $LOCAL_BINARY_PATH"
    exit 1
fi

chmod +x "$LOCAL_BINARY_PATH"

log_info "Binary successfully pulled to: backend/.remote-build/$REMOTE_BINARY_NAME (${duration_pull}s)"
