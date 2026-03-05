#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMMON_SCRIPT="$SCRIPT_DIR/../../scripts/remote/common_arch.sh"

if [ ! -f "$COMMON_SCRIPT" ]; then
    echo "Error: Cannot find common script at $COMMON_SCRIPT"
    exit 1
fi

source "$COMMON_SCRIPT"

# Preflight check against the basic binaries and SSH connection
preflight_check

# Default values if environment variables are not set
REGISTRY=${REGISTRY:-ghcr.io}
USERNAME=${USERNAME:-farismnrr}
IMAGE_BASE=${IMAGE_BASE:-sensio-backend}
IMAGE_TAG=${IMAGE_TAG:-latest}
PLATFORMS=${PLATFORMS:-linux/amd64,linux/arm64}

log_info "Synchronizing backend source to remote host..."
start_time_sync=$(date +%s)

repo_root=$(resolve_repo_root)

# Backend-only incremental rsync
run_rsync source --delete \
    --exclude=".git/" \
    --exclude=".idea/" \
    --exclude=".vscode/" \
    --exclude=".air/" \
    --exclude="tmp/" \
    --exclude="uploads/" \
    --exclude="models/" \
    --exclude="*.log" \
    --exclude="main" \
    --exclude="bin/" \
    --exclude="build/" \
    --exclude=".env" \
    "$repo_root/backend/" "$REMOTE_HOST:$REMOTE_REPO_DIR/backend/"

end_time_sync=$(date +%s)
duration_sync=$((end_time_sync - start_time_sync))
log_info "Backend sync completed in ${duration_sync}s."

# Check remote binaries explicitly required for pushing
log_info "Verifying remote Docker capabilities..."
ssh_exec "command -v docker >/dev/null 2>&1 || { echo 'Docker not found on remote.'; exit 1; }"
ssh_exec "docker buildx version >/dev/null 2>&1 || { echo 'Docker buildx not available on remote.'; exit 1; }"

# Format the command payload with variable pass-through
cmd_payload="cd $REMOTE_REPO_DIR/backend && \
    make lint && \
    make vet && \
    make build && \
    make push REGISTRY=$REGISTRY USERNAME=$USERNAME IMAGE_BASE=$IMAGE_BASE IMAGE_TAG=$IMAGE_TAG PLATFORMS=$PLATFORMS"

log_info "Executing remote lint, build, & push on $REMOTE_HOST..."
start_time_push=$(date +%s)

# Execute
ssh_exec "$cmd_payload"

end_time_push=$(date +%s)
duration_push=$((end_time_push - start_time_push))

log_info "Backend push-remote pipeline completed in ${duration_push}s."
log_info "Target Image: $REGISTRY/$USERNAME/$IMAGE_BASE:$IMAGE_TAG"
