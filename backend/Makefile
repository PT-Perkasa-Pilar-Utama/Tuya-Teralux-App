# Sensio App Backend - Makefile for Development Automation

.PHONY: help dev build-docker push pull run stop-docker start-docker test clean kill migrate-up migrate-down migrate-version migrate-force start-compose stop-compose stop lint vet

# Default target
help:
	@echo "Sensio App Backend - Available Commands:"
	@echo ""
	@echo "  --- LOCAL DEV ---"
	@echo "  make dev              - Run development server with hot reload (Air)"
	@echo "  make test             - Run all unit tests"
	@echo "  make lint             - Run golangci-lint"
	@echo ""
	@echo "  --- DOCKER & DEPLOY ---"
	@echo "  make build-docker     - Build image for local architecture"
	@echo "  make push             - Build and Push Multi-Arch image (AMD64+ARM64) to GHCR"
	@echo "  make pull             - Pull latest images from GHCR"
	@echo "  make run              - Start production stack with Docker Compose"
	@echo "  make start-docker     - Run local image for testing"
	@echo "  make stop-docker      - Stop and remove local container"
	@echo "  --- DATABASE ---"
	@echo "  make migrate-up       - Run all pending migrations"
	@echo "  make migrate-down     - Rollback last migration"
	@echo "  make migrate-force VERSION=3 - Force database to version 3"
	@echo "  make emqx-dev         - Start EMQX Auth Service for development"
	@echo ""

# Start MySQL for development
start-mysql-dev:
	docker compose -f docker-compose.dev.yml up -d --wait

# Stop all running containers (App, Compose, Dev DB) without deleting volumes
stop: stop-docker stop-compose stop-mysql-dev

# Stop MySQL for development
stop-mysql-dev:
	docker compose -f docker-compose.dev.yml down

# Run development server with hot reload (requires air)
dev: kill
	@trap 'make stop-mysql-dev' EXIT INT TERM; \
	set -e; \
	make start-mysql-dev; \
	echo "🟢 MySQL started"; \
	echo "🚀 Starting development server with hot reload..."; \
	if [ ! -f bin/whisper-cli ] || [ ! -f bin/llama-cli ] || [ ! -f bin/llama-server ]; then \
		echo "🛠️ Some AI binaries missing. Running setup..."; \
		make setup-dev; \
	fi; \
	mkdir -p .air && \
	$(shell go env GOPATH)/bin/air || (echo "❌ air not found. Installing..." && go install github.com/air-verse/air@latest && $(shell go env GOPATH)/bin/air)

# Run EMQX Auth Service separately
emqx-dev:
	@cd services/EMQX-Auth-Service && make dev

# Run development server without hot reload
start:
	@echo "🚀 Starting development server (no hot reload)..."
	CGO_CFLAGS="-w" go run main.go

# Install air for hot reload (Go equivalent of cargo-watch)
install-watch:
	@echo "📦 Installing Air..."
	go install github.com/air-verse/air@latest
	@echo "✅ Air installed successfully"

# Install swaggo/swag for API documentation
install-swagger:
	@echo "📦 Installing Swag..."
	go install github.com/swaggo/swag/cmd/swag@latest
	@echo "✅ Swag installed successfully"

# Generate Swagger documentation
swagger:
	@echo "📝 Generating Swagger documentation..."
	@$(shell go env GOPATH)/bin/swag init -g main.go -o docs/swagger --parseDependency --parseInternal
	@sed -i 's/package docs/package swagger/g' docs/swagger/docs.go
	@echo "✅ Swagger documentation updated"

# Build release binary
build:
	@echo "🔨 Building release binary..."
	CGO_CFLAGS="-w" go build -o main .

# --- Docker Configuration ---
REGISTRY      ?= ghcr.io
USERNAME      ?= farismnrr
IMAGE_BASE    ?= sensio-backend
IMAGE_TAG     ?= latest
PLATFORMS     ?= linux/amd64,linux/arm64
REPO          = $(REGISTRY)/$(USERNAME)/$(IMAGE_BASE)

# Build via Docker (Local Architecture)
build-docker:
	@echo "🐳 Building Docker image for local architecture..."
	docker build --build-arg AUTO_MIGRATE=true -t $(IMAGE_BASE):$(IMAGE_TAG) .

# Build and Push Multi-Arch Image (AMD64 + ARM64)
push:
	@echo "🔐 Ensuring buildx builder exists..."
	@docker buildx inspect multi-arch-builder > /dev/null 2>&1 || docker buildx create --name multi-arch-builder --use
	@echo "🚀 PUSH TARGET: $(REPO):$(IMAGE_TAG)"
	@echo "📍 PLATFORMS  : $(PLATFORMS)"
	@echo "🛠️ Starting Multi-Arch build and push..."
	docker buildx build --platform $(PLATFORMS) \
		--build-arg AUTO_MIGRATE=true \
		-t $(REPO):$(IMAGE_TAG) --push .
	@echo "✅ Successfully pushed to $(REPO):$(IMAGE_TAG)"

