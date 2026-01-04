# ENDPOINT: PUT /api/devices/:id/status

## Description
Update the status of a specific capability (code) for a device. This triggers the actual control command (e.g., turn light on).

## Test Scenarios

### 1. Update Status (Success)
- **URL**: `http://localhost:8080/api/devices/d1/status`
- **Method**: `PUT`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Pre-conditions**: Device `d1` exists.
- **Request Body**:
  ```json
  {
    "code": "switch_1",
    "value": true
  }
  ```
- **Expected Response**:
  ```json
  {
    "status": true,
    "message": "Status updated successfully",
    "data": {
      "device_id": "d1",
      "code": "switch_1",
      "value": true
    }
  }
  ```
  *(Status: 200 OK)*
- **Side Effects**:
  - Database updated.
  - MQTT/WebSocket command sent to Teralux Hub.

### 2. Update Status (Not Found - Device)
- **URL**: `http://localhost:8080/api/devices/unknown/status`
- **Method**: `PUT`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Request Body**:
  ```json
  { "code": "switch_1", "value": true }
  ```
- **Expected Response**:
  ```json
  {
    "status": false,
    "message": "Device not found"
  }
  ```
  *(Status: 404 Not Found)*

### 3. Update Status (Not Found - Invalid Code)
- **URL**: `http://localhost:8080/api/devices/d1/status`
- **Method**: `PUT`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Pre-conditions**: `d1` is a simple sensor, does not support `nuclear_launch`.
- **Request Body**:
  ```json
  { "code": "nuclear_launch", "value": true }
  ```
- **Expected Response**:
  ```json
  {
    "status": false,
    "message": "Invalid status code for this device"
  }
  ```
  *(Status: 404 Not Found)*

### 4. Validation: Invalid Value Type
- **URL**: `http://localhost:8080/api/devices/d1/status`
- **Method**: `PUT`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Pre-conditions**: `dimmer` expects integer 0-100.
- **Request Body**:
  ```json
  { "code": "dimmer", "value": "full_power" }
  ```
- **Expected Response**:
  ```json
  {
    "status": false,
    "message": "Invalid value for status code 'dimmer'"
  }
  ```
  *(Status: 422 Unprocessable Entity)*

### 5. Security: Unauthorized
- **URL**: `http://localhost:8080/api/devices/d1/status`
- **Method**: `PUT`
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
