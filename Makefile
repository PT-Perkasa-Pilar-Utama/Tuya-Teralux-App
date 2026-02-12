# Teralux App - Root Makefile for Project-Wide Automation

.PHONY: help setup backend-setup dev clean kill test vet

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

# Clean all
clean:
	@echo "🧹 Cleaning backend artifacts..."
	@$(MAKE) -C backend clean

# Kill all
kill:
	@echo "🔪 Killing backend service..."
	@$(MAKE) -C backend kill

# Run tests
test:
	@echo "🧪 Running backend tests..."
	@$(MAKE) -C backend test

# Run vet
vet:
	@echo "🔍 Running backend go vet..."
	@$(MAKE) -C backend vet
