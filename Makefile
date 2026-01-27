# Teralux App - Root Makefile for Project-Wide Automation

.PHONY: help setup backend-setup dev clean kill

# Default target
help:
	@echo "Teralux App - Available Project-Wide Commands:"
	@echo ""
	@echo "  make setup          - Setup Backend (install tools, run migrations)"
	@echo "  make backend-setup  - Setup Backend (install tools, run migrations)"
	@echo "  make dev            - Run Backend in dev mode"
	@echo "  make clean          - Clean backend artifacts"
	@echo "  make kill           - Kill backend service (port 8081)"
	@echo ""

# Setup everything
setup: backend-setup
	@echo "✅ Full project setup complete!"

# Backend Setup
backend-setup:
	@echo "🚀 Setting up Backend..."
	@cd backend && $(MAKE) install-watch
	@cd backend && $(MAKE) install-swagger
	@cd backend && $(MAKE) migrate-up

# Run development mode for the backend
# Use 'make dev' to run backend with hot-reload (Air)
dev:
	@echo "🚀 Starting backend in development mode..."
	@echo "💡 Tip: Use separate terminals for other tasks."
	@(cd backend && $(MAKE) dev)

# Docker-based development helpers (forwarded to backend)
dev-docker:
	@echo "🚀 Starting backend dev environment in Docker (forwarded)"
	@(cd backend && $(MAKE) dev-docker)

dev-docker-build:
	@echo "🚀 Starting backend dev environment in Docker with build (forwarded)"
	@(cd backend && $(MAKE) dev-docker-build)

dev-docker-stop:
	@echo "🛑 Stopping backend dev Docker environment (forwarded)"
	@(cd backend && $(MAKE) dev-docker-stop)

# Run setup script for whisper/whisper.cpp (forwarded to backend)
setup-stt:
	@echo "🚀 Running whisper/setup script in backend (forwarded)"
	@(cd backend && $(MAKE) setup)

# Ollama helper (forwarded to backend)
ollama-setup:
	@echo "🔧 Running Ollama setup (forwarded to backend)"
	@(cd backend && $(MAKE) ollama-setup)

# Clean all
clean:
	@echo "🧹 Cleaning backend artifacts..."
	@$(MAKE) -C backend clean

# Kill all
kill:
	@echo "🔪 Killing backend service..."
	@$(MAKE) -C backend kill
