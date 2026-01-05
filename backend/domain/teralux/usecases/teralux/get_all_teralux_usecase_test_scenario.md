# ENDPOINT: GET /api/teralux

## Description
Retrieve a paginated list of all Teralux devices, with optional filtering.

## Test Scenarios

### 1. Get All Teralux (Success - Empty List)
- **URL**: `http://localhost:8080/api/teralux`
- **Method**: `GET`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Pre-conditions**: Database is empty.
- **Expected Response**:
  ```json
  {
    "status": true,
    "message": "Teralux list retrieved successfully",
    "data": {
      "teralux": [],
      "total": 0,
      "page": 1,
      "per_page": 10
    }
  }
  ```
  *(Status: 200 OK)*

### 2. Get All Teralux (Success - With Data)
- **URL**: `http://localhost:8080/api/teralux`
- **Method**: `GET`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Pre-conditions**: 2 devices make up the dataset.
- **Expected Response**:
  ```json
  {
    "status": true,
    "message": "Teralux list retrieved successfully",
    "data": {
      "teralux": [
        { 
          "id": "t1", "name": "Hub 1", "mac_address": "M1", "room_id": "r1",
          "created_at": "2024-01-01T00:00:00Z", "updated_at": "2024-01-01T00:00:00Z" 
        },
        { 
          "id": "t2", "name": "Hub 2", "mac_address": "M2", "room_id": "r2",
          "created_at": "2024-01-01T00:00:00Z", "updated_at": "2024-01-01T00:00:00Z"
        }
      ],
      "total": 2,
      "page": 1,
      "per_page": 10
    }
  }
  ```
  *(Status: 200 OK)*

### 3. Pagination: Limit and Page
- **URL**: `http://localhost:8080/api/teralux?page=2&limit=5`
- **Method**: `GET`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Pre-conditions**: 15 devices exist.
- **Expected Response**:
  ```json
  {
    "status": true,
    "message": "Teralux list retrieved successfully",
    "data": {
      "teralux": [ ... ], // 5 items
      "total": 15,
      "page": 2,
      "per_page": 5
    }
  }
  ```
  *(Status: 200 OK)*

### 4. Filter: By Room ID
- **URL**: `http://localhost:8080/api/teralux?room_id=room-101`
- **Method**: `GET`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Pre-conditions**:
  - 3 devices in `room-101`.
  - 2 devices in `room-102`.
- **Expected Response**:
  ```json
  {
    "status": true,
    "message": "Teralux list retrieved successfully",
    "data": {
      "teralux": [ ... ], // Only devices with room_id=room-101
      "total": 3
    }
  }
  ```
  *(Status: 200 OK)*

### 5. Security: Unauthorized
- **URL**: `http://localhost:8080/api/teralux`
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
