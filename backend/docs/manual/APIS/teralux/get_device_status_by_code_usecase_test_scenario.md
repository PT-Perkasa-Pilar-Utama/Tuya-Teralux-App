# ENDPOINT: GET /api/device-statuses/code/:code

## Description
Retrieve specific status entries by their capability code.

## Test Scenarios

### 1. Get Status By Code (Success)
- **URL**: `http://localhost:8080/api/device-statuses/code/switch_1?device_id=d1`
- **Method**: `GET`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: Device `d1` has status `switch_1`.
- **Expected Response**:
```json
{
  "status": true,
  "message": "Status retrieved successfully",
  "data": {
    "device_status": {
      "device_id": "d1",
      "code": "switch_1",
      "value": true,
      "updated_at": "..."
    }
  }
}
```
  *(Status: 200 OK)*

### 2. Get Status By Code (Not Found - Code)
- **URL**: `http://localhost:8080/api/device-statuses/code/unknown_code?device_id=d1`
- **Method**: `GET`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: Device `d1` exists but has no `unknown_code` status.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Status code not found for this device"
}
```
  *(Status: 404 Not Found)*

### 3. Get Status By Code (Not Found - Device)
- **URL**: `http://localhost:8080/api/device-statuses/code/switch_1?device_id=unknown`
- **Method**: `GET`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Expected Response**:
```json
{
  "status": false,
  "message": "Device not found"
}
```
  *(Status: 404 Not Found)*

### 4. Validation: Missing Device ID
- **URL**: `http://localhost:8080/api/device-statuses/code/switch_1`
- **Method**: `GET`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Expected Response**:
```json
{
  "status": false,
  "message": "Validation Error",
  "details": [
    { "field": "device_id", "message": "device_id is required" }
  ]
}
```
  *(Status: 400 Bad Request)*

### 5. Security: Unauthorized
- **URL**: `http://localhost:8080/api/device-statuses/code/switch_1`
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
