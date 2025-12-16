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
3. Expose port `8080`.

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
| `make test` | Run all unit tests |
| `make clean` | Clean build artifacts |
| `make kill` | Kill any process running on port 8080 |
