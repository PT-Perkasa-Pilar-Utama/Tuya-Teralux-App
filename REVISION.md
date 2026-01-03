# REVISION.md - Comprehensive Refactor Summary

This document provides a full overview of all changes made during this session to the Teralux App Backend and requirements for the Android migration.

## 1. Database & Schema Changes

### DeviceStatus Refactor
- **Composite Primary Key**: Removed the standalone `id` field from `DeviceStatus`. It now uses a composite primary key consisting of `(device_id, code)`.
- **Foreign Key Migration**: Updated migration files and entities to reflect the composite key structure.
- **Tuya Fields**: Added comprehensive metadata fields to the `devices` table to store Tuya response data (Category, Model, IP, LocalKey, etc.).

## 2. API & Route Changes

### Route Standardization
- **Pluralization**: Standardized all DeviceStatus routes to use `/api/devices/statuses` (changed from `/api/device/statuses`).
- **Endpoint Cleanup**: 
    - Removed **independent** `POST` and `DELETE` routes for `DeviceStatus`.
    - Kept: `GET /`, `GET /:deviceId`, `GET /:deviceId/:code`, and `PUT /:deviceId/:code`.

### Create Device Automation
- **format**: `POST /api/devices`
- **Simplified Payload**: Removed manual `status` array. Now only requires Tuya `id`, `teralux_id`, and `name`.
- **Auto-Sync**: Backend now automatically fetches all device details and its 20+ status codes directly from Tuya during creation.

## 3. Logic & Business Rules

### Automated Lifecycle
- **Create**: statuses are now automatically populated from Tuya API response.
- **Delete**: implemented cascading soft-delete. Deleting a device now automatically cleans up its associated statuses.
- **Update**: statuses are upserted (Add if new, Update if exists) to prevent duplicates or orphaned records.

### Quality & Security
- **Security**: Disabled verbose GORM logging in production settings to prevent sensitive SQL queries from appearing in logs.
- **Documentation**: Fixed Swagger UI 404 and asset loading issues. Documentation now accurately reflects the pluralized and automated API.
- **Verification**: Implemented 31 unit tests covering all new business logic.

---

## 4. Android Migration Guide (Checklist)

### 🛑 Breaking Changes
1. **DeviceStatus ID**: Do **NOT** use `id` for device statuses. Identify them using `device_id` and `code`.
2. **Create Payload**: Update `CreateDeviceRequestDTO` to remove `status` and add `id` (Tuya Device ID).

### 🛠 Model Updates
- **`Device.kt`**: Add these fields to match backend response:
    - `remote_id`, `category`, `remote_category`, `product_name`, `icon`, `model`, `ip`, `local_key`, `gateway_id`.
- **`DeviceStatus.kt`**: Remove `id` field. Ensure `Value` is handled as a `String` (backend stores everything as string for flexibility).

### 🌐 API Interface Updates
- Update base path for status endpoints to `/api/devices/statuses`.
- Update `CreateDevice` call:
    ```kotlin
    // Old
    createDevice(TeraluxID, Name, List<Status>)
    // New
    createDevice(TuyaID, TeraluxID, Name)
    ```

### 📱 UI Updates
- You no longer need to pass captured statuses from the Tuya scan flow to the backend; simply passing the Tuya ID is enough.
- Use `GET /api/devices/statuses/{deviceId}` to refresh the status list on the device detail screen.

---
**Status**: All backend changes are verified via E2E tests and strictly follow the Tuya response structure.
