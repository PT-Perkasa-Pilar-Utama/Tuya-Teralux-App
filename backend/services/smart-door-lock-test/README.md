# Tuya Smart Door Lock - Test Suite

Clean Architecture implementation for testing Tuya smart door lock devices.

## Quick Start

### 1. Configure Credentials

Copy `.env.example` to `.env` and fill in your credentials:

```bash
cp .env.example .env
```

```env
TUYA_CLIENT_ID=your_client_id
TUYA_ACCESS_SECRET=your_access_secret
TUYA_BASE_URL=https://openapi-sg.iotbing.com
TUYA_USER_ID=your_user_id
TUYA_DEVICE_ID=your_device_id
```

### 2. Run CLI

```bash
cd cmd/cli
go run main.go
```

## Features

| Feature | Status | Notes |
|---------|--------|-------|
| Get Device Status | ✅ Working | |
| Lock/Unlock Door | ✅ Working | Requires device online |
| Generate Dynamic Password | ⚠️ API Required | Need subscription in Tuya IoT |
| Generate Temporary Password | ⚠️ API Required | Need subscription in Tuya IoT |

## Project Structure

```
smart-door-lock-test/
├── cmd/
│   └── cli/              # CLI application
│       └── main.go
├── internal/
│   ├── config/           # Configuration management
│   ├── domain/           # Business entities
│   ├── repository/       # External API layer
│   └── service/          # Business logic
├── .env                  # Credentials (DO NOT COMMIT)
├── .env.example          # Template
└── go.mod
```

## Architecture

Follows **Clean Architecture** principles:

- **Domain Layer**: Pure business entities (Device, Password, Command)
- **Repository Layer**: Tuya API communication
- **Service Layer**: Business logic orchestration
- **Handler Layer**: CLI interaction

### Design Principles

- **SOLID**: Single Responsibility, Open/Closed, Liskov Substitution, Interface Segregation, Dependency Inversion
- **DRY**: Centralized signature generation, authentication, error handling
- **Clean Architecture**: Dependencies point inward, easy to test

## API Reference

### Endpoints Used

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/v1.0/token?grant_type=1` | GET | Get access token |
| `/v1.0/devices/{id}` | GET | Get device info |
| `/v1.0/devices/{id}/specifications` | GET | Get device functions |
| `/v1.0/devices/{id}/commands` | POST | Send commands |
| `/v1.0/devices/{id}/door-lock/dynamic-password` | GET | Generate dynamic password* |
| `/v1.0/devices/{id}/door-lock/temp-password` | POST | Generate temporary password* |

*Requires API subscription

### Device Status Codes

| Code | Description |
|------|-------------|
| `lock_motor_state` | true=unlocked, false=locked |
| `battery_state` | high/medium/low/poweroff |
| `alarm_lock` | Alarm type (wrong_finger, wrong_password, etc.) |
| `unlock_fingerprint` | Fingerprint unlock count |
| `unlock_password` | Password unlock count |
| `unlock_app` | App unlock count |
| `doorbell` | Doorbell pressed |
| `hijack` | Duress alarm |

## Troubleshooting

### "sign invalid (code: 1004)"
- Check Client ID and Access Secret are correct
- Ensure signature uses correct content hash (SHA256 of empty string for GET)

### "No permissions. This API is not subscribed (code: 28841101)"
- Subscribe to Door Lock API in Tuya IoT Platform
- Go to: Cloud > API > Subscribe API > Smart Lock

### "device is offline (code: 2001)"
- Device must be connected to WiFi
- Check device status in Smart Life app

## Testing

```bash
# Build
go build ./...

# Run CLI
go run cmd/cli/main.go

# Run tests (when available)
go test ./...
```

## Device Info

**Test Device:** U688S-WiFi-Pro  
**Category:** `jtmspro`  
**Device ID:** `a3621a5ad61e644d91aaa2`

---

**Last Updated:** March 27, 2026
