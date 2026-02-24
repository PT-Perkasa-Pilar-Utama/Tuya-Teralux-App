# ENDPOINT: POST /api/devices

## Description
Register a new sub-device (e.g., switch, sensor) under a Teralux Controller.

## Test Scenarios

### 1. Create Device (Success)
- **URL**: `http://localhost:8080/api/devices`
- **Method**: `POST`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**:
  - Teralux Hub `tx-1` exists.
- **Request Body**:
```json
{
  "id": "tuya-device-123",
  "name": "Kitchen Light 1",
  "teralux_id": "tx-1"
}
```
- **Expected Response**:
```json
{
  "status": true,
  "message": "Device created successfully",
  "data": {
    "device_id": "<tuya-device-id>"
  }
}
```
  *(Status: 201 Created)*
- **Side Effects**:
  - Device record created.
  - Default status entries might be initialized.

### 2. Validation: Missing Required Fields
- **URL**: `http://localhost:8080/api/devices`
- **Method**: `POST`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: None.
- **Request Body**:
```json
{
  "name": "",
  "teralux_id": ""
}
```
- **Expected Response**:
```json
{
  "status": false,
  "message": "Validation Error",
  "details": [
    { "field": "name", "message": "name is required" },
    { "field": "teralux_id", "message": "teralux_id is required" }
  ]
}
```
  *(Status: 422 Unprocessable Entity)*

### 3. Constraint: Invalid Teralux ID
- **URL**: `http://localhost:8080/api/devices`
- **Method**: `POST`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: `tx-999` does not exist.
- **Request Body**:
```json
{
  "name": "Ghost Device",
  "teralux_id": "tx-999"
}
```
- **Expected Response**:
```json
{
  "status": false,
  "message": "Validation Error",
  "details": [
    { "field": "teralux_id", "message": "Teralux hub does not exist" }
  ]
}
```
  *(Status: 422 Unprocessable Entity)*

### 4. Security: Unauthorized
- **URL**: `http://localhost:8080/api/devices`
- **Method**: `POST`
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