# Pull latest image from Registry
pull:
	@echo "📥 Pulling latest images from $(REGISTRY)..."
	docker compose pull

# Run production stack
run:
	@echo "🚀 Starting Sensio Backend stack..."
	docker compose up -d

# Start Docker Compose (Production/Latest)
start-compose: run

# Stop Docker Compose
stop-compose:
	@echo "🛑 Stopping Docker Compose production stack..."
	docker compose down

# Stop dev/local container
stop-docker:
	@echo "🛑 Stopping local Docker container..."
	@docker stop $(IMAGE_BASE) 2>/dev/null || true
	@docker rm $(IMAGE_BASE) 2>/dev/null || true

# Start dev container locally
start-docker: stop-docker
	@echo "🚀 Starting Docker container locally..."
	docker run --rm -it -d -p 8081:8080 --name $(IMAGE_BASE) --env-file .env $(IMAGE_BASE):$(IMAGE_TAG)

# --- AI Components Configuration ---
WHISPER_TAG ?= v1.7.4
LLAMA_TAG   ?= b4765

# Model Metadata (Pinned from entrypoint.sh)
WHISPER_URL = https://huggingface.co/ggerganov/whisper.cpp/resolve/80da2d8bfee42b0e836fc3a9890373e5defc00a6/ggml-base.bin
WHISPER_SHA = 60ed5bc3dd14eea856493d334349b405782ddcaf0028d4b5df4088345fba2efe

LLAMA_URL = https://huggingface.co/unsloth/Llama-3.2-1B-Instruct-GGUF/resolve/b9a1b2bae395064b27533b8a56b1306f4d2d2f5f/Llama-3.2-1B-Instruct-Q4_K_M.gguf
LLAMA_SHA = 3f5a22426976ab26cfe84dba63c1d08391717abb1af893e10f1b2968d862dcc1

# Paths
WHISPER_SRC_DIR = services/whisper.cpp
WHISPER_BUILD_DIR = $(WHISPER_SRC_DIR)/build
LLAMA_SRC_DIR = services/llama.cpp
LLAMA_BUILD_DIR = $(LLAMA_SRC_DIR)/build

# Run backend setup (whisper & llama build) - Comprehensive Setup
setup:
	@echo "🚀 Running comprehensive backend setup..."
	@# Check for dependencies
	@command -v cmake >/dev/null 2>&1 || { echo >&2 "❌ Error: cmake is not installed. Please install it (e.g., 'sudo pacman -S cmake' or 'sudo apt install cmake')."; exit 1; }
	@mkdir -p bin
	@if [ ! -d $(WHISPER_SRC_DIR)/.git ]; then \
		echo "📥 Cloning whisper.cpp $(WHISPER_TAG)..."; \
		rm -rf $(WHISPER_SRC_DIR); \
		git clone --depth 1 --branch $(WHISPER_TAG) https://github.com/ggerganov/whisper.cpp.git $(WHISPER_SRC_DIR); \
	fi
	@if [ ! -d $(LLAMA_SRC_DIR)/.git ]; then \
		echo "📥 Cloning llama.cpp $(LLAMA_TAG)..."; \
		rm -rf $(LLAMA_SRC_DIR); \
		git clone --depth 1 --branch $(LLAMA_TAG) https://github.com/ggerganov/llama.cpp.git $(LLAMA_SRC_DIR); \
	fi
	@echo "🔨 Building AI binaries..."
	@rm -rf $(WHISPER_BUILD_DIR) $(LLAMA_BUILD_DIR)
	@mkdir -p $(WHISPER_BUILD_DIR)
	@cd $(WHISPER_BUILD_DIR) && cmake .. -DGGML_NATIVE=OFF && make whisper-cli -j$$(nproc)
	@mkdir -p $(LLAMA_BUILD_DIR)
	@cd $(LLAMA_BUILD_DIR) && cmake .. -DGGML_NATIVE=OFF && make llama-cli llama-server -j$$(nproc)
	@cp $(WHISPER_BUILD_DIR)/bin/whisper-cli bin/whisper-cli
	@cp $(LLAMA_BUILD_DIR)/bin/llama-cli bin/llama-cli
	@cp $(LLAMA_BUILD_DIR)/bin/llama-server bin/llama-server
	@$(MAKE) download-models
	@echo "✅ Setup complete. Use 'make dev' to start."

# Legacy target for backward compatibility or simple setup
setup-dev: setup

