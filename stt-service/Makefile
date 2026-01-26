# Teralux STT Service - Makefile for Development Automation

.PHONY: help setup dev start build tidy lint swagger clean kill ollama-ensure ollama-pull-whisper ollama-pull-llm ollama-setup

# Ollama defaults (can override when calling make, e.g. `make ollama-pull-llm LLM_MODEL=mistral`)
OLLAMA_IMAGE ?= ollama/ollama:latest
OLLAMA_DATA_DIR ?= $(HOME)/ollama
WHISPER_MODEL ?= whisper
LLM_MODEL ?= gemma

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
	@echo "  make ollama-ensure   - Ensure Ollama image & data dir exist"
	@echo "  make ollama-pull-whisper - Pull whisper model into Ollama (default: $(WHISPER_MODEL))"
	@echo "  make ollama-pull-llm - Pull LLM model into Ollama (default: $(LLM_MODEL))"
	@echo "  make ollama-setup   - Ensure Ollama image and default models (whisper + llm) are available"
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
dev-docker-build: dev-docker-stop
	@echo "🚀 Starting development environment in Docker (Forced Rebuild)..."
	docker compose -f docker-compose.dev.yml up --build

# Stop development environment
dev-docker-stop:
	@echo "🛑 Stopping development environment (removing volumes)..."
	docker compose -f docker-compose.dev.yml down -v

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

# Ensure Ollama image, data dir & container exist
ollama-ensure:
	@echo "🔧 Ensuring Ollama image, data dir & container..."
	@mkdir -p $(OLLAMA_DATA_DIR)
	@docker image inspect $(OLLAMA_IMAGE) >/dev/null 2>&1 || (echo "🔽 Pulling $(OLLAMA_IMAGE)..." && docker pull $(OLLAMA_IMAGE))
	@# Start a named container 'ollama' if it isn't running
	@if [ "$(shell docker ps -q -f name=ollama)" != "" ]; then \
		echo "ℹ️  Container 'ollama' already running"; \
	elif [ "$(shell docker ps -aq -f name=ollama)" != "" ]; then \
		echo "ℹ️  Starting existing container 'ollama'..."; docker start ollama; \
	else \
		echo "🔁 Creating and starting container 'ollama'..."; \
		docker run -d --restart unless-stopped -v $(OLLAMA_DATA_DIR):/root/.ollama -p 11434:11434 --name ollama $(OLLAMA_IMAGE) serve; \
	fi
	@echo "✅ Ollama image, data dir & container ready (data dir: $(OLLAMA_DATA_DIR))"

# Pull whisper model if missing
ollama-pull-whisper: ollama-ensure
	@echo "🔽 Ensuring Whisper model ($(WHISPER_MODEL))..."
	@docker exec -i ollama sh -c 'for i in 1 10; do ollama ps >/dev/null 2>&1 && exit 0 || sleep 1; done; exit 1' \
		&& docker exec -i ollama ollama show $(WHISPER_MODEL) >/dev/null 2>&1 || (docker exec -i ollama ollama pull $(WHISPER_MODEL) 2>&1 || echo "⚠️ Model '$(WHISPER_MODEL)' not found in Ollama repository - skipping"); true
	@echo "✅ Whisper model ensured (or skipped)"

# Pull LLM model if missing
ollama-pull-llm: ollama-ensure
	@echo "🔽 Ensuring LLM model ($(LLM_MODEL))..."
	@docker exec -i ollama sh -c 'for i in 1 10; do ollama ps >/dev/null 2>&1 && exit 0 || sleep 1; done; exit 1' \
		&& docker exec -i ollama ollama show $(LLM_MODEL) >/dev/null 2>&1 || (docker exec -i ollama ollama pull $(LLM_MODEL) 2>&1 || echo "⚠️ Model '$(LLM_MODEL)' not found in Ollama repository - skipping"); true
	@echo "✅ LLM model ensured (or skipped): $(LLM_MODEL)"

# Convenience: ensure both models and image
ollama-setup: ollama-pull-whisper ollama-pull-llm
	@echo "✅ Ollama setup complete (image + whisper + llm)"
