.PHONY: deploy build lint format

# Build DEBUG APK
build:
		@chmod +x gradlew
		@echo "🔨 Building Sensio Notification APK..."
		./gradlew :composeApp:assembleDebug --quiet
		@echo "✅ Build complete: composeApp/build/outputs/apk/debug/composeApp-debug.apk"

# Build and deploy to device
deploy:
		@chmod +x deploy.sh gradlew
		@./deploy.sh

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
