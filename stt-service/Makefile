# Teralux STT Service - Makefile for Development Automation

.PHONY: help setup dev start build tidy lint swagger clean kill

# Default target
help:
	@echo "Teralux STT Service - Available Commands:"
	@echo ""
	@echo "  make setup     - Run the setup script (install deps, build whisper.cpp)"
	@echo "  make dev       - Run development server with hot reload (Air)"
	@echo "  make start     - Run development server without hot reload"
	@echo "  make build     - Build release binary"
	@echo "  make dev-docker       - Run development environment in Docker"
	@echo "  make dev-docker-build - Run development environment in Docker (Rebuild)"
	@echo "  make dev-docker-stop  - Stop development environment"
	@echo "  make tidy      - Run go mod tidy"
	@echo "  make lint      - Run go fmt and go vet"
	@echo "  make swagger   - Generate Swagger documentation"
	@echo "  make clean     - Clean build artifacts and temp files"
	@echo "  make kill      - Kill process running on PORT 8081"
	@echo ""

# Run setup script
setup:
	@echo "🚀 Running setup script..."
	@./setup.sh

# Run development server with hot reload (requires air)
dev:
	@echo "🚀 Starting development server with hot reload..."
	@$(shell go env GOPATH)/bin/air || (echo "❌ air not found. Installing..." && go install github.com/air-verse/air@latest && $(shell go env GOPATH)/bin/air)

# Run in Docker (dev mode with hot reload)
dev-docker: dev-docker-stop
	@echo "🚀 Starting development environment in Docker..."
	docker compose -f docker-compose.dev.yml up

# Run development environment with forced rebuild
dev-docker-build:
	@echo "🚀 Starting development environment in Docker (Forced Rebuild)..."
	docker compose -f docker-compose.dev.yml up --build

# Stop development environment
dev-docker-stop:
	@echo "🛑 Stopping development environment..."
	docker compose -f docker-compose.dev.yml down

# Run development server without hot reload
start:
	@echo "🚀 Starting development server (no hot reload)..."
	go run main.go

# Build release binary
build:
	@echo "🔨 Building release binary..."
	go build -o stt-app main.go

# Run go mod tidy
tidy:
	@echo "🧹 Tidying up Go modules..."
	go mod tidy

# Run linting
lint:
	@echo "🔍 Running fmt and vet..."
	go fmt ./...
	go vet ./...

# Generate Swagger documentation
swagger:
	@echo "📝 Generating Swagger documentation..."
	@$(shell go env GOPATH)/bin/swag init -g main.go
	@echo "✅ Swagger documentation updated"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -f stt-app
	rm -rf tmp/
	@echo "✅ Clean completed"

# Kill process running on port 8081
kill:
	@echo "🔪 Killing processes on port 8081..."
	@lsof -ti:8081 | xargs -r kill -9 || echo "✅ No process running on port 8081"
