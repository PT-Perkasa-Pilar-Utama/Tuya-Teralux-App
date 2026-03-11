# RAG+Whisper Service Makefile

.PHONY: help install dev test clean

# Default target
help:
	@echo "RAG+Whisper Service - Available Commands:"
	@echo ""
	@echo "  make install    - Install dependencies in venv"
	@echo "  make dev        - Run development server with hot reload"
	@echo "  make test       - Run tests"
	@echo "  make clean      - Remove venv and cache files"
	@echo ""

# Install dependencies
install:
	@echo "📦 Installing dependencies..."
	./venv/bin/pip install --upgrade pip
	./venv/bin/pip install .
	@echo "✅ Dependencies installed"

# Run development server
dev:
	@echo "🚀 Starting RAG+Whisper Service..."
	./venv/bin/python -m uvicorn main:app --host 0.0.0.0 --port 8000 --reload

# Run tests
test:
	./venv/bin/pytest

# Clean artifacts
clean:
	rm -rf venv
	find . -type d -name "__pycache__" -exec rm -rf {} +
	find . -type f -name "*.pyc" -delete
