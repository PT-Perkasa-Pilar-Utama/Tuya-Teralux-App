# ENDPOINT: GET /api/devices

## Description
Retrieve a list of devices, optionally filtered by Teralux Hub.

## Test Scenarios

### 1. Get All Devices (Success - Filter by Teralux)
- **URL**: `http://localhost:8080/api/devices?teralux_id=tx-1`
- **Method**: `GET`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Pre-conditions**:
  - `tx-1` has 2 devices.
- **Expected Response**:
  ```json
  {
    "status": true,
    "message": "Devices retrieved successfully",
    "data": {
      "devices": [
        { "id": "d1", "name": "Light 1", "teralux_id": "tx-1" },
        { "id": "d2", "name": "Light 2", "teralux_id": "tx-1" }
      ],
      "total": 2
    }
  }
  ```
  *(Status: 200 OK)*

### 2. Get All Devices (Success - Empty)
- **URL**: `http://localhost:8080/api/devices`
- **Method**: `GET`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Pre-conditions**: No devices in system.
- **Expected Response**:
  ```json
  {
    "status": true,
    "data": { "devices": [], "total": 0 }
  }
  ```
  *(Status: 200 OK)*

### 3. Security: Unauthorized
- **URL**: `http://localhost:8080/api/devices`
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
  { "status": false, "message": "Unauthorized" }
  ```
  *(Status: 401 Unauthorized)*
