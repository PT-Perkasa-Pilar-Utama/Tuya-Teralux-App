.PHONY: build deploy clean logs

# Build APK only
build:
	@echo "🔨 Building Whisper Android APK..."
	./gradlew assembleDebug --quiet
	@echo "✅ Build complete: app/build/outputs/apk/debug/app-debug.apk"

# Build and deploy to device
deploy:
	@./deploy.sh

# Deploy to specific device
deploy-%:
	@./deploy.sh $*

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	./gradlew clean --quiet
	@echo "✅ Clean complete"

# View logcat from app
logs:
	adb logcat | grep whisper_android

# Install dependencies
install-deps:
	@echo "Installing dependencies (if needed)..."
	./gradlew build --quiet

.DEFAULT_GOAL := deploy
