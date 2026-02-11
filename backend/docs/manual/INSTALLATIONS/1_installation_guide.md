# Installation & User Manual

Welcome to the **Teralux Smart Home System** installation guide. This document provides step-by-step instructions to set up, build, and run both the Backend and Android applications.

## 📋 Prerequisites

Before proceeding, ensure you have the following installed on your system:

*   **Go (Golang) v1.20+** (for local backend development)
*   **Air** (for hot reload during development)
*   **Make** (for executing build commands)
*   **Android Studio** (for Android development)
*   **Java Development Kit (JDK) 17** (required for Android build)
*   **Docker & Docker Compose** (for production deployment only)

---

## 🖥️ Backend Setup

The backend is built using Go. For development, we use `air` for hot reloading. Docker is used primarily for production deployment. All commands are standardized using `Make`.

### 1. Navigate to Backend Directory
```bash
cd backend
```

### 2. Configuration (Environment Variables)
Copy the example environment file:
```bash
cp .env.example .env
```
Edit `.env` to configure your database credentials, Tuya API keys, and other settings.

### 3. Available Commands (`make help`)
Run `make help` to see all available commands.
```bash
make help
```

### 4. Run Development Server
To start the server locally with hot-reload enabled (requires `air`):
```bash
make dev
```
If you don't have `air` installed, you can run:
```bash
make start
```

### 5. Build for Production
To build the release binary:
```bash
make build
```

### 6. Docker Deployment
To start the backend using Docker Compose (includes Database, etc.):
```bash
make start-compose
```
To stop the docker environment:
```bash
make stop-compose
```

### 7. Clean Up
To remove build artifacts:
```bash
make clean
```

---

## 📱 Android Setup

The Android application is built using Kotlin and Jetpack Compose. We have provided a `Makefile` to simplify common Gradle tasks.

### 1. Navigate to Android Directory
```bash
cd android
```

### 2. Configuration (local.properties)
Create or edit `local.properties` in the `android` directory to include your SDK path (if not already there) and your API Key:
```properties
sdk.dir=/path/to/your/android/sdk
API_KEY=your_api_key_here
```

### 3. Available Commands (`make help`)
Run `make help` to see all available commands.
```bash
make help
```

### 4. Building the App
To compile the Debug APK:
```bash
make build
```
*This runs `./gradlew assembleDebug` under the hood.*

### 5. Running Tests
To run unit tests:
```bash
make test
```

### 6. Installing on Device
Connect your Android device (ensure USB Debugging is on) or start an Emulator, then run:
```bash
make install
```
*This runs `./gradlew installDebug` and installs the app directly to your device.*

### 7. Clean Up
To clean the build directory:
```bash
make clean
```

---

## 🧹 Full Project Cleanup

If you want to clean both projects entirely:

**Backend:**
```bash
cd backend && make clean
```

**Android:**
```bash
cd android && make clean
```
