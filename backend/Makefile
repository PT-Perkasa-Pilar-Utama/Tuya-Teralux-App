# Teralux App Backend - Makefile for Development Automation

.PHONY: help dev start install-watch build build-docker start-docker stop-docker push push-local pull-docker start-compose stop-compose update test clean kill migrate-up migrate-down migrate-version migrate-force

# Default target
help:
	@echo "Teralux App Backend - Available Commands:"
	@echo ""
	@echo "  make dev              - Run development server with hot reload (Air)"
	@echo "  make start            - Run development server without hot reload"
	@echo "  make install-watch    - Install Air for hot reload"
	@echo "  make build            - Build release binary"
	@echo "  make build-docker     - Build Docker image"
	@echo "  make start-docker     - Run Docker image (with .env)"
	@echo "  make stop-docker      - Stop and remove Docker container"
	@echo "  make push             - Push Docker image to GHCR"
	@echo "  make push-local       - Build+push Docker image locally (source from GitHub)"
	@echo "  make pull-docker      - Pull latest image for Docker Compose"
	@echo "  make start-compose    - Start Docker Compose stack (pulls first)"
	@echo "  make stop-compose     - Stop Docker Compose stack"
	@echo "  make update           - Update running container using Watchtower"
	@echo "  make lint             - Run golangci-lint"
	@echo "  make test             - Run all unit tests"
	@echo "  make clean            - Clean build artifacts"
	@echo "  make kill             - Kill process running on PORT 8081"
	@echo "  make migrate-up       - Run all pending migrations"
	@echo "  make migrate-down     - Rollback last migration"
	@echo "  make migrate-version  - Show current migration version"
	@echo "  make migrate-force    - Force database to specific version (VERSION=N)"
	@echo ""

# Start MySQL for development
start-mysql-dev:
	docker compose -f docker-compose.dev.yml up -d --wait

# Stop MySQL for development
stop-mysql-dev:
	docker compose -f docker-compose.dev.yml down -v

# Run development server with hot reload (requires air)
dev:
	@trap 'make stop-mysql-dev' EXIT INT TERM; \
	set -e; \
	make start-mysql-dev; \
	echo "🟢 MySQL started"; \
	echo "🚀 Starting development server with hot reload..."; \
	if [ ! -f bin/whisper-cli ] || [ ! -f bin/llama-cli ] || [ ! -f bin/llama-server ]; then \
		echo "🛠️ Some AI binaries missing. Running setup..."; \
		make setup-dev; \
	fi; \
	$(shell go env GOPATH)/bin/air || (echo "❌ air not found. Installing..." && go install github.com/air-verse/air@latest && $(shell go env GOPATH)/bin/air)

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
DOCKER_IMAGE_NAME = teralux-backend
GIT_REPO_URL = https://github.com/farismnrr/Tuya-Teralux-App.git
# GHCR_REPO = ghcr.io/pt-perkasa-pilar-utama/teralux-backend
GHCR_REPO = ghcr.io/farismnrr/teralux-backend
DOCKER_TAG = latest

