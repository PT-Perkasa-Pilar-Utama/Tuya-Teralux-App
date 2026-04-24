# Backend Architecture Documentation

## Overview

This document describes the architecture of the Sensio backend, including persistence strategies, module boundaries, and design decisions.

## Storage Boundaries

### Storage Layer Summary

```
┌────────────────────────────────────────────────────────────────┐
│                        Application Layer                        │
├────────────────────────────────────────────────────────────────┤
│  MySQL              │  BadgerDB           │  In-Memory         │
│  (Primary Store)    │  (Cache/Status)     │  (Session/State)   │
├─────────────────────┼─────────────────────┼────────────────────┤
│  • Terminals        │  • Entity cache     │  • Task progress   │
│  • Devices          │  • Task status      │  • Temp state      │
│  • Scenes           │  • Session tokens   │                    │
│  • Recordings       │  • Rate limit data  │                    │
│  • MQTT Users       │                     │                    │
│  • Device Statuses  │                     │                    │
└────────────────────────────────────────────────────────────────┘
```

### Primary Database: MySQL

**Role:** **Source of Truth** - All persistent business data.

**Purpose:** Primary persistent storage for all business entities.

**Entities stored:**
- Terminal devices (`terminals` table)
- MQTT users (`mqtt_users` table)
- Devices - Tuya integration (`devices` table)
- Scenes (`scenes` table)
- Recordings (`recordings` table)
- Device statuses (`device_statuses` table)

**Technology:** MySQL 8.0+ with GORM ORM

**Characteristics:**
- Durable, ACID-compliant
- Relational with foreign key constraints
- Indexed for fast lookups (MAC address, terminal ID, room ID)
- Soft-delete support for audit trail

**Access pattern:**
```go
// Repository pattern with caching
type ITerminalRepository interface {
    GetByID(id string) (*entities.Terminal, error)
    GetByMacAddress(macAddress string) (*entities.Terminal, error)
    Create(terminal *entities.Terminal) error
    Update(terminal *entities.Terminal) error
    Delete(id string) error
}
```

**When to use:**
- ✅ All business entity persistence
- ✅ Data that must survive restarts
- ✅ Data requiring relational queries
- ✅ Audit trails and historical data

### Cache Layer: BadgerDB

**Role:** **Performance Optimization** - NOT a source of truth.

**Purpose:** High-performance cache for frequently accessed data and transient state.

**Data stored:**
- Terminal entity cache (key: `terminal:{id}`) - **cache only, not source of truth**
- Pipeline task status (key: `pipeline:{task_id}`) - **transient, TTL-based**
- Transcription task status (key: `transcribe:{id}`) - **transient, TTL-based**
- Session state and temporary processing status

**TTL:** Cache entries are invalidated on write operations via `InvalidateCache()`

**Technology:** BadgerDB (embedded key-value store)

**Characteristics:**
- Fast key-value lookups (sub-millisecond)
- Embedded (no separate service)
- TTL support for automatic expiration
- **Data may be lost - never the sole store of critical data**

**Access pattern:**
```go
// Cache-first strategy with fallback to MySQL
func (r *TerminalRepository) GetByID(id string) (*entities.Terminal, error) {
    // 1. Try cache first
    cachedData, err := r.cache.Get("terminal:" + id)
    if err == nil {
        return unmarshal(cachedData)
    }

    // 2. Cache miss - fetch from database
    terminal, err := r.db.First(&Terminal{}, id).Error

    // 3. Populate cache
    r.cache.Set("terminal:"+id, marshal(terminal))

    return terminal, err
}
```

**When to use:**
- ✅ Frequently accessed read data
- ✅ Computed results that are expensive to regenerate
- ✅ Task status for async operations
- ✅ Session/cache data with TTL

**When NOT to use:**
- ❌ As primary data store
- ❌ For data that must persist across restarts
- ❌ For data requiring relational queries

### In-Memory State

**Role:** **Transient Processing State** - Lost on restart.

