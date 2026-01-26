#!/bin/bash

# Exit on error
set -e

echo "🚀 Starting STT Service Setup..."

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
echo "🛠️ Building whisper.cpp..."
cd whisper.cpp
cmake -B build -DBUILD_SHARED_LIBS=OFF
cmake --build build --config Release -j$(nproc)
cd ..

# 3. Setup bin directory
echo "📂 Setting up bin directory..."
mkdir -p bin

# Find whisper-cli executable (it might be in build/bin or just build/)
CLI_PATH=$(find whisper.cpp -name "whisper-cli" -type f | head -n 1)

if [ -z "$CLI_PATH" ]; then
    echo "❌ Error: whisper-cli binary not found after build"
    exit 1
fi

echo "Found whisper-cli at: $CLI_PATH"
cp "$CLI_PATH" bin/

# 4. Download model
if [ ! -f "bin/ggml-base.bin" ]; then
    echo "📥 Downloading Whisper base model..."
    ./whisper.cpp/models/download-ggml-model.sh base
    
    # Handle different download locations (root vs models dir)
    if [ -f "ggml-base.bin" ]; then
        mv ggml-base.bin bin/
    elif [ -f "whisper.cpp/models/ggml-base.bin" ]; then
        mv whisper.cpp/models/ggml-base.bin bin/
    else
        echo "❌ Error: Could not find downloaded model file"
        exit 1
    fi
fi

# 5. Build Go Service
echo "🏗️ Building STT Go Service..."
go build -o stt-app main.go

echo "✅ Setup complete! Run the service with: ./stt-app"
