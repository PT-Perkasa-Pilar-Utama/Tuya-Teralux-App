# Sensio App - Root Makefile for Project-Wide Automation

.PHONY: help setup backend-setup dev dev-compose dev-server clean kill kill-compose kill-server test vet push-local adb-reverse scrcpy install-remote sync-remote

# Default target
help:
	@echo "Sensio App - Available Project-Wide Commands:"
	@echo ""
	@echo "  make setup          - Setup Backend (install tools, run migrations)"
	@echo "  make backend-setup  - Setup Backend (install tools, run migrations)"
	@echo "  make dev            - Run Backend in dev mode with log capture"
	@echo "  make clean          - Clean backend artifacts"
	@echo "  make kill           - Kill all backend and compose services"
	@echo "  make kill-server    - Kill only backend server and RAG processes"
	@echo "  make adb-reverse    - Expose backend port (8081) to Android device via ADB"
	@echo "  make install-remote - Pull APK from arch host and install to connected device"
	@echo "  make sync-remote    - Sync source code to arch host (delta sync)"
	@echo "  make scrcpy         - Run scrcpy with ADB log capture to @logs/android.log"
	@echo ""
	@echo "  Log output goes to @logs/ directory:"
	@echo "    @logs/backend.log  - Backend server logs"
	@echo "    @logs/android.log   - Android device logs (when connected)"

# Setup everything
setup: backend-setup
	@echo "✅ Full project setup complete!"

# Backend Setup
backend-setup:
	@echo "🚀 Setting up Backend..."
	@$(MAKE) -C backend setup

# Run development mode for the backend with log capture
dev:
	@echo "🚀 Starting backend in development mode..."
	@echo "💡 Tip: Use separate terminals for other tasks."
	@mkdir -p @logs
	@echo "📱 Checking ADB connection..."
	@ADB_DEVICES=$$(adb devices 2>/dev/null | grep -E "^[a-zA-Z0-9]+	device$$" | wc -l || echo "0"); \
	ADB_CONNECTED=$$(adb devices 2>/dev/null | grep -E "^[a-zA-Z0-9]+	device$$" | head -1); \
	ADB_AVAILABLE=false; \
	if [ "$$ADB_DEVICES" -eq "0" ]; then \
		echo ""; \
		echo "⚠️  WARNING: No Android device connected via ADB!"; \
		echo "   ADB log capture will be skipped."; \
		echo ""; \
		echo "Options:"; \
		echo "  1. Press [Enter] to continue WITHOUT ADB log capture"; \
		echo "  2. Type 'fix' and press [Enter] to attempt adb devices - any repair"; \
		echo "  3. Type 'exit' and press [Enter] to cancel"; \
		echo ""; \
		read -r -p "Choose option (1/fix/exit): " CHOICE; \
		case "$$CHOICE" in \
			fix|FIX|Fix) \
				echo "🔧 Attempting ADB fix..."; \
				adb start-server 2>/dev/null || true; \
				adb devices -any 2>/dev/null || true; \
				adb kill-server 2>/dev/null || true; \
				adb start-server 2>/dev/null || true; \
				ADB_DEVICES=$$(adb devices 2>/dev/null | grep -E "^[a-zA-Z0-9]+	device$$" | wc -l || echo "0"); \
				if [ "$$ADB_DEVICES" -eq "0" ]; then \
					echo "❌ ADB fix failed. Proceeding WITHOUT ADB log capture..."; \
				else \
					echo "✅ ADB device reconnected!"; \
					ADB_AVAILABLE=true; \
				fi \
				;; \
			exit|EXIT|Exit) \
				echo "❌ Cancelled."; \
				exit 0 \
				;; \
			*) \
				echo "▶️  Proceeding WITHOUT ADB log capture..." \
				;; \
		esac; \
	else \
		echo "✅ ADB device found: $$ADB_CONNECTED"; \
		ADB_AVAILABLE=true; \
	fi; \
	echo ""; \
	echo "📝 Capturing backend logs to @logs/backend.log..."; \
	(cd backend && $(MAKE) dev 2>&1 | tee -a ../@logs/backend.log) & BACKEND_PID=$$!; \
	ADB_PID=""; \
	if [ "$$ADB_AVAILABLE" = "true" ]; then ADB_WAS_CONNECTED=true; else ADB_WAS_CONNECTED=false; fi; \
	start_adb_capture() { \
		echo "📱 Android connected - capturing to @logs/android.log"; \
		(adb logcat -c 2>/dev/null || true; adb logcat > @logs/android.log 2>&1) & ADB_PID=$$!; \
	}; \
	if [ "$$ADB_AVAILABLE" = "true" ]; then \
		start_adb_capture; \
	else \
		echo "📱 Android NOT connected - skipping ADB log capture"; \
	fi; \
	(while true; do \
		sleep 5; \
		CURRENT_DEVICES=$$(adb devices 2>/dev/null | grep -E "^[a-zA-Z0-9]+	device$$" | wc -l || echo "0"); \
		if [ "$$ADB_WAS_CONNECTED" = "true" ] && [ "$$CURRENT_DEVICES" -eq "0" ]; then \
			echo ""; \
			echo "⚠️  WARNING: ADB device disconnected! ADB log capture stopped."; \
			echo "   Plugin device again to auto-restart ADB log capture"; \
			[ -n "$$ADB_PID" ] && kill $$ADB_PID 2>/dev/null || true; \
			ADB_PID=""; \
			ADB_WAS_CONNECTED=false; \
		elif [ "$$ADB_WAS_CONNECTED" = "false" ] && [ "$$CURRENT_DEVICES" -gt "0" ]; then \
			echo ""; \
			echo "📱 Android connected - capturing to @logs/android.log"; \
			start_adb_capture; \
			ADB_WAS_CONNECTED=true; \
		fi; \
	done) & ADB_MONITOR_PID=$$!; \
	cleanup() { \
		echo ""; \
		echo "🛑 Stopping backend (MySQL keeps running)..."; \
		[ -n "$$BACKEND_PID" ] && kill $$BACKEND_PID 2>/dev/null || true; \
		[ -n "$$ADB_PID" ] && kill $$ADB_PID 2>/dev/null || true; \
		[ -n "$$ADB_MONITOR_PID" ] && kill $$ADB_MONITOR_PID 2>/dev/null || true; \
		$(MAKE) -C backend kill-server 2>/dev/null || true; \
		echo "✅ Backend stopped. Use 'make kill' to stop MySQL too."; \
	}; \
	trap 'cleanup' INT TERM EXIT; \
	wait $$BACKEND_PID 2>/dev/null || true; \
	EXIT_CODE=$$?; \
	kill $$ADB_PID 2>/dev/null || true; \
	kill $$ADB_MONITOR_PID 2>/dev/null || true; \
	trap - INT TERM EXIT; \
	exit $$EXIT_CODE

