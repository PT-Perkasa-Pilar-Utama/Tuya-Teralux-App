# ENDPOINT: GET /api/tuya/devices/{id}/sensor

## Description
Retrieves formatted sensor data (e.g., Temperature, Humidity) for a specific device. This endpoint filters and transforms raw Tuya status codes into a standard sensor format.

## Authentication
- **Type**: BearerAuth
- **Header**: `Authorization: Bearer <token>`

## Path Parameters
- `id` (string, required) - Device ID

## Test Scenarios

### 1. Get Sensor Data (Success)
- **Method**: `GET`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: Device is a sensor (e.g., TH Sensor) and is online.
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
- **Side Effects**: None (read-only).

### 2. Validation: Not a Sensor
- **Method**: `GET`
- **Pre-conditions**: Device ID refers to a device category that does not provide sensor data (e.g., a simple switch).
- **Expected Response**:
```json
{
  "status": false,
  "message": "Device is not a sensor or data unavailable"
}
```
  *(Status: 400 Bad Request)*

### 3. Security: Unauthorized
- **Headers**: No Authorization header.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Unauthorized"
}
```
  *(Status: 401 Unauthorized)*

### 4. Error: Device Offline
- **Pre-conditions**: Device is offline in Tuya Cloud.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Device is offline"
}
```
  *(Status: 500 Internal Server Error)*