**Purpose:** Temporary state for in-progress operations.

**Examples:**
- Pipeline job progress tracking
- Upload session chunk tracking
- Rate limiting counters

**When to use:**
- ✅ Very short-lived state (seconds to minutes)
- ✅ State that can be reconstructed
- ✅ Performance-critical counters

---

## Source of Truth Matrix

| Entity              | Primary Store | Cache Strategy          | Cache Invalidation          | Notes                              |
|---------------------|---------------|-------------------------|-----------------------------|------------------------------------|
| Terminal            | MySQL         | By ID (`terminal:{id}`) | On Update/Delete            | Cache is optimization only         |
| Device              | MySQL         | None                    | N/A                         | Direct DB access                   |
| Scene               | MySQL         | None                    | N/A                         | Direct DB access                   |
| Recording           | MySQL         | None                    | N/A                         | Direct DB access                   |
| Device Status       | MySQL         | None                    | N/A                         | Direct DB access                   |
| MQTT User           | MySQL         | None                    | N/A                         | Managed by EMQX service            |
| Pipeline Task       | BadgerDB      | N/A (already in cache)  | TTL-based + manual cleanup  | Transient - status only            |
| Transcription Task  | BadgerDB      | N/A (already in cache)  | TTL-based + manual cleanup  | Transient - status only            |

**Key Principle:** MySQL is always the source of truth. BadgerDB is a cache/transient store only. If BadgerDB data is lost, it can be reconstructed from MySQL or the operation can be retried.

---

## Persistence Strategy

### Primary Database: MySQL

**Purpose:** Primary source of truth for all business entities.

**Entities stored:**
- Terminal devices
- MQTT users
- Devices (Tuya integration)
- Scenes
- Recordings
- Device statuses

**Technology:** MySQL 8.0+ with GORM ORM

**Access pattern:**
```go
// Repository pattern with caching
type ITerminalRepository interface {
    GetByID(id string) (*entities.Terminal, error)
    GetByMacAddress(macAddress string) (*entities.Terminal, error)
    Create(terminal *entities.Terminal) error
    Update(terminal *entities.Terminal) error
    Delete(id string) error
}
```

### Cache Layer: BadgerDB

**Purpose:** High-performance cache for frequently accessed entities and transient state.

**Data stored:**
- Terminal entity cache (key: `terminal:{id}`)
- Task status cache (pipeline/transcription jobs)
- Session state
- Temporary processing status

**TTL:** Cache entries are invalidated on write operations via `InvalidateCache()`

**Technology:** BadgerDB (embedded key-value store)

**Access pattern:**
```go
// Cache-first strategy with fallback to MySQL
func (r *TerminalRepository) GetByID(id string) (*entities.Terminal, error) {
    // 1. Try cache first
    cachedData, err := r.cache.Get("terminal:" + id)
    if err == nil {
        return unmarshal(cachedData)
    }
    
    // 2. Cache miss - fetch from database
    terminal, err := r.db.First(&Terminal{}, id).Error
    
    // 3. Populate cache
    r.cache.Set("terminal:"+id, marshal(terminal))
    
    return terminal, err
}
```

### Source of Truth Matrix

| Entity              | Primary Store | Cache Strategy          | Cache Invalidation          |
|---------------------|---------------|-------------------------|-----------------------------|
| Terminal            | MySQL         | By ID (`terminal:{id}`) | On Update/Delete            |
| MQTT User           | MySQL         | None                    | N/A                         |
| Pipeline Task       | BadgerDB      | N/A (already in cache)  | TTL-based + manual cleanup  |
| Transcription Task  | BadgerDB      | N/A (already in cache)  | TTL-based + manual cleanup  |
| Device              | MySQL         | None                    | N/A                         |
| Scene               | MySQL         | None                    | N/A                         |

## Module Architecture

### Domain Modules

