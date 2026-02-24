
.PHONY: clean build test install help

# Default target
help:
	@echo "Android Project - Available Commands:"
	@echo ""
	@echo "  make clean    - Clean build directories"
	@echo "  make build    - Build Debug APK"
	@echo "  make test     - Run unit tests"
	@echo "  make install  - Install Debug APK to connected device"
	@echo ""

clean:
	@echo "🧹 Cleaning Android project..."
	./gradlew clean

build:
	@echo "🔨 Building Debug APK..."
	./gradlew assembleDebug

test:
	@echo "🧪 Running tests..."
	./gradlew test

install:
	@echo "📲 Installing Debug APK..."
	./gradlew installDebug
