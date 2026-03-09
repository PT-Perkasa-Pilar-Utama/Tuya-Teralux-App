.PHONY: deploy build lint format

# Build DEBUG APK
build:
		@chmod +x gradlew
		@echo "🔨 Building Sensio Notification APK..."
		./gradlew :composeApp:assembleDebug --quiet
		@echo "✅ Build complete: composeApp/build/outputs/apk/debug/composeApp-debug.apk"

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
		@rsync -az arch:~/Documents/Tuya-Teralux-App/sensio_notification/composeApp/build/outputs/apk/debug/composeApp-debug.apk .remote-build/composeApp-debug.apk
		@echo "✅ APK successfully pulled to: .remote-build/composeApp-debug.apk"

# Install built APK to device
install:
	@echo "📲 Installing APK..."
	@adb install -t -r -g composeApp/build/outputs/apk/debug/composeApp-debug.apk

# Build, Lint, and Install
deploy: build lint install

# Deploy to specific device
deploy-%:
		@chmod +x deploy.sh gradlew
		@./deploy.sh $*

# Run Linter (ktlint)
lint:
		@chmod +x gradlew
		@echo "🔍 Running Kotlin Linter..."
		./gradlew ktlintCheck
		@echo "✅ Lint check complete"

# Fix Lint issues automatically
format:
		@chmod +x gradlew
		@echo "🔧 Fixing Kotlin Lint issues..."
		./gradlew ktlintFormat
		@echo "✅ Lint fix complete"

# Clean build artifacts
clean:
		@echo "🧹 Cleaning build artifacts..."
		./gradlew clean --quiet
		@echo "✅ Clean complete"

.DEFAULT_GOAL := deploy