```
backend/
├── domain/
│   ├── common/           # Shared utilities, DTOs, middleware
│   ├── infrastructure/    # DB, cache, storage, MQ infrastructure
│   ├── speech/           # AI/ML services (Whisper, LLM providers)
│   ├── notification/     # WhatsApp and push notifications
│   ├── terminal/         # Terminal management domain
│   │   ├── terminal/     # Terminal aggregate
│   │   ├── device/       # Device aggregate (Tuya)
│   │   └── device_status/# Device status tracking
│   ├── tuya/             # Tuya IoT platform integration
│   ├── scene/            # Scene automation
│   ├── recordings/       # Meeting recordings
│   ├── models/           # AI pipeline (native Go)
│   └── mail/            # Email notifications
```

### Module Dependencies

```
┌─────────────────────────────────────────────────────────────┐
│                      HTTP Layer (Gin)                        │
├─────────────────────────────────────────────────────────────┤
│                    Controllers (HTTP handlers)                │
├─────────────────────────────────────────────────────────────┤
│                    Use Cases (Business logic)                 │
├─────────────────────────────────────────────────────────────┤
│                   Repositories (Data access)                  │
├─────────────────────────────────────────────────────────────┤
│         Infrastructure (MySQL, BadgerDB, External APIs)       │
└─────────────────────────────────────────────────────────────┘
```

## Singleton Usage

### Global Singletons

The backend uses several global singletons for infrastructure components:

1. **`infrastructure.DB`** - MySQL database connection
   - **Rationale:** Single connection pool shared across all repositories
   - **Scope:** Application lifetime
   - **Thread-safe:** Yes (GORM connection pool)

2. **`utils.GetConfig()`** - Configuration accessor
   - **Rationale:** Centralized configuration management
   - **Scope:** Application lifetime
   - **Thread-safe:** Yes (read-only after initialization)

3. **`infrastructure.BadgerService`** - Cache layer
   - **Rationale:** Single cache instance shared across domains
   - **Scope:** Application lifetime
   - **Thread-safe:** Yes (BadgerDB is concurrent)

### Dependency Injection

Modules are initialized in `main.go` with explicit dependencies:

```go
// Example: Terminal module initialization
terminalRepo := terminal_repositories.NewTerminalRepository(badgerService)
terminalUseCase := terminal_usecases.NewCreateTerminalUseCase(terminalRepo, mqttService)
terminalController := terminal_controllers.NewCreateTerminalController(terminalUseCase)
```

## Migration Strategy

### Go AutoMigrate

**Used for:** MySQL schema management

**Location:** `backend/main.go`

**Entities managed:**
- Terminal
- MQTTUser
- Device
- Scene
- Recording
- DeviceStatus

**Command:**
```go
// Automatic migration on startup
infrastructure.DB.AutoMigrate(
    &entities.Terminal{},
    &entities.MQTTUser{},
    &entities.Device{},
    // ...
)
```

### Rust Migrations (EMQX Auth Service)

**Used for:** EMQX Auth Service database (SQLite)

**Location:** `backend/services/EMQX-Auth-Service/migration/`

**Rationale:**
- EMQX service is a separate Rust/Actix application
- Maintains its own database for MQTT user credentials
- Migrations are managed independently from main backend

**Operational Note:**
The coexistence of Go AutoMigrate and Rust migrations is intentional:
- Go manages the primary MySQL database
- Rust manages the EMQX auth database (SQLite)
- These are separate services with separate databases

## API Versioning

### Current Versions

- **V1 (Legacy):** `/api/models/...` - Direct Python service integration
- **V1 (Current):** `/api/models/v1/...` - gRPC/REST hybrid architecture

### Route Registration

All routes are registered in domain modules:

