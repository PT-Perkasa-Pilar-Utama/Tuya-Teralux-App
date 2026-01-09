# ENDPOINT: GET /api/tuya/devices/sync

## Description
Synchronizes the local device status with the latest data from Tuya Cloud.

## Test Scenarios

### 1. Sync Status (Success)
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
    "message": "Device status synced successfully"
  }
  ```
  *(Status: 200 OK)*
- **Side Effects**:
  - Local database status records are updated.
  - Device online/offline states are refreshed.

### 2. Error: Tuya Cloud Unreachable
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
    "message": "Failed to sync status"
  }
  ```
  *(Status: 500 Internal Server Error)*

### 3. Security: Unauthorized
- **URL**: `http://localhost:8080/api/tuya/devices/sync`
- **Method**: `GET`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json"
    // Missing Authorization
  }
  ```
- **Pre-conditions**: No Auth Token.
- **Expected Response**:
  ```json
  {
    "status": false,
    "message": "Unauthorized"
  }
  ```
  *(Status: 401 Unauthorized)*
