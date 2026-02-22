# ENDPOINT: GET /api/devices/teralux/:teralux_id

## Description
Retrieve all devices linked to a specific Teralux ID.

## Test Scenarios

### 1. Get Devices By Teralux ID (Success)
- **URL**: `http://localhost:8080/api/devices/teralux/tx-1`
- **Method**: `GET`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**:
  - Devices exist for `tx-1`.
- **Expected Response**:
```json
{
  "status": true,
  "message": "Devices retrieved successfully",
  "data": {
    "devices": [
      {
        "id": "dev-1",
        "teralux_id": "tx-1",
        "name": "Light 1",
        "category": "dj",
        "create_time": 1700000000,
        "update_time": 1700000000,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "per_page": 1
  }
}
```
  *(Status: 200 OK)*

### 1.1 Get Devices By Teralux ID (Success - Pagination)
- **URL**: `http://localhost:8080/api/devices/teralux/tx-1?page=1&limit=1`
- **Method**: `GET`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**:
  - `tx-1` has 2 devices.
- **Expected Response**:
```json
{
  "status": true,
  "message": "Devices retrieved successfully",
  "data": {
    "devices": [
      {
        "id": "dev-1",
        // ...
      }
    ],
    "total": 2,
    "page": 1,
    "per_page": 1
  }
}
```
  *(Status: 200 OK)*

### 2. Get Devices By Teralux ID (Success - Empty)
- **URL**: `http://localhost:8080/api/devices/teralux/tx-empty`
- **Method**: `GET`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**:
  - No devices exist for `tx-empty`.
- **Expected Response**:
```json
{
  "status": true,
  "message": "Devices retrieved successfully",
  "data": {
    "devices": [],
    "total": 0,
    "page": 1,
    "per_page": 0
  }
}
```
  *(Status: 200 OK)*

### 3. Get Devices By Teralux ID (Not Found - Teralux ID)
- **URL**: `http://localhost:8080/api/devices/teralux/tx-999`
- **Method**: `GET`
- **Expected Response**:
```json
{
  "status": false,
  "message": "Teralux hub not found"
}
```
  *(Status: 404 Not Found)*

### 4. Validation: Invalid Teralux ID Format
- **URL**: `http://localhost:8080/api/devices/teralux/INVALID`
- **Method**: `GET`
- **Expected Response**:
```json
{
  "status": false,
  "message": "Validation Error",
  "details": [
    { "field": "teralux_id", "message": "Invalid Teralux ID format" }
  ]
}
```
  *(Status: 400 Bad Request)*

### 4. Security: Unauthorized
- **URL**: `http://localhost:8080/api/devices/teralux/tx-1`
- **Method**: `GET`
- **Headers**: `Authorization` missing
- **Expected Response**:
```json
{
  "status": false,
  "message": "Unauthorized"
}
```
  *(Status: 401 Unauthorized)*
