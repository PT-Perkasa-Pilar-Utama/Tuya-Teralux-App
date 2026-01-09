# ENDPOINT: POST /api/teralux

## Description
Create a new Teralux device. This endpoint handles the registration of a new central hub or controller in a specific room.

## Test Scenarios

### 1. Create Teralux (Success Condition)
- **URL**: `http://localhost:8080/api/teralux`
- **Method**: `POST`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Pre-conditions**:
  - `room-101` exists.
- **Request Body**:
  ```json
  {
    "name": "Master Bedroom Hub",
    "mac_address": "AA:BB:CC:11:22:33",
    "room_id": "room-101"
  }
  ```
- **Expected Response**:
  ```json
  {
    "status": true,
    "message": "Teralux created successfully",
    "data": {
      "teralux_id": "<uuid>"
    }
  }
  ```
  *(Status: 201 Created)*
- **Side Effects**:
  - New Teralux record created in database.

### 2. Validation: Empty Fields
- **URL**: `http://localhost:8080/api/teralux`
- **Method**: `POST`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Request Body**:
  ```json
  {
    "name": "",
    "mac_address": "",
    "room_id": ""
  }
  ```
- **Expected Response**:
  ```json
  {
    "status": false,
    "message": "Validation Error",
    "details": [
      { "field": "name", "message": "name is required" },
      { "field": "mac_address", "message": "mac_address is required" },
      { "field": "room_id", "message": "room_id is required" }
    ]
  }
  ```
  *(Status: 422 Unprocessable Entity)*

### 3. Validation: Invalid MAC Address Format
- **URL**: `http://localhost:8080/api/teralux`
- **Method**: `POST`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Request Body**:
  ```json
  {
    "name": "Living Room",
    "mac_address": "INVALID-MAC-ADDRESS",
    "room_id": "room-101"
  }
  ```
- **Expected Response**:
  ```json
  {
    "status": false,
    "message": "Validation Error",
    "details": [
      { "field": "mac_address", "message": "invalid mac address format" }
    ]
  }
  ```
  *(Status: 422 Unprocessable Entity)*

### 4. Validation: Name Too Long
- **URL**: `http://localhost:8080/api/teralux`
- **Method**: `POST`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Request Body**:
  ```json
  {
    "name": "<string with 256+ characters>",
    "mac_address": "AA:BB:CC:11:22:33",
    "room_id": "room-101"
  }
  ```
- **Expected Response**:
  ```json
  {
    "status": false,
    "message": "Validation Error",
    "details": [
      { "field": "name", "message": "name must be 255 characters or less" }
    ]
  }
  ```
  *(Status: 422 Unprocessable Entity)*

### 5. Idempotent: Duplicate MAC Address Returns Existing ID
- **URL**: `http://localhost:8080/api/teralux`
- **Method**: `POST`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Pre-conditions**:
  - Device with MAC `AA:BB:CC:11:22:33` already exists with ID `existing-id`.
- **Request Body**:
  ```json
  {
    "name": "Duplicate Hub",
    "mac_address": "AA:BB:CC:11:22:33",
    "room_id": "room-102"
  }
  ```
- **Expected Response**:
  ```json
  {
    "status": true,
    "message": "Teralux created successfully",
    "data": {
      "teralux_id": "existing-id"
    }
  }
  ```
  *(Status: 200 OK)*
- **Side Effects**:
  - No new record created (idempotent for device booting).

### 6. Security: Unauthorized
- **URL**: `http://localhost:8080/api/teralux`
- **Method**: `POST`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json"
    // Missing Authorization
  }
  ```
- **Request Body**: Valid JSON body.
- **Expected Response**:
  ```json
  {
    "status": false,
    "message": "Unauthorized"
  }
  ```
  *(Status: 401 Unauthorized)*