# Start only compose services
dev-compose:
	@echo "🚀 Starting backend compose services in background..."
	@(cd backend && $(MAKE) dev-compose)

# Start only the backend server and RAG service
dev-server:
	@echo "🚀 Starting backend server..."
	@(cd backend && $(MAKE) dev-server)

# Clean all (including logs)
clean:
	@echo "🧹 Cleaning backend artifacts..."
	@$(MAKE) -C backend clean
	@echo "🧹 Note: @logs/ directory preserved. Delete manually if needed."

# Kill all
kill:
	@echo "🔪 Killing all project components..."
	@$(MAKE) -C backend kill
	@# Also kill any orphaned adb logcat processes
	@pkill -f "adb logcat" 2>/dev/null || true

# Kill only server processes
kill-server:
	@echo "🔪 Killing backend server processes..."
	@$(MAKE) -C backend kill-server

# Kill and cleanup Docker Compose
kill-compose:
	@echo "🐳 Cleaning up Docker containers..."
	@$(MAKE) -C backend kill-compose

# Run tests
test:
	@echo "🧪 Running backend tests..."
	@$(MAKE) -C backend test

# Run vet
vet:
	@echo "🔍 Running backend go vet..."
	@$(MAKE) -C backend vet

# Push local
push-local:
	@echo "🚀 Pushing Docker image locally..."
	@$(MAKE) -C backend push-local TAG=$(TAG)

# Expose backend port to Android device via ADB
adb-reverse:
	@$(MAKE) -C backend adb-reverse

# Pull APK from arch host and install to connected device
install-remote:
	@$(MAKE) -C sensio_app install-remote

# Sync source code to arch host
sync-remote:
	@$(MAKE) -C sensio_app sync-remote

# Run scrcpy with ADB log capture
scrcpy:
	@echo "📱 Starting scrcpy with ADB log capture..."
	@mkdir -p @logs
	@ADB_PID=""; \
	(adb logcat -c 2>/dev/null || true; adb logcat 2>&1 | tee -a @logs/android.log) & ADB_PID=$$!; \
	trap 'kill $$ADB_PID 2>/dev/null || true; exit' INT TERM EXIT; \
	scrcpy; \
	EXIT_CODE=$$?; \
	kill $$ADB_PID 2>/dev/null || true; \
	trap - INT TERM EXIT; \
	exit $$EXIT_CODE
