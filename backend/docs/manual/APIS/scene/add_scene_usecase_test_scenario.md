# ENDPOINT: POST /api/scenes

## Description
Create a new Scene. A scene is a preset configuration of multiple devices that can be activated simultaneously.

## Test Scenarios

### 1. Create Scene (Success)
- **URL**: `http://localhost:8080/api/scenes`
- **Method**: `POST`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Request Body**:
```json
{
  "name": "Movie Night",
  "actions": [
    {
      "device_id": "device-001",
      "code": "switch_1",
      "value": false
    },
    {
      "device_id": "ir-device-002",
      "code": "TV_Power",
      "remote_id": "remote-xyz-123",
      "value": "ON"
    },
    {
      "topic": "home/livingroom/notify",
      "value": "Movie Mode Activated"
    }
  ]
}
```
- **Expected Response**:
```json
{
  "status": true,
  "message": "Scene created successfully",
  "data": {
    "scene_id": "<uuid>"
  }
}
```
  *(Status: 201 Created)*

### 2. Validation: Missing Name
- **Method**: `POST`
- **Request Body**:
```json
{
  "name": ""
}
```
- **Expected Response**:
```json
{
  "status": false,
  "message": "Validation Error",
  "details": [
    { "field": "name", "message": "name is required" }
  ]
}
```
  *(Status: 422 Unprocessable Entity)*

### 3. Application Security: Unauthorized
- **Headers**: Missing or invalid token.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Unauthorized"
}
```
  *(Status: 401 Unauthorized)*
