# Sensio App Backend - Makefile for Development Automation

.PHONY: help dev dev-compose dev-server build-docker push pull update run stop-docker start-docker test clean kill migrate-up migrate-down migrate-version migrate-force start-compose stop-compose stop lint lint-strict vet push-remote

# Default target
help:
	@echo "Sensio App Backend - Available Commands:"
	@echo ""
	@echo "  --- LOCAL DEV ---"
	@echo "  make dev              - Run development server with hot reload (Air)"
	@echo "  make test             - Run all unit tests"
	@echo "  make lint             - Run go vet, golangci-lint, and go build check"
	@echo "  make lint-strict      - Run all linters (vet, golangci-lint, build check)"
	@echo ""
	@echo "  --- API DOCS (OpenAPI 3.1) ---"
	@echo "  make openapi-tools    - Install OpenAPI converter and validator tools"
	@echo "  make openapi          - Generate OpenAPI 3.1 docs (auto: swagger → convert → validate)"
	@echo "  make swagger          - Generate Swagger 2.0 docs only (intermediate step)"
	@echo "  make validate-openapi - Validate OpenAPI 3.1 specification"
	@echo "  make openapi-check    - Check if docs are in sync (for CI)"
	@echo ""
	@echo "  📖 Scalar API Reference: http://localhost:8080/openapi"
	@echo ""
	@echo "  --- DOCKER & DEPLOY ---"
	@echo "  make build-docker     - Build image for local architecture"
	@echo "  make push             - Trigger GitHub CI Multi-Arch Build & Push"
	@echo "  make push-local       - Build and Push for local host architecture only"
	@echo "  make pull             - Pull latest images from GHCR"
	@echo "  make run              - Start production stack with Docker Compose"
	@echo "  make start-docker     - Run local image for testing"
	@echo "  make stop-docker      - Stop and remove local container"
	@echo "  --- DATABASE ---"
	@echo "  make migrate-up       - Run all pending migrations"
	@echo "  make migrate-down     - Rollback last migration"
	@echo "  make migrate-force VERSION=3 - Force database to version 3"
	@echo "  make emqx-dev         - Start EMQX Auth Service for development"
	@echo "  make adb-reverse      - Expose backend port (8081) to Android device via ADB"
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
	@trap 'echo "🛑 Stopping all services..."; \
	  [ -n "$$AIR_PID" ] && kill $$AIR_PID 2>/dev/null || true; \
	  [ -n "$$LOGS_PID" ] && kill $$LOGS_PID 2>/dev/null || true; \
	  make stop-mysql-dev' EXIT INT TERM; \
	set -e; \
	make start-mysql-dev; \
	echo "🟢 MySQL started"; \
	echo "📂 Following Docker logs..."; \
	docker compose -f docker-compose.dev.yml & LOGS_PID=$$!; \
	make dev-server

# Start only compose services
dev-compose: start-mysql-dev
	@echo "✅ Docker Compose services started in background."

# Start only the backend server and RAG service
dev-server: kill-server
	@echo "🚀 Starting development server with hot reload..."; \
	if [ ! -f bin/whisper-cli ] || [ ! -f bin/llama-cli ] || [ ! -f bin/llama-server ]; then \
		echo "🛠️ Some AI binaries missing. Running setup..."; \
		make setup-dev; \
	fi; \
	if [ ! -d services/rag-whisper-service/venv ]; then \
		echo "🛠️ Python venv missing. Running RAG service setup..."; \
		cd services/rag-whisper-service && python3 -m venv venv && ./venv/bin/pip install --upgrade pip && ./venv/bin/pip install .; \
	fi; \
	mkdir -p .air; \
	echo "🚀 Starting backend with Air (Port 8081)..."; \
	($(shell go env GOPATH)/bin/air || (echo "❌ air not found. Installing..." && go install github.com/air-verse/air@latest && $(shell go env GOPATH)/bin/air)) & AIR_PID=$$!; \
	echo "🚀 Starting RAG+Whisper Service (Port 8000)..."; \
	cd services/rag-whisper-service && make dev

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

# Install OpenAPI tools (converter and validator)
openapi-tools:
	@echo "📦 Installing OpenAPI tools..."
	@mkdir -p bin
	@cd ../scripts/backend/openapi_convert && go build -o ../../../bin/openapi_convert .
	@cd ../scripts/backend/openapi_validate && go build -o ../../../bin/openapi_validate .
	@echo "✅ OpenAPI tools installed successfully"

# Generate Swagger documentation (legacy, intermediate step)
swagger:
	@echo "📝 Generating Swagger documentation..."
	@$(shell go env GOPATH)/bin/swag init -g main.go -o docs/swagger --parseDependency --parseInternal
	@sed -i 's/package docs/package swagger/g' docs/swagger/docs.go
	@echo "✅ Swagger documentation updated"

