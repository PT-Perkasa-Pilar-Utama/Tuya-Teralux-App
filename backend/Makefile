.PHONY: help test test-auth test-devices test-coverage clean

help:
	@echo "Available commands:"
	@echo "  make dev            - Run app with hot reload (using air)"
	@echo "  make start          - Run app normally"
	@echo "  make test           - Run all E2E tests"
	@echo "  make test-auth      - Run auth endpoint tests only"
	@echo "  make test-devices   - Run device endpoint tests only"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make clean          - Clean test cache"

# Run app with hot reload
dev:
	@echo "Starting app with air..."
	go run github.com/air-verse/air@latest

# Run app normally
start:
	@echo "Starting app..."
	go run main.go

# Run all E2E tests
test:
	@echo "Running all E2E tests..."
	go test -v ./test/...

# Run auth tests only
test-auth:
	@echo "Running auth E2E tests..."
	go test -v ./test/auth/...

# Run device tests only
test-devices:
	@echo "Running device E2E tests..."
	go test -v ./test/devices/...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./test/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean test cache
clean:
	@echo "Cleaning test cache..."
	go clean -testcache
