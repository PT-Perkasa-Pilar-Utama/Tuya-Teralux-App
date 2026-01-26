#!/bin/bash

# Exit on error
set -e

# Make script behave the same regardless of current working directory by
# switching into the script's directory (backend/) first.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "🚀 Starting Backend STT Setup... (whisper.cpp build + tools)"

# 1. Install system dependencies if missing
if ! command -v cmake > /dev/null 2>&1; then
    echo "📦 Installing cmake..."
    if [ -f /etc/arch-release ]; then
        sudo pacman -S --noconfirm cmake
    elif [ -f /etc/debian_version ]; then
        apt-get update && apt-get install -y cmake
    fi
fi

if ! command -v ffmpeg > /dev/null 2>&1; then
    echo "📦 Installing ffmpeg..."
    if [ -f /etc/arch-release ]; then
        sudo pacman -S --noconfirm ffmpeg
    elif [ -f /etc/debian_version ]; then
        apt-get update && apt-get install -y ffmpeg
    fi
fi

# 2. Build whisper.cpp
if [ -d "whisper.cpp" ]; then
    echo "🛠️ Building whisper.cpp..."
    cd whisper.cpp
    cmake -B build -DBUILD_SHARED_LIBS=OFF
    cmake --build build --config Release -j$(nproc)
    cd ..
else
    echo "⚠️ whisper.cpp not found in repository root. If you run inside Docker, ensure whisper.cpp exists at /app/whisper.cpp"
fi

# 3. Setup bin directory
echo "📂 Setting up bin directory..."
mkdir -p bin

# Find whisper-cli executable (it might be in build/bin or just build/)
CLI_PATH=$(find whisper.cpp -name "whisper-cli" -type f | head -n 1)

if [ -n "$CLI_PATH" ]; then
    echo "Found whisper-cli at: $CLI_PATH"
    cp "$CLI_PATH" bin/ || true
else
    echo "ℹ️ whisper-cli not found (build may be skipped). Continuing..."
fi

# 4. Download model (only if missing locally)
if [ ! -f "bin/ggml-base.bin" ]; then
    if [ -d "whisper.cpp" ]; then
        echo "📥 Downloading Whisper base model..."
        ./whisper.cpp/models/download-ggml-model.sh base || true

        if [ -f "ggml-base.bin" ]; then
            mv ggml-base.bin bin/ || true
        elif [ -f "whisper.cpp/models/ggml-base.bin" ]; then
            mv whisper.cpp/models/ggml-base.bin bin/ || true
        else
            echo "⚠️ Model file not found after download step; it's optional for setup"
        fi
    fi
fi

# 5. Build Go Service
echo "🏗️ Building Backend (to ensure Go modules work)..."
# Build from the script directory (backend/) so this works whether the
# script is run from repo root or from inside backend/ directly.
if [ -f "main.go" ] || [ -f "go.mod" ]; then
    go build -o main . || true
else
    echo "ℹ️ No Go files found to build in $SCRIPT_DIR"
fi

echo "✅ Setup complete!"