#!/bin/bash
set -e

# Source common functions
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMMON_SCRIPT="$SCRIPT_DIR/../../scripts/remote/common_arch.sh"

if [ ! -f "$COMMON_SCRIPT" ]; then
    echo "Error: Cannot find common script at $COMMON_SCRIPT"
    exit 1
fi

source "$COMMON_SCRIPT"

preflight_check
sync_source_delta

log_info "Starting remote lint and build for Backend on $REMOTE_HOST..."
start_time=$(date +%s)

ssh "$REMOTE_HOST" "bash -lc 'cd $REMOTE_REPO_DIR/backend && make lint-strict && make vet && make build'"

end_time=$(date +%s)
duration=$((end_time - start_time))
log_info "Backend remote lint and build completed in ${duration}s."
log_info "Remote binary is located at: $REMOTE_HOST:$REMOTE_REPO_DIR/backend/main"
