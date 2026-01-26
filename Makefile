# Teralux App - Root Makefile for Project-Wide Automation

.PHONY: help setup stt-setup ollama-setup backend-setup dev clean kill

# Default target
help:
	@echo "Teralux App - Available Project-Wide Commands:"
	@echo ""
	@echo "  make setup          - Setup EVERYTHING (STT, Ollama, Backend)"
	@echo "  make stt-setup      - Setup STT service (build whisper.cpp, models, etc)"
	@echo "  make ollama-setup   - Setup Ollama container and pull required models"
	@echo "  make backend-setup  - Setup Backend (install tools, run migrations)"
	@echo "  make dev            - Run both Backend and STT service in dev mode"
	@echo "  make clean          - Clean all build artifacts"
	@echo "  make kill           - Kill processes on backend (8080) and STT (8081) ports"
	@echo ""

# Setup everything
setup: stt-setup ollama-setup backend-setup
	@echo "✅ Full project setup complete!"

# STT Service Setup
stt-setup:
	@echo "🚀 Setting up STT Service..."
	@$(MAKE) -C stt-service setup

# Ollama Setup
ollama-setup:
	@echo "🚀 Setting up Ollama..."
	@$(MAKE) -C stt-service ollama-setup

# Backend Setup
backend-setup:
	@echo "🚀 Setting up Backend..."
	@cd backend && $(MAKE) install-watch
	@cd backend && $(MAKE) install-swagger
	@cd backend && $(MAKE) migrate-up

# Run development mode for both services (using air)
# Note: This runs them in background and tails logs for both. 
# Better usage: open two terminals and run 'make dev-stt' and 'make dev-backend'
dev:
	@echo "🚀 Starting both services in development mode..."
	@echo "💡 Tip: Logs will be interleaved. Use separate terminals for better control."
	@((cd stt-service && $(MAKE) dev) & (cd backend && $(MAKE) dev) & wait)

# Clean all
clean:
	@echo "🧹 Cleaning all artifacts..."
	@$(MAKE) -C stt-service clean
	@$(MAKE) -C backend clean

# Kill all
kill:
	@echo "🔪 Killing all services..."
	@$(MAKE) -C stt-service kill
	@$(MAKE) -C backend kill