# Build via Docker
build-docker:
	@echo "🐳 Building Docker image..."
	@mkdir -p ~/.docker/cli-plugins
	@if [ ! -f ~/.docker/cli-plugins/docker-buildx ]; then \
		echo "📦 Installing docker-buildx plugin..."; \
		BUILDX_VERSION=$$(curl -s https://api.github.com/repos/docker/buildx/releases/latest | grep '"tag_name"' | sed -E 's/.*"v([^"]+)".*/\1/'); \
		curl -fsSL "https://github.com/docker/buildx/releases/download/v$$BUILDX_VERSION/buildx-v$$BUILDX_VERSION.linux-amd64" -o ~/.docker/cli-plugins/docker-buildx; \
		chmod +x ~/.docker/cli-plugins/docker-buildx; \
		echo "✅ docker-buildx installed"; \
	fi
	@docker buildx version || (echo "❌ buildx installation failed" && exit 1)
	@docker buildx build --build-arg AUTO_MIGRATE=true -t $(DOCKER_IMAGE_NAME):latest --load .

# Stop dev container
stop-docker:
	@echo "🛑 Stopping Docker container..."
	@docker stop $(DOCKER_IMAGE_NAME) 2>/dev/null || true
	@docker rm $(DOCKER_IMAGE_NAME) 2>/dev/null || true

# Start dev container locally
start-docker: stop-docker
	@echo "🚀 Starting Docker container..."
	docker run --rm -it -d -p 8081:8080 --name $(DOCKER_IMAGE_NAME) --env-file .env $(DOCKER_IMAGE_NAME):latest

# --- Development (Docker) & Ollama helpers (migrated from stt-service)
OLLAMA_IMAGE ?= ollama/ollama:latest
OLLAMA_DATA_DIR ?= $(HOME)/ollama
WHISPER_MODEL ?= whisper
LLM_MODEL ?= gemma

# Local Inference (whisper.cpp & llama.cpp)
WHISPER_SRC_DIR = services/whisper.cpp
WHISPER_BUILD_DIR = $(WHISPER_SRC_DIR)/build
LLAMA_SRC_DIR = services/llama.cpp
LLAMA_BUILD_DIR = $(LLAMA_SRC_DIR)/build

# Run backend setup (whisper & llama build) - Non Interactive
setup-dev:
	@echo "🚀 Running backend setup (whisper & llama build)..."
	@echo "🧹 Cleaning previous build artifacts to avoid CMake cache issues..."
	@rm -rf $(WHISPER_BUILD_DIR) $(LLAMA_BUILD_DIR)
	@mkdir -p $(WHISPER_BUILD_DIR)
	@cd $(WHISPER_BUILD_DIR) && cmake .. && make whisper-cli -j$$(nproc)
	@mkdir -p $(LLAMA_BUILD_DIR)
	@cd $(LLAMA_BUILD_DIR) && cmake .. && make llama-cli llama-server -j$$(nproc)
	@mkdir -p bin
	@cp $(WHISPER_BUILD_DIR)/bin/whisper-cli bin/whisper-cli
	@cp $(LLAMA_BUILD_DIR)/bin/llama-cli bin/llama-cli
	@cp $(LLAMA_BUILD_DIR)/bin/llama-server bin/llama-server
	@$(MAKE) download-models
	@echo "✅ Setup complete. use 'make dev' to start."

# Run setup script (build whisper.cpp + tools)
setup: 
	@echo "🚀 Running backend setup (build + tools)"
	@chmod +x setup.sh || true
	@./setup.sh || true

# Download default models for dev
download-models:
	@echo "📥 Downloading models to bin/..."
	@mkdir -p bin
	@if [ ! -f bin/ggml-base.bin ]; then \
		echo "Downloading Whisper base model..."; \
		curl -L https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.bin -o bin/ggml-base.bin; \
	fi
	@if [ ! -f bin/llama-model.gguf ]; then \
		echo "Downloading Llama-3.2-1B model..."; \
		curl -L https://huggingface.co/unsloth/Llama-3.2-1B-Instruct-GGUF/resolve/main/Llama-3.2-1B-Instruct-Q4_K_M.gguf -o bin/llama-model.gguf; \
	fi
	@echo "✅ Models downloaded."

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

# Push to GHCR via GitHub Actions
push:
	@echo "🚀 Triggering GitHub Actions workflow for Docker push..."
	@command -v gh >/dev/null 2>&1 || ( \
		if command -v apt-get >/dev/null 2>&1; then \
			echo "⬇️  Installing GitHub CLI via apt..."; \
			SUDO=$$(command -v sudo >/dev/null 2>&1 && echo sudo || echo); \
			$$SUDO apt-get update && $$SUDO apt-get install -y gh || { echo "❌ Failed to install gh"; exit 1; }; \
		else \
			echo "❌ GitHub CLI 'gh' not found and auto-install is not configured for this OS."; \
			echo "   Install from https://cli.github.com/ then rerun 'make push'"; \
			exit 1; \
		fi \
	)
	@if [ -z "$(TAG)" ]; then \
		read -p "Enter Docker Tag (default: latest): " input_tag; \
		TAG="$${input_tag:-latest}"; \
	else \
		TAG="$(TAG)"; \
	fi; \
	REF="$$TAG"; \
	[ "$$TAG" = "latest" ] && REF="main"; \
	echo "📦 Using tag '$$TAG' (ref: $$REF)"; \
	gh workflow run backend-docker-publish.yml --ref "$$REF" -f tag="$$TAG"; \
	echo "✅ Workflow dispatched. Track with 'gh run watch --latest'"

# Push to GHCR locally (multi-arch) without GitHub Actions
push-local:
	@echo "🔐 Logging in to GHCR..."
	# @export $$(grep -v '^#' .env | grep -v '^$$' | sed 's/#.*//g' | xargs) && \
	# (echo "$$${CR_PAT:-$$GITHUB_TOKEN}" | docker login ghcr.io -u farismnrr --password-stdin)
	@if [ -z "$(TAG)" ]; then \
		read -p "Enter Docker Tag (default: latest): " input_tag; \
		TAG="$${input_tag:-latest}"; \
	else \
		TAG="$(TAG)"; \
	fi; \
	REF="$$TAG"; \
	[ "$$TAG" = "latest" ] && REF="main"; \
	echo "🚀 Using tag '$$TAG' (source ref: $$REF)"; \
	docker buildx inspect multi-arch-builder > /dev/null 2>&1 || docker buildx create --name multi-arch-builder --use; \
	echo "🚀 Building and pushing '$$TAG' (multi-arch: amd64, arm64)..."; \
	docker buildx build --platform linux/amd64,linux/arm64 --build-arg AUTO_MIGRATE=true -t $(GHCR_REPO):$$TAG --push .; \
	echo "✅ Images pushed to GHCR (Tags: $$TAG)"

# Pull latest image for Docker Compose
pull-docker:
	@echo "📥 Pulling Docker images..."
	docker compose pull

# Start Docker Compose (Production/Latest)
start-compose:
	@echo "🚀 Starting Docker Compose (Latest)..."
	docker compose -f docker-compose.yml up -d

# Stop Docker Compose
stop-compose:
	@echo "🛑 Stopping Docker Compose usage..."
	docker compose -f docker-compose.yml down

# Update running container using Watchtower
update:
	@echo "🔄 Checking for updates with Watchtower..."
	@export $$(grep -v '^#' .env | grep -v '^$$' | xargs) && \
	docker run --rm \
		-v /var/run/docker.sock:/var/run/docker.sock \
		--env DOCKER_API_VERSION=1.45 \
		--env REPO_USER=farismnrr \
		--env REPO_PASS="$${CR_PAT:-$$GITHUB_TOKEN}" \
		containrrr/watchtower \
		--debug \
		--run-once \
		$(DOCKER_IMAGE_NAME)

# Run all tests
test:
	@go run scripts/test_runner.go

# Run go vet excluding tmp
vet:
	@echo "🔍 Running go vet..."
	@go list ./... 2>/dev/null | grep -v /tmp/ | xargs go vet

# Run golangci-lint
lint:
	@echo "🔍 Running golangci-lint..."
	@$(GOLANGCI_LINT) run ./... || (echo "❌ golangci-lint not found. Installing..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && $(GOLANGCI_LINT) run ./...)


# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	go clean
	rm -f main
	@echo "✅ Clean completed"

# Kill process running on port 8081 (server's default port)
kill:
	@echo "🔪 Killing processes on port 8081..."
	@lsof -ti:8081 | xargs -r kill -9 || echo "✅ No process running on port 8081"

# --- Database Migration Configuration ---
# Note: We build from source to include SQLite support (prebuilt binaries don't have it)
MIGRATE_BIN = $(HOME)/go/bin/migrate
GOLANGCI_LINT = $(HOME)/go/bin/golangci-lint

# Load environment variables
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Database connection string (MySQL only)
DB_URL = mysql://$(MYSQL_USER):$(MYSQL_PASSWORD)@tcp($(MYSQL_HOST):$(MYSQL_PORT))/$(MYSQL_DATABASE)?charset=utf8mb4&parseTime=True&loc=Local

# Install golang-migrate from source with SQLite support
$(MIGRATE_BIN):
	@echo "📦 Installing golang-migrate with SQLite support..."
	@go install -tags 'sqlite3 mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "✅ golang-migrate installed successfully"

# Run all pending migrations
migrate-up: $(MIGRATE_BIN)
	@echo "⬆️  Running migrations..."
	@$(MIGRATE_BIN) -path domain/teralux/migrations -database "$(DB_URL)" up
	@echo "✅ Migrations completed"

# Rollback last migration
migrate-down: $(MIGRATE_BIN)
	@echo "⬇️  Rolling back last migration..."
	@$(MIGRATE_BIN) -path domain/teralux/migrations -database "$(DB_URL)" down 1
	@echo "✅ Rollback completed"

# Show current migration version
migrate-version: $(MIGRATE_BIN)
	@echo "📊 Current migration version:"
	@$(MIGRATE_BIN) -path domain/teralux/migrations -database "$(DB_URL)" version

# Force database to a specific version (use: make migrate-force VERSION=3)
migrate-force: $(MIGRATE_BIN)
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ Error: VERSION is required. Usage: make migrate-force VERSION=3"; \
		exit 1; \
	fi
	@echo "⚠️  Forcing database to version $(VERSION)..."
	@$(MIGRATE_BIN) -path domain/teralux/migrations -database "$(DB_URL)" force $(VERSION)
	@echo "✅ Database forced to version $(VERSION)"