# Download and verify default models
download-models:
	@echo "📥 Downloading and verifying models..."
	@mkdir -p bin
	@# Verify or download Whisper model
	@if [ -f bin/ggml-base.bin ]; then \
		echo "Verifying Whisper model checksum..."; \
		if ! echo "$(WHISPER_SHA)  bin/ggml-base.bin" | sha256sum -c --status; then \
			echo "⚠️ Whisper checksum mismatch! Re-downloading..."; \
			curl -L $(WHISPER_URL) -o bin/ggml-base.bin; \
		fi; \
	else \
		echo "Downloading Whisper base model..."; \
		curl -L $(WHISPER_URL) -o bin/ggml-base.bin; \
	fi
	@echo "✅ Whisper model verified."
	
	@# Verify or download Llama model
	@if [ -f bin/llama-model.gguf ]; then \
		echo "Verifying Llama model checksum..."; \
		if ! echo "$(LLAMA_SHA)  bin/llama-model.gguf" | sha256sum -c --status; then \
			echo "⚠️ Llama checksum mismatch! Re-downloading..."; \
			curl -L $(LLAMA_URL) -o bin/llama-model.gguf; \
		fi; \
	else \
		echo "Downloading Llama-3.2-1B model..."; \
		curl -L $(LLAMA_URL) -o bin/llama-model.gguf; \
	fi
	@echo "✅ Llama model verified."

# Conveniences
# Usage: make rag text="turn on the lamp"
rag:
	@echo "🔧 Running RAG helper (auth -> submit -> poll)..."
	@# Determine port preference: .env PORT > env PORT > 8081
	@PORT_VAL=$$(grep -h '^PORT=' .env 2>/dev/null | sed 's/PORT=//' || true); \
	if [ -z "$$PORT_VAL" ]; then PORT_VAL="$${PORT:-8081}"; fi; \
	if [ -n "${text}" ]; then \
		RAG_TEXT="${text}" bash scripts/rag.sh "" "${API_KEY}" "http://localhost:$$PORT_VAL"; \
	else \
		bash scripts/rag.sh "" "${API_KEY}" "http://localhost:$$PORT_VAL"; \
	fi

# Run all tests
test:
	@go run scripts/test_runner.go

# Run go vet excluding tmp
vet:
	@echo "🔍 Running go vet..."
	@go vet ./domain/...

# Run golangci-lint
lint:
	@echo "🔍 Running golangci-lint..."
	@$(GOLANGCI_LINT) run ./domain/... || (echo "❌ golangci-lint not found. Installing..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && $(GOLANGCI_LINT) run ./domain/...)

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -f main
	rm -rf .air
	@echo "✅ Clean completed"

# Kill process running on port 8081 (server's default port)
kill:
	@echo "🔪 Killing processes on port 8081..."
	@# Kill local processes
	@lsof -ti:8081 | xargs -r kill -9 || true
	@fuser -k 8081/tcp 2>/dev/null || true
	@# Kill Docker processes/containers
	@echo "🐳 Cleaning up Docker containers..."
	@docker compose down --remove-orphans 2>/dev/null || true
	@docker compose -f docker-compose.dev.yml down --remove-orphans 2>/dev/null || true
	@docker stop sensio-backend sensio-db dev-mysql 2>/dev/null || true
	@docker rm sensio-backend sensio-db dev-mysql 2>/dev/null || true
	@echo "✅ Kill completed"

# --- Database Migration Configuration ---
# Note: We build from source to include required support
MIGRATE_BIN = $(HOME)/go/bin/migrate
GOLANGCI_LINT = $(HOME)/go/bin/golangci-lint

# Load environment variables
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Database connection string (MySQL only)
DB_URL = mysql://$(MYSQL_USER):$(MYSQL_PASSWORD)@tcp($(MYSQL_HOST):$(MYSQL_PORT))/$(MYSQL_DATABASE)?charset=utf8mb4&parseTime=True&loc=Local

# Install golang-migrate from source
$(MIGRATE_BIN):
	@echo "📦 Installing golang-migrate..."
	@go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "✅ golang-migrate installed successfully"

# Run all pending migrations
migrate-up: $(MIGRATE_BIN)
	@echo "⬆️  Running migrations..."
	@$(MIGRATE_BIN) -path domain/terminal/migrations -database "$(DB_URL)" up
	@echo "✅ Migrations completed"

# Rollback last migration
migrate-down: $(MIGRATE_BIN)
	@echo "⬇️  Rolling back last migration..."
	@$(MIGRATE_BIN) -path domain/terminal/migrations -database "$(DB_URL)" down 1
	@echo "✅ Rollback completed"

# Show current migration version
migrate-version: $(MIGRATE_BIN)
	@echo "📊 Current migration version:"
	@$(MIGRATE_BIN) -path domain/terminal/migrations -database "$(DB_URL)" version

# Force database to a specific version (use: make migrate-force VERSION=3)
migrate-force: $(MIGRATE_BIN)
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ Error: VERSION is required. Usage: make migrate-force VERSION=3"; \
		exit 1; \
	fi
	@echo "⚠️  Forcing database to version $(VERSION)..."
	@$(MIGRATE_BIN) -path domain/terminal/migrations -database "$(DB_URL)" force $(VERSION)
	@echo "✅ Database forced to version $(VERSION)"

