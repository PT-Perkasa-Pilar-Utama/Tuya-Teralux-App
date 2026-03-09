.PHONY: build deploy clean logs

# Build DEBUG APK
build:
	@chmod +x gradlew
	@echo "🔨 Building Whisper Android APK..."
	./gradlew assembleDebug --quiet
	@echo "✅ Build complete: app/build/outputs/apk/debug/app-debug.apk"

# Run Linter (ktlint)
lint:
	@chmod +x gradlew
	@echo "🔍 Running Kotlin Linter..."
	@./gradlew ktlintCheck
	@echo "✅ Lint check complete"

# Fix Lint issues automatically
lint-fix:
	@chmod +x gradlew
	@echo "🔧 Fixing Kotlin Lint issues..."
	@./gradlew ktlintFormat
	@echo "✅ Lint fix complete"

# Run complete deploy-remote pipeline: lint, build, and deploy via local ADB
deploy-remote:
	@bash scripts/deploy_remote_arch.sh

# Run remote lint and build on arch host
remote-lint-build:
	@bash scripts/remote_lint_build_arch.sh

# Pull built APK from remote host manually
remote-pull-apk:
	@echo "📥 Pulling APK from arch host..."
	@mkdir -p .remote-build
	@rsync -az arch:~/Documents/Tuya-Teralux-App/sensio_app/app/build/outputs/apk/debug/app-debug.apk .remote-build/app-debug.apk
	@echo "✅ APK successfully pulled to: .remote-build/app-debug.apk"

# Install built APK to device
install:
	@echo "📲 Installing APK..."
	@adb install -t -r -g app/build/outputs/apk/debug/app-debug.apk

# Build, Lint, and Install
deploy: build lint install

# Deploy to specific device
deploy-%:
	@chmod +x deploy.sh gradlew
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
