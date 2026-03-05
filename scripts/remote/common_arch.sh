#!/bin/bash
# Shared helper for remote lint and build

# Hard defaults, overrideable via env
REMOTE_HOST=${REMOTE_HOST:-"arch"}
REMOTE_REPO_DIR=${REMOTE_REPO_DIR:-"~/Documents/Tuya-Teralux-App"}

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

SYNC_PROGRESS=${SYNC_PROGRESS:-1}
SYNC_STATS=${SYNC_STATS:-1}

log_info() { echo -e "${GREEN}[INFO]${NC} $1" >&2; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1" >&2; }
log_error() { echo -e "${RED}[ERROR]${NC} $1" >&2; }

run_rsync() {
    local mode="$1"
    shift
    
    local rsync_opts=("-az")

    if [ "$mode" = "source" ]; then
        if [ "$SYNC_PROGRESS" = "1" ]; then
            rsync_opts+=("--info=progress2" "--human-readable" "--partial")
        fi
        if [ "$SYNC_STATS" = "1" ]; then
            rsync_opts+=("--stats")
        fi
    fi
    
    # Custom SSH with robust transport & keepalives
    rsync_opts+=("-e" "ssh -o ConnectTimeout=30 -o ServerAliveInterval=15 -o ServerAliveCountMax=6")

    # Capture original SIGINT handler
    local orig_trap=$(trap -p INT)
    
    # Set controlled exit trap
    trap 'log_warn "Sync dibatalkan oleh user."; eval "$orig_trap"; exit 130' SIGINT

    rsync "${rsync_opts[@]}" "$@" || {
        local exit_code=$?
        
        # Restore trap early upon internal crash to not muddy outer scope
        eval "$orig_trap"
        
        if [ $exit_code -eq 255 ]; then
            log_error "Sync failed (Exit Code 255): SSH connection dropped or transport timed out."
        elif [ $exit_code -eq 130 ]; then
             log_warn "Sync dibatalkan oleh user."
        else
            log_error "rsync failed with exit code $exit_code"
        fi
        exit $exit_code
    }
    
    eval "$orig_trap"
}

preflight_check() {
    log_info "Running preflight checks..."
    
    # Check local binaries
    for cmd in ssh rsync bash; do
        if ! command -v $cmd > /dev/null 2>&1; then
            log_error "Required local command not found: $cmd"
            exit 1
        fi
    done
    
    # Check remote connectivity
    if ! ssh -q "$REMOTE_HOST" 'echo ok' > /dev/null 2>&1; then
        log_error "Cannot connect to remote host: $REMOTE_HOST"
        log_error "Please ensure SSH access is configured and host is reachable."
        exit 1
    fi
    log_info "Preflight checks passed."
}

resolve_repo_root() {
    local root=$(git rev-parse --show-toplevel 2>/dev/null)
    if [ -z "$root" ]; then
        log_error "Must be run from within a git repository."
        exit 1
    fi
    echo "$root"
}

ssh_exec() {
    local cmd="$1"
    # Execute the command by piping it to bash on the remote host
    # This prevents complex quoting/injection issues with single quotes inside the command string
    if ! printf '%s\n' "$cmd" | ssh "$REMOTE_HOST" "bash -se"; then
        log_error "Remote execution failed for: $cmd"
        exit 1
    fi
}

adb_install_apk() {
    local apk_path="$1"
    local package_name="$2"
    
    log_info "Preparing to install APK via ADB..."
    
    if ! command -v adb > /dev/null 2>&1; then
        log_error "ADB not found. Please install Android platform tools."
        exit 1
    fi

    local selected_device=""
    
    if [ -n "${DEVICE_ID:-}" ]; then
        selected_device="$DEVICE_ID"
        log_info "Using specified device: $selected_device"
    else
        # Count connected devices (excluding "List of devices attached" line and empty lines)
        local device_count=$(adb devices | grep -v "List of devices" | grep "device$" | wc -l)
        if [ "$device_count" -eq 0 ]; then
            log_error "No Android devices connected via ADB."
            exit 1
        elif [ "$device_count" -gt 1 ]; then
            log_error "Multiple Android devices connected."
            log_error "Please specify DEVICE_ID=<adb_serial>."
            adb devices
            exit 1
        else
            selected_device=$(adb devices | grep -v "List of devices" | grep "device$" | awk '{print $1}')
            log_info "Auto-selected single connected device: $selected_device"
        fi
    fi

    log_info "Installing $apk_path to device $selected_device..."
    local install_output
    if ! install_output=$(adb -s "$selected_device" install -r "$apk_path" 2>&1); then
        if echo "$install_output" | grep -q "INSTALL_FAILED_UPDATE_INCOMPATIBLE"; then
            log_warn "Installation failed due to signature incompatibility. Attempting uninstall..."
            adb -s "$selected_device" uninstall "$package_name" || true
            log_info "Retrying installation..."
            if ! adb -s "$selected_device" install -r "$apk_path"; then
                log_error "Re-installation failed."
                exit 1
            fi
        else
            log_error "APK installation failed: $install_output"
            exit 1
        fi
    fi
    log_info "APK successfully installed."
}

detect_remote_sdk_dir() {
    log_info "Detecting remote Android SDK directory..."
    local sdk_dir=""
    
    # Check common paths
    if ssh -q "$REMOTE_HOST" 'test -d ~/Android/Sdk'; then
        sdk_dir="~/Android/Sdk"
    elif ssh -q "$REMOTE_HOST" 'test -d /opt/android-sdk'; then
        sdk_dir="/opt/android-sdk"
    else
        log_error "Could not find Android SDK on remote host ($REMOTE_HOST)."
        log_error "Checked: ~/Android/Sdk, /opt/android-sdk"
        exit 1
    fi
    
    # Resolve to absolute path
    sdk_dir=$(ssh -q "$REMOTE_HOST" "bash -lc 'readlink -f $sdk_dir'")
    log_info "Found remote SDK at: $sdk_dir"
    echo "$sdk_dir"
}

sync_remote_configs() {
    local module_name="$1"
    local repo_root=$(resolve_repo_root)
    
    log_info "Syncing configs for module: $module_name"
    
    if [ "$module_name" = "backend" ]; then
        if [ -f "$repo_root/backend/.env" ]; then
            log_info "Copying backend/.env to remote host..."
            run_rsync config "$repo_root/backend/.env" "$REMOTE_HOST:$REMOTE_REPO_DIR/backend/.env"
            ssh -q "$REMOTE_HOST" "chmod 600 $REMOTE_REPO_DIR/backend/.env"
        else
            log_warn "Local backend/.env not found, skipping sync."
        fi
        
    elif [ "$module_name" = "sensio_app" ] || [ "$module_name" = "sensio_notification" ]; then
        local remote_sdk=$(detect_remote_sdk_dir)
        local local_props_src="$repo_root/sensio_app/local.properties"
        local remote_props_dest="$REMOTE_REPO_DIR/$module_name/local.properties"
        
        if [ ! -f "$local_props_src" ]; then
            log_error "Local config not found: $local_props_src"
            exit 1
        fi
        
        log_info "Generating $module_name/local.properties for remote host..."
        
        # Create a temporary local.properties with rewritten sdk.dir
        local tmp_props=$(mktemp)
        
        # Copy original but filter out sdk.dir
        grep -v "^sdk.dir=" "$local_props_src" > "$tmp_props" || true
        
        # Append remote sdk.dir
        echo "sdk.dir=$remote_sdk" >> "$tmp_props"
        
        # Sync to remote
        run_rsync config "$tmp_props" "$REMOTE_HOST:$remote_props_dest"
        ssh -q "$REMOTE_HOST" "chmod 600 $remote_props_dest"
        
        rm -f "$tmp_props"
        log_info "Config synced to $remote_props_dest successfully."
    else
        log_error "Unknown module for config sync: $module_name"
        exit 1
    fi
}

sync_source() {
    log_info "Syncing source code to $REMOTE_HOST:$REMOTE_REPO_DIR..."
    
    local start_time=$(date +%s)
    
    local repo_root=$(resolve_repo_root)
    
    # Run rsync incrementally
    run_rsync source --delete \
        --exclude=".git/" \
        --exclude=".idea/" \
        --exclude=".vscode/" \
        --exclude=".gradle/" \
        --exclude="**/build/" \
        --exclude="**/.kotlin/" \
        --exclude="**/*.apk" \
        --exclude="**/*.aab" \
        --exclude="**/*.keystore" \
        --exclude="**/*.jks" \
        --exclude="**/*.p12" \
        --exclude="**/*.pem" \
        --exclude="backend/.env" \
        --exclude="sensio_app/local.properties" \
        --exclude="sensio_notification/local.properties" \
        --exclude=".DS_Store" \
        "$repo_root/" "$REMOTE_HOST:$REMOTE_REPO_DIR/"
        
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    log_info "Sync completed in ${duration}s."
}
