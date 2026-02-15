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
install-deps: setup-vosk
	@echo "Installing Gradle dependencies..."
	./gradlew build --quiet

VOSK_MODEL_DIR := app/src/main/assets/vosk-model
VOSK_MODEL_URL := https://alphacephei.com/vosk/models/vosk-model-small-en-us-0.15.zip

# Download and setup Vosk Offline Model
setup-vosk:
	@if [ ! -d "$(VOSK_MODEL_DIR)" ] || [ ! -f "$(VOSK_MODEL_DIR)/uuid" ]; then \
		echo "📥 Downloading Vosk offline model..."; \
		mkdir -p $(VOSK_MODEL_DIR); \
		curl -L $(VOSK_MODEL_URL) -o vosk-model.zip; \
		echo "📦 Extracting model..."; \
		unzip -q vosk-model.zip; \
		mv vosk-model-small-en-us-0.15/* $(VOSK_MODEL_DIR)/; \
		echo "9f9392e2-0f0c-4c0c-8d3b-39a7d3c0a5e2" > $(VOSK_MODEL_DIR)/uuid; \
		rm -rf vosk-model.zip vosk-model-small-en-us-0.15; \
		echo "✅ Vosk model setup complete."; \
	else \
		echo "✅ Vosk model already exists in $(VOSK_MODEL_DIR)"; \
	fi

.DEFAULT_GOAL := deploy
