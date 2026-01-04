# ENDPOINT: GET /api/devices/:id/statuses

## Description
Retrieve all status entries for a specific device.

## Test Scenarios

### 1. Get Statuses By Device ID (Success)
- **URL**: `http://localhost:8080/api/devices/d1/statuses`
- **Method**: `GET`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Pre-conditions**:
  - Device `d1` exists.
  - `d1` has statuses `switch_1` (true) and `brightness` (80).
- **Expected Response**:
  ```json
  {
    "status": true,
    "message": "Statuses retrieved successfully",
    "data": [
      { "code": "switch_1", "value": true },
      { "code": "brightness", "value": 80 }
    ]
  }
  ```
  *(Status: 200 OK)*

### 2. Get Statuses By Device ID (Success - Empty)
- **URL**: `http://localhost:8080/api/devices/d2/statuses`
- **Method**: `GET`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Pre-conditions**:
  - Device `d2` exists but has no status history.
- **Expected Response**:
  ```json
  {
    "status": true,
    "data": []
  }
  ```
  *(Status: 200 OK)*

### 3. Get Statuses By Device ID (Not Found)
- **URL**: `http://localhost:8080/api/devices/unknown/statuses`
- **Method**: `GET`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Pre-conditions**: Device `unknown` does not exist.
- **Expected Response**:
  ```json
  {
    "status": false,
    "message": "Device not found"
  }
  ```
  *(Status: 404 Not Found)*

### 4. Security: Unauthorized
- **URL**: `http://localhost:8080/api/devices/d1/statuses`
- **Method**: `GET`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json"
    // Missing Authorization
  }
  ```
- **Expected Response**:
  ```json
  { "status": false, "message": "Unauthorized" }
  ```
  *(Status: 401 Unauthorized)*
