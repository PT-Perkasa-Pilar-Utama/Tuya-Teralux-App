# Sensio App - Root Makefile for Project-Wide Automation

.PHONY: help setup backend-setup dev dev-compose dev-server clean kill kill-compose kill-server test vet push-local adb-reverse

# Default target
help:
	@echo "Sensio App - Available Project-Wide Commands:"
	@echo ""
	@echo "  make setup          - Setup Backend (install tools, run migrations)"
	@echo "  make backend-setup  - Setup Backend (install tools, run migrations)"
	@echo "  make dev            - Run Backend in dev mode"
	@echo "  make clean          - Clean backend artifacts"
	@echo "  make kill           - Kill all backend and compose services"
	@echo "  make kill-server    - Kill only backend server and RAG processes"
	@echo "  make kill-compose   - Kill and cleanup Docker Compose services"
	@echo "  make adb-reverse    - Expose backend port (8081) to Android device via ADB"
	@echo ""

# Setup everything
setup: backend-setup
	@echo "✅ Full project setup complete!"

# Backend Setup
backend-setup:
	@echo "🚀 Setting up Backend..."
	@$(MAKE) -C backend setup

# Run development mode for the backend
# Use 'make dev' to run backend with hot-reload (Air)
dev:
	@echo "🚀 Starting backend in development mode..."
	@echo "💡 Tip: Use separate terminals for other tasks."
	@(cd backend && $(MAKE) dev)

# Start only compose services
dev-compose:
	@echo "🚀 Starting backend compose services in background..."
	@(cd backend && $(MAKE) dev-compose)

# Start only the backend server and RAG service
dev-server:
	@echo "🚀 Starting backend server..."
	@(cd backend && $(MAKE) dev-server)

# Clean all
clean:
	@echo "🧹 Cleaning backend artifacts..."
	@$(MAKE) -C backend clean

# Kill all
kill:
	@echo "🔪 Killing all project components..."
	@$(MAKE) -C backend kill

# Kill only server processes
kill-server:
	@echo "🔪 Killing backend server processes..."
	@$(MAKE) -C backend kill-server

# Kill and cleanup Docker Compose
kill-compose:
	@echo "🐳 Cleaning up Docker containers..."
	@$(MAKE) -C backend kill-compose

# Run tests
test:
	@echo "🧪 Running backend tests..."
	@$(MAKE) -C backend test

# Run vet
vet:
	@echo "🔍 Running backend go vet..."
	@$(MAKE) -C backend vet

# Push local
push-local:
	@echo "🚀 Pushing Docker image locally..."
	@$(MAKE) -C backend push-local TAG=$(TAG)

# Expose backend port to Android device via ADB
adb-reverse:
	@$(MAKE) -C backend adb-reverse
