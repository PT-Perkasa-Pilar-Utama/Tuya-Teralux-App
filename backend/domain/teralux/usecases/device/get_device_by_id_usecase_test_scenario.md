# ENDPOINT: GET /api/devices/:id

## Description
Retrieve a single device by its ID.

## Test Scenarios

### 1. Get Device By ID (Success)
- **URL**: `http://localhost:8080/api/devices/dev-1`
- **Method**: `GET`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Pre-conditions**: Device `dev-1` exists.
- **Expected Response**:
  ```json
  {
    "status": true,
    "message": "Device retrieved successfully",
    "data": {
      "id": "dev-1",
      "name": "Kitchen Switch",
      "teralux_id": "tx-1",
      "type": "switch",
      "status": [...] // Optional current status
    }
  }
  ```
  *(Status: 200 OK)*

### 2. Get Device By ID (Not Found)
- **URL**: `http://localhost:8080/api/devices/dev-unknown`
- **Method**: `GET`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Pre-conditions**: Device does not exist.
- **Expected Response**:
  ```json
  {
    "status": false,
    "message": "Device not found"
  }
  ```
  *(Status: 404 Not Found)*

### 3. Security: Unauthorized
- **URL**: `http://localhost:8080/api/devices/dev-1`
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