```go
// Example: Models V1 module
func InitModule(protected *gin.RouterGroup, cfg *utils.Config) {
    // Initialize services
    whisperGrpcSvc := whisperServices.NewGrpcWhisperService(cfg)
    pipelineSvc := pipelineServices.NewPythonPipelineService(cfg)
    
    // Initialize use cases
    pipelineUC := pipelineUsecases.NewPipelineUseCase(pipelineSvc)
    
    // Initialize controllers
    whisperCtrl := whisperControllers.NewWhisperController(whisperGrpcSvc)
    pipelineCtrl := pipelineControllers.NewPipelineController(pipelineUC)
    
    // Setup routes
    whisperRoutes.SetupWhisperRoutes(protected, whisperCtrl)
    pipelineRoutes.SetupPipelineRoutes(protected, pipelineCtrl)
}
```

## External Service Integration

### EMQX Auth Service

**Purpose:** MQTT user authentication and authorization

**Location:** `backend/services/EMQX-Auth-Service/`

**Authentication:** `x-api-key` header (all endpoints)

**Key endpoints:**
- `POST /mqtt/create` - Create MQTT user
- `GET /mqtt/users/{username}` - Get MQTT credentials
- `DELETE /mqtt/{username}` - Delete MQTT user

### Python AI Services

**Purpose:** AI/ML model inference (Whisper, RAG, Pipeline)

**Location:** `backend/services/rag-whisper-service/`

**Communication:**
- gRPC (Whisper V1)
- HTTP/REST (Pipeline, RAG V1)

## Testing Strategy

### Unit Tests

**Location:** `domain/**/*_test.go`

**Coverage:** Business logic, use cases, utilities

**Run:** `make test`

### Contract Tests

**Location:** `domain/common/utils/route_contract_test.go`

**Purpose:** Ensure OpenAPI spec matches runtime routes

**Run:** Included in `make test`

### E2E Tests

**Location:** `e2e/`

**Purpose:** Test full user flows (terminal bootstrap, etc.)

**Run:** `go test ./e2e/...`

## Security

### Authentication

- **API Key:** `X-API-KEY` header (terminal registration, public endpoints)
- **Bearer Token:** `Authorization: Bearer <token>` (authenticated user endpoints)
- **MQTT Auth:** `x-api-key` header (EMQX service-to-service)

### Authorization

- Middleware-based authorization in `domain/common/middlewares/`
- Role-based access control for admin endpoints

## Authentication Endpoints

### Terminal Login
- **Endpoint**: `POST /api/common/login`
- **Auth**: `X-API-KEY` header (API Key)
- **Purpose**: Authenticates a terminal with Tuya and returns JWT tokens
- **Request Body**:
  ```json
  { "terminal_id": "<uuid>" }
  ```
- **Response**:
  ```json
  {
    "status": true,
    "data": {
      "terminal_id": "...",
      "access_token": "...",
      "status": "renewed"
    }
  }
  ```

## Configuration

Environment variables (see `.env.example`):

```bash
# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=secret
DB_NAME=sensio

# Cache
BADGER_DIR=./data/badger

# API Keys
SENSIO_API_KEY=your-api-key
TUYA_API_KEY=your-tuya-key

# Service URLs
EMQX_AUTH_SERVICE_URL=http://localhost:8082
RAG_WHISPER_SERVICE_URL=http://localhost:8000
```

## Performance Considerations

### Caching Strategy

- Cache hit ratio target: >80% for terminal lookups
- Cache invalidation on write operations
- TTL-based expiration for task status

### Database Optimization

- Connection pooling (GORM default: 100 max connections)
- Indexed queries on MAC address, room ID, terminal ID
- Soft delete pattern for audit trail

### Async Processing

- Pipeline jobs processed asynchronously
- Status polling via `/api/models/v1/pipeline/status/{task_id}`
- BadgerDB for fast status lookups

## Monitoring

### Logging

- Structured logging via `utils.LogInfo`, `utils.LogError`, `utils.LogDebug`
- Log levels: INFO, ERROR, DEBUG

### Health Checks

- `/api/health` - Backend health
- Service-specific health endpoints for Python/Rust services

---

**Last Updated:** 2026-04-21  
**Maintainer:** Backend Team
