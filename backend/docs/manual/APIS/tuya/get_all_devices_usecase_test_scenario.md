# ENDPOINT: GET /api/tuya/devices

## Description
Retrieves a list of all devices associated with the Tuya project/user account.

## Test Scenarios

### 1. Get All Devices (Success)
- **URL**: `http://localhost:8080/api/tuya/devices`
- **Method**: `GET`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Pre-conditions**:
  - Tuya account is linked and devices are synchronized.
- **Request Body**: None.
- **Expected Response**:
  ```json
  {
    "status": true,
    "message": "Devices retrieved successfully",
    "data": [
      {
        "id": "bf0d2...",
        "name": "Smart Light",
        "category": "dj",
        "product_name": "Smart Bulb",
        "online": true
      },
      {
        "id": "bf1a3...",
        "name": "Living Room Switch",
        "category": "kg",
        "product_name": "Zigbee Switch",
        "online": false
      }
    ]
  }
  ```
  *(Status: 200 OK)*
- **Side Effects**: None.

### 2. Security: Unauthorized
- **URL**: `http://localhost:8080/api/tuya/devices`
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
