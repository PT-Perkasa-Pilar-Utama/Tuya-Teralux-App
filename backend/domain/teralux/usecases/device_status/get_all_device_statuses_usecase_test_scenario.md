# ENDPOINT: GET /api/device-statuses

## Description
Retrieve all device statuses across the system. Mostly used for administrative monitoring or analytics.

## Test Scenarios

### 1. Get All Statuses (Success)
- **URL**: `http://localhost:8080/api/device-statuses`
- **Method**: `GET`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Pre-conditions**: System has tracked statuses.
- **Expected Response**:
  ```json
  {
    "status": true,
    "message": "Statuses retrieved successfully",
    "data": {
      "statuses": [
        { "id": "s1", "device_id": "d1", "code": "switch_1", "value": true },
        { "id": "s2", "device_id": "d1", "code": "dimmer", "value": 50 }
      ],
      "total": 2
    }
  }
  ```
  *(Status: 200 OK)*

### 2. Get All Statuses (Empty)
- **URL**: `http://localhost:8080/api/device-statuses`
- **Method**: `GET`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Pre-conditions**: No statuses logged.
- **Expected Response**:
  ```json
  {
    "status": true,
    "data": { "statuses": [], "total": 0 }
  }
  ```
  *(Status: 200 OK)*

### 3. Security: Unauthorized
- **URL**: `http://localhost:8080/api/device-statuses`
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
  { "status": false, "message": "Unauthorized" }
  ```
  *(Status: 401 Unauthorized)*
