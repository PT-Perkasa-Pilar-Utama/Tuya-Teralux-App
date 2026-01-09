# ENDPOINT: POST /api/tuya/devices/:id/commands/*

## Description
Controls Tuya devices by sending commands. Supports standard switch controls and IR remote commands (AC, TV, etc).

## Test Scenarios

### 1. Control Switch (Success)
- **URL**: `http://localhost:8080/api/tuya/devices/:id/commands/switch`
- **Method**: `POST`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Pre-conditions**: Device supports switch commands (e.g., lights, breakers).
- **Request Body**:
  ```json
  {
    "value": true
  }
  ```
- **Expected Response**:
  ```json
  {
    "status": true,
    "message": "Command sent successfully",
    "data": true
  }
  ```
  *(Status: 200 OK)*
- **Side Effects**: Physical device turns ON.

### 2. Control IR Device (AC/TV) (Success)
- **URL**: `http://localhost:8080/api/tuya/devices/:id/commands/ir`
- **Method**: `POST`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Pre-conditions**: Device is an IR Blaster and `remote_id` is configured.
- **Request Body**:
  ```json
  {
    "remote_id": "remote_123",
    "key": "PowerOn"
  }
  ```
- **Expected Response**:
  ```json
  {
    "status": true,
    "message": "IR command sent successfully"
  }
  ```
  *(Status: 200 OK)*
- **Side Effects**: IR Blaster emits signal.

### 3. Validation: Missing Parameters
- **URL**: `http://localhost:8080/api/tuya/devices/:id/commands/switch`
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
  {} 
  ```
- **Expected Response**:
  ```json
  {
    "status": false,
    "message": "Invalid request body"
  }
  ```
  *(Status: 400 Bad Request)*
