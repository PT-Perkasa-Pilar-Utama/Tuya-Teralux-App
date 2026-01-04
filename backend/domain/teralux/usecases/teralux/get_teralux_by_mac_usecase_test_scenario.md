# ENDPOINT: GET /api/teralux/mac/:mac_address

## Description
Retrieve a single Teralux device by its MAC address.

## Test Scenarios

### 1. Get Teralux By MAC (Success)
- **URL**: `http://localhost:8080/api/teralux/mac/AA:BB:CC:11:22:33`
- **Method**: `GET`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Pre-conditions**: Device with MAC `AA:BB:CC:11:22:33` exists.
- **Expected Response**:
  ```json
  {
    "status": true,
    "message": "Teralux retrieved successfully",
    "data": {
      "id": "<uuid>",
      "name": "Living Room",
      "mac_address": "AA:BB:CC:11:22:33",
      "room_id": "r1"
    }
  }
  ```
  *(Status: 200 OK)*

### 2. Get Teralux By MAC (Not Found)
- **URL**: `http://localhost:8080/api/teralux/mac/XX:YY:ZZ:00:00:00`
- **Method**: `GET`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Pre-conditions**: No device with this MAC exists.
- **Expected Response**:
  ```json
  {
    "status": false,
    "message": "Teralux not found",
    "data": null
  }
  ```
  *(Status: 404 Not Found)*

### 3. Validation: Invalid MAC Format
- **URL**: `http://localhost:8080/api/teralux/mac/INVALID-MAC`
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
    "message": "Invalid MAC address format"
  }
  ```
  *(Status: 400 Bad Request)*

### 4. Security: Unauthorized
- **URL**: `http://localhost:8080/api/teralux/mac/AA:BB:CC`
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
  {
    "status": false,
    "message": "Unauthorized"
  }
  ```
  *(Status: 401 Unauthorized)*
