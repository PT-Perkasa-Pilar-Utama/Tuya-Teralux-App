# ENDPOINT: GET /api/teralux/:id

## Description
Retrieve a single Teralux device by its unique ID.

## Test Scenarios

### 1. Get Teralux By ID (Success)
- **URL**: `http://localhost:8080/api/teralux/t1`
- **Method**: `GET`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: Device `t1` exists.
- **Expected Response**:
```json
{
  "status": true,
  "message": "Teralux retrieved successfully",
  "data": {
    "teralux": {
      "id": "t1",
      "name": "Living Room",
      "mac_address": "AA:BB",
      "room_id": "r1",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  }
}
```
  *(Status: 200 OK)*

### 2. Get Teralux By ID (Not Found)
- **URL**: `http://localhost:8080/api/teralux/unknown-id`
- **Method**: `GET`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: Device `unknown-id` does not exist.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Teralux not found",
  "data": null
}
```
  *(Status: 404 Not Found)*

### 3. Validation: Invalid ID Format
- **URL**: `http://localhost:8080/api/teralux/INVALID-FORMAT`
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
    { "field": "id", "message": "Invalid ID format" }
  ]
}
```
  *(Status: 400 Bad Request)*

### 4. Security: Unauthorized
- **URL**: `http://localhost:8080/api/teralux/t1`
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
