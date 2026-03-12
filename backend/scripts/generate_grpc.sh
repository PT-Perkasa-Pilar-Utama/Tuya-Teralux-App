#!/bin/bash
# generate_grpc.sh - Regenerates gRPC code for the RAG service

set -e

SERVICE_DIR="backend/services/rag-whisper-service"
PROTO_DIR="$SERVICE_DIR/proto"
OUT_DIR="$SERVICE_DIR/src/modules/whisper/interfaces/grpc"

echo "🔨 Generating gRPC code from $PROTO_DIR/whisper.proto..."

# Ensure we are in the project root
if [ ! -d "$SERVICE_DIR" ]; then
    echo "❌ Error: Run this script from the project root."
    exit 1
fi

# Activate venv if it exists
if [ -f "$SERVICE_DIR/venv/bin/activate" ]; then
    source "$SERVICE_DIR/venv/bin/activate"
fi

# Install grpcio-tools if not present
pip install grpcio-tools

# Run protoc
python3 -m grpc_tools.protoc \
    -I"$PROTO_DIR" \
    --python_out="$OUT_DIR" \
    --grpc_python_out="$OUT_DIR" \
    "$PROTO_DIR/whisper.proto"

# Fix absolute imports in generated files
echo "🔧 Fixing imports in generated files..."
sed -i 's/^import whisper_pb2 as whisper__pb2/from . import whisper_pb2 as whisper__pb2/' "$OUT_DIR/whisper_pb2_grpc.py"

echo "✅ gRPC code generated and fixed successfully."
