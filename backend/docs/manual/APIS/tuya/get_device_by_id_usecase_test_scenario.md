# ENDPOINT: GET /api/tuya/devices/:id

## Description
Retrieves detailed information for a specific Tuya device identified by its ID.

## Test Scenarios

### 1. Get Device By ID (Success)
- **URL**: `http://localhost:8080/api/tuya/devices/bf0d2...`
- **Method**: `GET`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Pre-conditions**:
  - Device with ID `bf0d2...` exists in the Tuya project.
- **Request Body**: None.
- **Expected Response**:
  ```json
  {
    "status": true,
    "message": "Device retrieved successfully",
    "data": {
       "id": "bf0d2...",
       "name": "Smart Light",
       "local_key": "abc123...",
       "category": "dj",
       "product_name": "Smart Bulb",
       "product_id": "xyz...",
       "sub": false,
       "uuid": "uuid...",
       "online": true,
       "active_time": 1678888...,
       "create_time": 1678888...,
       "update_time": 1678888...
    }
  }
  ```
  *(Status: 200 OK)*
- **Side Effects**: None.

### 2. Validation: Device Not Found
- **URL**: `http://localhost:8080/api/tuya/devices/invalid_id`
- **Method**: `GET`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Pre-conditions**: Device ID does not exist.
- **Expected Response**:
  ```json
  {
    "status": false,
    "message": "Device not found"
  }
  ```
  *(Status: 500 Internal Server Error / 404 Not Found depending on upstream)*

### 3. Security: Unauthorized
- **URL**: `http://localhost:8080/api/tuya/devices/bf0d2...`
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
