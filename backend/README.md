# Teralux App Backend

Backend service for the Teralux application, built with Go.

## 🚀 Getting Started

### Prerequisites
- **Go** (for manual execution)
- **Docker** & **Docker Compose** (for containerized execution)
- **Make** (optional but recommended for running commands)

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

The application uses **SQLite** as its database engine. This provides a zero-configuration setup that is perfect for development and production for this scale.

### SQLite Configuration

```env
DB_TYPE=sqlite
DB_SQLITE_PATH=./tmp/teralux.db
```

**Features:**
- ✅ **Zero configuration**: No external database server required.
- ✅ **Auto-migration**: Entities are automatically migrated on startup.
- ✅ **Persistence**: When running in Docker, the database is persisted in the `teralux_data` volume.

### Migration Files

Current migrations and entity definitions are handled automatically by GORM auto-migration.

**Tables Created:**
1. **teralux** - Main teralux devices table with soft delete support
2. **devices** - Device information table
3. **device_statuses** - Device status tracking table

---

## �📚 API Documentation (Swagger)

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

### Option 2: Running with Docker

#### Run Single Container (Local Build)
To build the Docker image locally and run it immediately:
```bash
make start-docker
```
This command will:
1. Build the image as `teralux-backend`.
2. Run the container with your local `.env` file.
3. Expose port `8081`.

#### Run via Docker Compose
To run using Docker Compose (pulls the latest image from the registry):
```bash
make start-compose
```
*Alternatively: `docker compose up -d`*

To stop the Docker Compose stack:
```bash
make stop-compose
```

**Development (Docker) with whisper.cpp & Ollama**

For local development that requires `whisper.cpp` binaries and Ollama support, use the docker-compose development helpers:

```bash
make dev-docker
make dev-docker-build
make dev-docker-stop
```

If you need to build local whisper artifacts (whisper-cli, models) run:

```bash
make setup
# or from repository root:
make setup-stt
```

**Whisper submodule**

The repository includes `whisper.cpp` as a Git submodule located at `backend/whisper.cpp`. After cloning (or when switching branches), make sure to initialize and fetch submodules:

```bash
# from repository root
git submodule update --init --recursive
```

This ensures the native Whisper sources and scripts are available for `make setup` and Docker builds.

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
| `make build-docker` | Build the Docker image locally |
| `make start-docker` | Run the Docker image locally |
| `make push` | Build and push multi-arch image to GHCR |
| `make pull-docker` | Pull the latest image for Docker Compose |
| `make start-compose` | Start the Docker Compose stack |
| `make stop-compose` | Stop the Docker Compose stack |
| `make update` | Update running container using Watchtower |
| `make clean` | Clean build artifacts |
| `make kill` | Kill any process running on port 8081 |
| `make rag text="turn on the lamp"` | Helper to authenticate, submit RAG text, poll and print final decision |

Notes:
- You can run `./scripts/rag.sh` without args if you set `RAG_TEXT` (env) and `API_KEY` in `.env`.
- You can also pipe text: `echo "turn on the lamp" | ./scripts/rag.sh`
- `make rag` will use `PORT` from `.env` if present, otherwise it uses the exported `PORT` env var, and finally defaults to `8081`.
