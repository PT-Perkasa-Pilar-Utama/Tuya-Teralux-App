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
    Important: duration-based settings must use Go duration format (for example: `8h`, `30m`, `24h`).
    The backend is configured to fail fast on startup if required duration env vars are missing or invalid.

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

3.  **Model-Specific Transcription (`POST /api/speech/models/{provider}`)**:
    - Direct access to specific model providers (Gemini, OpenAI, Groq, Orion, Whisper.cpp).
    - Returns a `task_id` for asynchronous polling.
4.  **Status Tracking (`GET /api/speech/transcribe/{task_id}`)**:
    - Consolidated endpoint to check status and fetch results for any transcription task.

---

## 📂 Recordings Management

A dedicated domain for managing audio recording files:

- **Storage**: Files are stored as UUID-named `.wav` or original format in the local `uploads` directory.
- **Serving**: Static files are served at `/uploads/recordings/`.
- **API**: `GET /api/recordings` provides a paginated list of all successfully processed recordings with metadata.

---

## 📚 API Documentation (OpenAPI 3.1)

The project uses [Swaggo](https://github.com/swaggo/swag) to generate Swagger 2.0, which is then automatically converted to **OpenAPI 3.1.0** (strict).

### Access API Docs

- **Primary (OpenAPI 3.1)**: `http://localhost:8081/openapi/` - Modern interactive docs using Scalar
- **Legacy (Swagger 2.0)**: `http://localhost:8081/swagger/` - Redirects to OpenAPI docs
- **Raw JSON**: `http://localhost:8081/openapi/openapi.json`
- **Raw YAML**: `http://localhost:8081/openapi/openapi.yaml`

### Auto-Generation Pipeline

API documentation is **automatically generated** during development and build:

```
Go Annotations → Swagger 2.0 → OpenAPI 3.1 → Validation → Publish
```

**Manual Commands:**

```bash
# Install OpenAPI tools (converter + validator)
make openapi-tools

# Generate full OpenAPI 3.1 docs (Swagger → Convert → Validate)
make openapi

# Validate OpenAPI spec only
make validate-openapi

# Check if docs are in sync (for CI)
make openapi-check
```

**Development Workflow:**

- `make dev` - Auto-generates docs on every code change (via Air pre-command)
- `make build` - Auto-generates docs before building binary

**CI/CD:**

A GitHub Actions workflow (`backend-openapi-check.yml`) validates OpenAPI spec on every push/PR:
- Generates docs from source
- Validates OpenAPI 3.1 compliance
- Fails if generated docs are not committed

---

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

_Alternatively: `go run main.go`_

#### Development Mode (Hot Reload)

To run with hot reload enabled (uses [Air](https://github.com/air-verse/air)):

```bash
make dev
```

_Note: If `air` is not installed, the command will attempt to install it for you. You can also manually install it with `make install-watch`._

---

### Running with Docker (Production)

The production stack uses pre-built images from `ghcr.io/farismnrr/sensio-backend`.

**Deployment Runbook:**

```bash
# Pull latest images and restart containers
docker compose pull && docker compose up -d
```

**Important Notes:**

- **Registry**: Images are pulled from `ghcr.io/farismnrr/sensio-backend`.
- **Migrations**: Production images are built with `AUTO_MIGRATE=false`. You must run migrations manually using `make migrate-up` from a dev environment or a jump host with access to the production database.
- **Local Rebuilds**: The `docker-compose.yml` does not include a `build` context to prevent accidental local rebuilds on production hosts.

To stop the stack:

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

| Command                            | Description                                                            |
| :--------------------------------- | :--------------------------------------------------------------------- |
| `make help`                        | Show all available commands                                            |
| `make dev`                         | Run development server with hot reload                                 |
| `make start`                       | Run development server without hot reload                              |
| `make install-watch`               | Install Air (hot reload tool)                                          |
| `make build`                       | Build the project binary                                               |
| `make start-compose`               | Start the Docker Compose stack (Production)                            |
| `make stop-compose`                | Stop the Docker Compose stack (Production)                             |
| `make update`                      | Update running container using Watchtower                              |
| `make clean`                       | Clean build artifacts                                                  |
| `make kill`                        | Kill any process running on port 8081                                  |
| `make rag text="turn on the lamp"` | Helper to authenticate, submit RAG text, poll and print final decision |

Notes:

- You can run `./scripts/rag.sh` without args if you set `RAG_TEXT` (env) and `API_KEY` in `.env`.
- You can also pipe text: `echo "turn on the lamp" | ./scripts/rag.sh`
- `make rag` will use `PORT` from `.env` if present, otherwise it uses the exported `PORT` env var, and finally defaults to `8081`.
