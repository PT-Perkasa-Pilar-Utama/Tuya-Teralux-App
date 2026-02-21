# Sensio App Backend

Backend service for the Sensio application, built with Go.

## 🚀 Getting Started

### Prerequisites
- **Go** (for execution and development)
- **Air** (for hot reload during development)
- **Docker** & **Docker Compose** (for production deployment only)
- **Make** (standardized command runner)

### Setup
1.  **Clone the repository** (if you haven't already).
2.  **Environment Variables**:
    Copy the example environment file and configure it:
    ```bash
    cp .env.example .env
    ```
    Edit `.env` with your specific configuration (database credentials, API keys, etc.).

---

## 🗄️ Database Configuration

The application uses **MySQL** as its database engine.

**Features:**
- ✅ **Standard RDBMS**: Uses MySQL for robust data management.
- ✅ **Auto-migration**: Entities are automatically migrated on startup.
- ✅ **Persistence**: Database is persisted in the `mysql_dev_data` volume during development.

---

## 🎙️ Speech Processing (Whisper)

The application provides a robust speech-to-text pipeline with multiple processing paths:

1.  **Standard Transcription (`POST /api/speech/transcribe`)**: 
    - Optimized for short audio clips.
    - Automatic fallback: Attempts **Orion (Outsystems)** first, falls back to **Local Whisper** if Orion is offline.
    - Post-processing: Integrated RAG for grammar correction and spelling refinement.
2.  **Long Transcription (`POST /api/speech/transcribe/whisper/cpp`)**:
    - Direct access to the heavy-duty **Local Whisper.cpp** engine.
    - Explicit language selection support.
3.  **Orion Transcription (`POST /api/speech/transcribe/orion`)**:
    - Direct proxy to external Outsystems Orion service.
4.  **Status Tracking (`GET /api/speech/transcribe/{task_id}`)**:
    - Consolidated endpoint to check status and fetch results for any transcription task.

---

## 📂 Recordings Management

A dedicated domain for managing audio recording files:

- **Storage**: Files are stored as UUID-named `.wav` or original format in the local `uploads` directory.
- **Serving**: Static files are served at `/uploads/recordings/`.
- **API**: `GET /api/recordings` provides a paginated list of all successfully processed recordings with metadata.

---

## 📚 API Documentation (Swagger)

The project uses [Swaggo](https://github.com/swaggo/swag) to generate Swagger/OpenAPI documentation.

- **Access**: When the server is running, visit `http://localhost:8081/swagger/index.html` to view the interactive API docs.
- **Update Docs**: If you modify API comments, run `swag init` (or `make build` if configured) to regenerate the documentation.

## ⚡ Caching

The application uses **BadgerDB**, a fast embedded key-value store, for caching purposes to enhance performance.

- **Storage**: Data is cached locally on disk/memory using Badger.
- **Management**: 
  - There is an API endpoint to flush the cache if needed: `DELETE /api/cache/flush`.
  - This is useful for clearing stale data without restarting the server.

---

## 🏃 Running the Application

You can run the application manually (directly on your machine) or using Docker. The project includes a `Makefile` to simplify these commands.

### Option 1: Manual Execution

#### Standard Run
To run the server normally:
```bash
make start
```
*Alternatively: `go run main.go`*

#### Development Mode (Hot Reload)
To run with hot reload enabled (uses [Air](https://github.com/air-verse/air)):
```bash
make dev
```
*Note: If `air` is not installed, the command will attempt to install it for you. You can also manually install it with `make install-watch`.*

---

### Running with Docker (Production)

To start the backend using Docker Compose (pulls latest image from registry):
```bash
make start-compose
```
*Alternatively: `docker compose up -d`*

To stop the Docker Compose stack:
```bash
make stop-compose
```

**Setup for Native Whisper (whisper.cpp)**

If you need to build local whisper artifacts (whisper-cli, models) run:

```bash
make setup
```

**Whisper submodule**

The repository includes `whisper.cpp` as a Git submodule located at `backend/whisper.cpp`. After cloning (or when switching branches), make sure to initialize and fetch submodules:

```bash
# from repository root
git submodule update --init --recursive
```

This ensures the native Whisper sources and scripts are available for `make setup`.

---

## 🛠️ Available Make Commands

The `Makefile` includes several utility commands to manage the project:

| Command | Description |
| :--- | :--- |
| `make help` | Show all available commands |
| `make dev` | Run development server with hot reload |
| `make start` | Run development server without hot reload |
| `make install-watch` | Install Air (hot reload tool) |
| `make build` | Build the project binary |
| `make start-compose` | Start the Docker Compose stack (Production) |
| `make stop-compose` | Stop the Docker Compose stack (Production) |
| `make update` | Update running container using Watchtower |
| `make clean` | Clean build artifacts |
| `make kill` | Kill any process running on port 8081 |
| `make rag text="turn on the lamp"` | Helper to authenticate, submit RAG text, poll and print final decision |

Notes:
- You can run `./scripts/rag.sh` without args if you set `RAG_TEXT` (env) and `API_KEY` in `.env`.
- You can also pipe text: `echo "turn on the lamp" | ./scripts/rag.sh`
- `make rag` will use `PORT` from `.env` if present, otherwise it uses the exported `PORT` env var, and finally defaults to `8081`.
