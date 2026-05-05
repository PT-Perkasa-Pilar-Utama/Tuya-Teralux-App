# Teralux Smart Home System

Welcome to the **Teralux** project source code. this repository hosts the codebase for the Teralux smart home ecosystem, designed to provide seamless control and management of smart devices.

## 📂 Repository Structure

The project is divided into two main components:

*   **`android/`**
    *   Contains the native Android mobile application source code.
    *   Built using **Kotlin** and **Jetpack Compose**.
    *   Acts as the client interface for users to interact with smart home environment.

*   **`backend/`**
    *   Contains the server-side logic and API implementation.
    *   Built using **Go (Golang)**.
    *   Manages device connectivity, state synchronization, and integration with third-party platforms like Tuya.

### Installation & User Guide
For detailed instructions on how to configure, build, and run both the Backend and Android applications (using `Makefile`), please consult the dedicated documentation.

Backend local development now uses Docker Compose v2 (`docker compose`) to start supporting containers such as MySQL before `make dev` launches the Go server.

👉 **[Read the Installation Guide](backend/docs/manual/INSTALLATIONS/1_installation_guide.md)**

## 🛠️ Key Technologies

*   **Mobile**: Android, Kotlin, Jetpack Compose
*   **Backend**: Go
*   **Database**: SQLite
*   **IoT Integration**: Tuya Smart Cloud

## 🗂️ Recent change

- ✅ **`domain/rag`** and **`domain/speech`** have been migrated from `stt-service` into `backend/domain`.
- ⚠️ `stt-service` is no longer required for these features; the code is now maintained under `backend`.

## 🔌 MQTT Configuration

### Android Client (sensio_app)

The Android app connects to the MQTT broker using **MQTTS (MQTT over TLS)** on port 8883.

**Security Model:**
- 🔒 MQTT password is **never stored** on the Android device
- 🔑 Password is fetched from backend on-demand before each MQTT connection
- 🛡️ Password is generated using **SHA256 random** by EMQX Auth Service
- 🔐 Android must authenticate via Bearer token to fetch MQTT credentials

**Configuration in `sensio_app/local.properties`:**

```properties
# MQTTS (secure, recommended for production)
mqtt.broker_url=ssl://your-broker-url:8883

# MQTT (plain, for local development only)
# mqtt.broker_url=tcp://your-broker-url:1883

# WSS (WebSocket secure, alternative)
# mqtt.broker_url=wss://your-broker-url:8084/mqtt
```

**Supported protocols:**
- `ssl://` - MQTTS (TLS encrypted, port 8883)
- `tcps://` - Alternative TCP SSL syntax
- `tcp://` - Plain MQTT (unencrypted, port 1883)
- `wss://` - MQTT over WebSocket Secure (port 8084)
- `ws://` - MQTT over WebSocket (port 8083)

### MQTT Connection Flow

```
┌──────────────┐     ┌──────────┐     ┌─────────────────┐     ┌──────────────┐
│   Android    │     │ Backend  │     │ EMQX Auth Svc   │     │ MQTT Broker  │
│              │     │  (Go)    │     │    (Rust)       │     │    (EMQX)    │
└──────┬───────┘     └────┬─────┘     └────────┬────────┘     └──────┬───────┘
       │                  │                     │                     │
       │ GET /api/mqtt/   │                     │                     │
       │ users/{user}                          │                     │
       │ (Bearer token)   │                     │                     │
       │─────────────────>│                     │                     │
       │                  │                     │                     │
       │                  │ GET /mqtt/users/{user}                    │
       │                  │ (x-api-key)         │                     │
       │                  │────────────────────>│                     │
       │                  │                     │                     │
       │                  │ {username, password}│                     │
       │                  │<────────────────────│                     │
       │                  │                     │                     │
       │ {username, password}                   │                     │
       │<─────────────────│                     │                     │
       │                  │                     │                     │
       │ CONNECT (username, password)            │                     │
       │────────────────────────────────────────>│                     │
       │                  │                     │                     │
       │                  │     /emqx/auth      │                     │
       │                  │     (verify)        │                     │
       │                  │<────────────────────│                     │
       │                  │                     │                     │
       │ CONNECTED        │                     │                     │
       │<────────────────────────────────────────│                     │
       │                  │                     │                     │
```

> ⚠️ **Security Note**: Always use `ssl://` (MQTTS) in production environments to encrypt data in transit.

