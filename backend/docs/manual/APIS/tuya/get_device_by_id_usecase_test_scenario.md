# ENDPOINT: GET /api/tuya/devices/{id}

## Description
Retrieves detailed information for a specific Tuya device identified by its ID.

## Authentication
- **Type**: BearerAuth
- **Header**: `Authorization: Bearer <token>`

## Path Parameters
- `id` (string, required) - Device ID

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
    "device": {
      "id": "bf0d2...",
      "name": "Smart Light",
      "local_key": "abc123...",
      "category": "dj",
      "product_name": "Smart Bulb",
      "product_id": "xyz...",
      "sub": false,
      "uuid": "uuid...",
      "online": true,
      "active_time": 1678888000,
      "create_time": 1678888000,
      "update_time": 1678888000,
      "status": [
        {"code": "switch_1", "value": true},
        {"code": "bright_value", "value": 255}
      ],
      "icon": "https://...",
      "ip": "192.168.1.100",
      "model": "model_123",
      "gateway_id": "",
      "collections": []
    }
  }
}
```
  *(Status: 200 OK)*
- **Side Effects**: None (read-only).

### 2. Validation: Device Not Found
- **URL**: `http://localhost:8080/api/tuya/devices/invalid_id_12345`
- **Method**: `GET`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: Device ID does not exist in Tuya.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Device not found"
}
```
  *(Status: 500 Internal Server Error)*

### 3. Validation: Invalid ID Format
- **URL**: `http://localhost:8080/api/tuya/devices/`
- **Method**: `GET`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: Empty or malformed device ID.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Validation Error",
  "details": [
    { "field": "id", "message": "Invalid device ID format" }
  ]
}
```
  *(Status: 400 Bad Request)*

### 4. Security: Unauthorized
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

### 5. Error: Server Error
- **URL**: `http://localhost:8080/api/tuya/devices/bf0d2...`
- **Method**: `GET`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: Tuya Cloud service is down.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Failed to retrieve device"
}
```
  *(Status: 500 Internal Server Error)*
