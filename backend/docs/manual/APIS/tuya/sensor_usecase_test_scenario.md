# ENDPOINT: GET /api/tuya/devices/:id/sensor

## Description
Retrieves formatted sensor data (e.g., Temperature, Humidity) for a specific device.

## Test Scenarios

### 1. Get Sensor Data (Success)
- **URL**: `http://localhost:8080/api/tuya/devices/:id/sensor`
- **Method**: `GET`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Pre-conditions**: Device is a sensor and online.
- **Request Body**: None.
- **Expected Response**:
  ```json
  {
    "status": true,
    "message": "Sensor data retrieved successfully",
    "data": {
      "temperature": 25.5,
      "humidity": 60,
      "battery": 80
    }
  }
  ```
  *(Status: 200 OK)*
- **Side Effects**: None.

### 2. Validation: Not a Sensor
- **URL**: `http://localhost:8080/api/tuya/devices/:id/sensor`
- **Method**: `GET`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Pre-conditions**: Device ID refers to a light bulb, not a sensor.
- **Expected Response**:
  ```json
  {
    "status": false,
    "message": "Device is not a sensor or data unavailable"
  }
  ```
  *(Status: 400 Bad Request or 422 Unprocessable Entity)*

### 3. Security: Unauthorized
- **URL**: `http://localhost:8080/api/tuya/devices/:id/sensor`
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