# Generate OpenAPI 3.1 documentation (full pipeline)
openapi: swagger generate-openapi validate-openapi
	@echo "✅ OpenAPI 3.1 documentation generated and validated"

# Generate OpenAPI from Swagger (conversion step)
generate-openapi:
	@echo "🔄 Converting Swagger 2.0 to OpenAPI 3.1..."
	@bash ../scripts/backend/generate_openapi.sh
	@echo "✅ OpenAPI conversion completed"

# Validate OpenAPI specification
validate-openapi:
	@echo "✅ Validating OpenAPI 3.1 specification..."
	@./bin/openapi_validate
	@echo "✅ OpenAPI validation passed"

# Check if OpenAPI docs are in sync with code (for CI)
openapi-check: swagger generate-openapi validate-openapi
	@echo "🔍 Checking for uncommitted changes in generated docs..."
	@if ! git diff --exit-code -- docs/openapi/; then \
		echo "❌ Error: Generated OpenAPI docs are not committed!"; \
		echo "   Please run 'make openapi' and commit the changes."; \
		exit 1; \
	fi
	@echo "✅ OpenAPI docs are in sync"

# Build release binary (with auto-generated OpenAPI docs)
build: openapi
	@echo "🔨 Building release binary..."
	CGO_CFLAGS="-w" go build -o main .

# --- Docker Configuration ---
REGISTRY      ?= ghcr.io
USERNAME      ?= farismnrr
IMAGE_BASE    ?= sensio-backend
IMAGE_TAG     ?= latest
PLATFORMS     ?= linux/amd64,linux/arm64
REPO          = $(REGISTRY)/$(USERNAME)/$(IMAGE_BASE)

# Push Configuration
PUSH_RETRY_MAX ?= 2
PUSH_RETRY_BACKOFF ?= 5
DOCKERFILE_SYNTAX_FALLBACK ?= 0

# Build via Docker (Local Architecture)
build-docker:
	@echo "🐳 Building Docker image for local architecture..."
	@mkdir -p scripts && cp ../scripts/backend/entrypoint.sh scripts/entrypoint.sh
	docker build --build-arg AUTO_MIGRATE=true -t $(IMAGE_BASE):$(IMAGE_TAG) .

# Direct push to GHCR (No GitHub CI)
push:
	@echo "🚀 Building and Pushing to GHCR directly..."
	@mkdir -p scripts && cp ../scripts/backend/entrypoint.sh scripts/entrypoint.sh
	@ARCH=$$(uname -m); \
	if [ "$$ARCH" = "x86_64" ]; then PLATFORM="linux/amd64"; \
	elif [ "$$ARCH" = "aarch64" ] || [ "$$ARCH" = "arm64" ]; then PLATFORM="linux/arm64"; \
	else PLATFORM="linux/amd64"; fi; \
	echo "📍 Target Platform: $$PLATFORM"; \
	docker buildx build --platform $$PLATFORM --build-arg AUTO_MIGRATE=true -t $(REPO):$(IMAGE_TAG) --push .


# Build and Push Multi-Arch Image from remote host (arch)
push-remote:
	@bash ../scripts/backend/push_remote_arch.sh

# Pull latest image from Registry
pull:
	@echo "📥 Pulling latest images from $(REGISTRY)..."
	docker compose pull

# Update production stack (Pull + Restart with Orphan Cleanup)
update:
	@echo "📥 Pulling latest images from GHCR..."
	docker compose pull
	@echo "🚀 Restarting production stack and cleaning orphans..."
	docker compose up -d --remove-orphans
	@echo "✅ Update complete!"

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
	@echo "📥 Synchronizing and initializing submodules..."
	@git submodule update --init --recursive
	@echo "🔨 Building AI binaries..."
	@rm -rf $(WHISPER_BUILD_DIR) $(LLAMA_BUILD_DIR)
	@mkdir -p $(WHISPER_BUILD_DIR)
	@cd $(WHISPER_BUILD_DIR) && cmake .. -DGGML_NATIVE=OFF && make whisper-cli -j$$(nproc)
	@mkdir -p $(LLAMA_BUILD_DIR)
	@echo "🔨 Configuring llama.cpp..."
	@cd $(LLAMA_BUILD_DIR) && cmake .. -DGGML_NATIVE=OFF || { echo "❌ CMake configuration failed"; exit 1; }
	@echo "🔨 Building llama.cpp..."
	@cd $(LLAMA_BUILD_DIR) && (make llama-cli llama-server -j$$(nproc) || make main server -j$$(nproc) || \
		{ echo "❌ Build failed. Diagnostics:"; \
		  make help | grep -E "llama-cli|main|server" || true; \
		  ls -la bin || true; \
		  exit 1; })
	@cp $(WHISPER_BUILD_DIR)/bin/whisper-cli bin/whisper-cli
	@if [ -f $(LLAMA_BUILD_DIR)/bin/llama-cli ]; then \
		cp $(LLAMA_BUILD_DIR)/bin/llama-cli bin/llama-cli; \
	elif [ -f $(LLAMA_BUILD_DIR)/bin/main ]; then \
		cp $(LLAMA_BUILD_DIR)/bin/main bin/llama-cli; \
	else \
		echo "❌ llama-cli binary not found"; exit 1; \
	fi
	@if [ -f $(LLAMA_BUILD_DIR)/bin/llama-server ]; then \
		cp $(LLAMA_BUILD_DIR)/bin/llama-server bin/llama-server; \
	elif [ -f $(LLAMA_BUILD_DIR)/bin/server ]; then \
		cp $(LLAMA_BUILD_DIR)/bin/server bin/llama-server; \
	else \
		echo "❌ llama-server binary not found"; exit 1; \
	fi
	@echo "🐍 Setting up RAG+Whisper Service..."
	@if [ ! -d services/rag-whisper-service/venv ]; then \
		cd services/rag-whisper-service && python3 -m venv venv; \
	fi
	@cd services/rag-whisper-service && ./venv/bin/pip install --upgrade pip && ./venv/bin/pip install .
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
	@if [ -n "${text}" ]; then \
		RAG_TEXT="${text}" bash ../scripts/backend/rag.sh "" "${API_KEY}" "http://localhost:$$PORT_VAL"; \
	else \
		bash ../scripts/backend/rag.sh "" "${API_KEY}" "http://localhost:$$PORT_VAL"; \
	fi

