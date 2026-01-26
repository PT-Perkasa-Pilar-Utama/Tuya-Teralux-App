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
	@echo "  make kill           - Kill backend service (port 8080)"
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

# Run development mode for both services (using air)
# Note: This runs them in background and tails logs for both. 
# Better usage: open two terminals and run 'make dev-stt' and 'make dev-backend'
dev:
	@echo "🚀 Starting backend in development mode..."
	@echo "💡 Tip: Use separate terminals for other tasks."
	@(cd backend && $(MAKE) dev)

# Clean all
clean:
	@echo "🧹 Cleaning backend artifacts..."
	@$(MAKE) -C backend clean

# Kill all
kill:
	@echo "🔪 Killing backend service..."
	@$(MAKE) -C backend kill
