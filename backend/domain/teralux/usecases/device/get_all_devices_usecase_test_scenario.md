# ENDPOINT: GET /api/devices

## Description
Retrieve a list of devices, optionally filtered by Teralux Hub.

## Test Scenarios

### 1. Get All Devices (Success - Filter by Teralux)
- **URL**: `http://localhost:8080/api/devices?teralux_id=tx-1`
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
          "id": "d1",
          "name": "Light 1",
          "teralux_id": "tx-1",
          "remote_id": "rem-1",
          "category": "l",
          "remote_category": "light",
          "product_name": "Tuya Light",
          "remote_product_name": "Smart Light",
          "icon": "icon1",
          "custom_name": "L1",
          "model": "M1",
          "ip": "1.1.1.1",
          "local_key": "k1",
          "gateway_id": "g1",
          "create_time": 100,
          "update_time": 100,
          "collections": "[]",
          "created_at": "2024-01-01T00:00:00Z",
          "updated_at": "2024-01-01T00:00:00Z"
        },
        {
          "id": "d2",
          "name": "Light 2",
          "teralux_id": "tx-1",
          "remote_id": "rem-2",
          "category": "l",
          "remote_category": "light",
          "product_name": "Tuya Light",
          "remote_product_name": "Smart Light",
          "icon": "icon2",
          "custom_name": "L2",
          "model": "M2",
          "ip": "1.1.1.2",
          "local_key": "k2",
          "gateway_id": "g1",
          "create_time": 101,
          "update_time": 101,
          "collections": "[]",
          "created_at": "2024-01-01T00:00:00Z",
          "updated_at": "2024-01-01T00:00:00Z"
        }
      ],
      "total": 2,
      "page": 1,
      "per_page": 2
    }
  }
  ```
  *(Status: 200 OK)*

### 1.1 Get All Devices (Success - Pagination)
- **URL**: `http://localhost:8080/api/devices?page=1&limit=1`
- **Method**: `GET`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Pre-conditions**:
  - `tx-1` has 2 devices (d1, d2).
- **Expected Response**:
  ```json
  {
    "status": true,
    "message": "Devices retrieved successfully",
    "data": {
      "devices": [
        {
          "id": "d1",
          "name": "Light 1",
          // ... details ...
        }
      ],
      "total": 2,
      "page": 1,
      "per_page": 1
    }
  }
  ```
  *(Status: 200 OK)*

### 2. Get All Devices (Success - Empty)
- **URL**: `http://localhost:8080/api/devices`
- **Method**: `GET`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Pre-conditions**: No devices in system.
- **Expected Response**:
  ```json
  {
    "status": true,
    "data": { "devices": [], "total": 0, "page": 1, "per_page": 0 }
  }
  ```
  *(Status: 200 OK)*

### 3. Security: Unauthorized
- **URL**: `http://localhost:8080/api/devices`
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
  { "status": false, "message": "Unauthorized" }
  ```
  *(Status: 401 Unauthorized)*