# Run all tests
test:
	@go run ../scripts/backend/test_runner.go

# Run go vet excluding tmp
vet:
	@echo "🔍 Running go vet..."
	@go vet . ./domain/...

# Run go vet and go build check
lint:
	@$(MAKE) lint-strict

# Run go vet, golangci-lint, and go build check
lint-strict:
	@echo "🔍 Running go vet..."
	@go vet . ./domain/...
	@echo "✨ Running golangci-lint..."
	@golangci-lint run
	@echo "🔨 Running go build check..."
	@CGO_CFLAGS="-w" go build -o /dev/null . ./domain/...

# Run complete deploy-remote pipeline: lint, build, and fetch binary from remote host
deploy-remote:
	@bash ../scripts/backend/deploy_remote_arch.sh

# Run remote lint and build on arch host (compatibility alias/standalone workflow)
remote-lint-build:
	@bash ../scripts/backend/remote_lint_build_arch.sh

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -f main
	rm -rf .air
	@echo "✅ Clean completed"

# Kill all dev components
kill: kill-server kill-compose
	@echo "✅ Full kill completed"

# Kill only server processes (Go, Air, Python RAG)
kill-server:
	@echo "🔪 Killing server processes (Ports 8081, 8080, 8000)..."
	@# Kill processes by ports
	@lsof -ti:8081,8080,8000 | xargs -r kill -9 2>/dev/null || true
	@# Kill air specifically if still alive (use more specific matching)
	@pkill -x air 2>/dev/null || true
	@echo "✅ Server processes killed"

# Kill and cleanup Docker Compose services
kill-compose:
	@echo "🐳 Cleaning up Docker containers..."
	@docker compose down --remove-orphans 2>/dev/null || true
	@docker compose -f docker-compose.dev.yml down --remove-orphans 2>/dev/null || true
	@docker stop sensio-backend sensio-db dev-mysql 2>/dev/null || true
	@docker rm sensio-backend sensio-db dev-mysql 2>/dev/null || true
	@echo "✅ Docker cleanup completed"

# Expose backend port to Android device via ADB
adb-reverse:
	@echo "📲 Exposing backend port (8081) to Android device..."
	@adb reverse tcp:8081 tcp:8081
	@echo "✅ ADB reverse completed"

# --- Database Migration Configuration ---
# Note: We build from source to include required support
MIGRATE_BIN = $(HOME)/go/bin/migrate

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
	@$(MIGRATE_BIN) -path migrations -database "$(DB_URL)" up
	@echo "✅ Migrations completed"

# Rollback last migration
migrate-down: $(MIGRATE_BIN)
	@echo "⬇️  Rolling back last migration..."
	@$(MIGRATE_BIN) -path migrations -database "$(DB_URL)" down 1
	@echo "✅ Rollback completed"

# Show current migration version
migrate-version: $(MIGRATE_BIN)
	@echo "📊 Current migration version:"
	@$(MIGRATE_BIN) -path migrations -database "$(DB_URL)" version

# Force database to a specific version (use: make migrate-force VERSION=3)
migrate-force: $(MIGRATE_BIN)
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ Error: VERSION is required. Usage: make migrate-force VERSION=3"; \
		exit 1; \
	fi
	@echo "⚠️  Forcing database to version $(VERSION)..."
	@$(MIGRATE_BIN) -path migrations -database "$(DB_URL)" force $(VERSION)
	@echo "✅ Database forced to version $(VERSION)"
