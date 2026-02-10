# ENDPOINT: GET /api/tuya/devices/sync

## Description
Fetches real-time status from Tuya Cloud. **Does NOT update local database**. Returns fresh device list with current status. This is a read-only operation that provides the latest device states directly from Tuya.

## Authentication
- **Type**: ApiKeyAuth OR BearerAuth (both accepted)
- **Headers**: 
  - Option 1: `X-API-KEY: <api_key>`
  - Option 2: `Authorization: Bearer <token>`

## Test Scenarios

### 1. Sync Status (Success - Using Bearer Token)
- **URL**: `http://localhost:8080/api/tuya/devices/sync`
- **Method**: `GET`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: Backend is connected to Tuya Cloud.
- **Request Body**: None.
- **Expected Response**:
```json
{
  "status": true,
  "message": "Device status synced successfully",
  "data": [
    {
      "id": "bf0d2...",
      "name": "Smart Light",
      "online": true,
      "create_time": 1678888000,
      "status": [
        {"code": "switch_1", "value": true},
        {"code": "bright_value", "value": 255}
      ]
    },
    {
      "id": "bf1a3...",
      "name": "Living Room Switch",
      "online": false,
      "create_time": 1678888100,
      "status": [
        {"code": "switch_1", "value": false}
      ]
    }
  ]
}
```
  *(Status: 200 OK)*
- **Side Effects**:
  - **NONE** - Does NOT update local database
  - Returns fresh data from Tuya Cloud only
  - Device online/offline states are fetched in real-time

### 2. Sync Status (Success - Using API Key)
- **URL**: `http://localhost:8080/api/tuya/devices/sync`
- **Method**: `GET`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "X-API-KEY": "<api_key>"
}
```
- **Pre-conditions**: Backend is connected to Tuya Cloud, valid API key.
- **Request Body**: None.
- **Expected Response**:
```json
{
  "status": true,
  "message": "Device status synced successfully",
  "data": [/* array of devices with current status */]
}
```
  *(Status: 200 OK)*
- **Side Effects**: None (read-only).

### 3. Error: Tuya Cloud Unreachable
- **URL**: `http://localhost:8080/api/tuya/devices/sync`
- **Method**: `GET`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: Tuya API is down or credentials invalid.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Failed to sync status from Tuya Cloud"
}
```
  *(Status: 500 Internal Server Error)*

### 4. Security: Unauthorized (No Auth)
- **URL**: `http://localhost:8080/api/tuya/devices/sync`
- **Method**: `GET`
- **Headers**:
```json
{
  "Content-Type": "application/json"
  // Missing both X-API-KEY and Authorization
}
```
- **Pre-conditions**: No authentication provided.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Unauthorized"
}
```
  *(Status: 401 Unauthorized)*

### 5. Security: Invalid Credentials
- **URL**: `http://localhost:8080/api/tuya/devices/sync`
- **Method**: `GET`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer invalid_token_12345"
}
```
- **Pre-conditions**: Invalid Bearer token provided.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Unauthorized"
}
```
  *(Status: 401 Unauthorized)*
