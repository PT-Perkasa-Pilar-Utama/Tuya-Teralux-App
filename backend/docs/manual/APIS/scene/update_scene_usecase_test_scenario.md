# ENDPOINT: PUT /api/teralux/:id/scenes/:scene_id

## Description
Update an existing Scene configuration for a specific Teralux device.

## Test Scenarios

### 1. Update Scene (Success)
- **URL**: `http://localhost:8080/api/teralux/:id/scenes/:scene_id`
- **Method**: `PUT`
- **Path Parameters**:
    - `id`: The UUID of the scene to update.
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
  "name": "Romantic Evening",
  "actions": [
    {
      "device_id": "device-001",
      "code": "bright_value",
      "value": 1000
    },
    {
      "device_id": "ir-device-002",
      "code": "AC_Temp",
      "remote_id": "remote-abc-456",
      "value": "24"
    },
    {
      "topic": "home/bedroom/alert",
      "value": "Romance Mode"
    }
  ]
}
```
- **Expected Response**:
```json
{
  "status": true,
  "message": "Scene updated successfully",
  "data": {
    "scene_id": "<uuid>"
  }
}
```
  *(Status: 200 OK)*

### 2. Validation: Empty Name
- **Request Body**: `{"name": ""}`
- **Expected Response**:
```json
{
  "status": false,
  "message": "Validation Error",
  "details": [
    { "field": "name", "message": "name cannot be empty" }
  ]
}
```
  *(Status: 422 Unprocessable Entity)*

### 3. Error: Scene Not Found
- **Path Parameters**: `id` = `non-existent-id`
- **Expected Response**:
```json
{
  "status": false,
  "message": "Scene not found"
}
```
  *(Status: 404 Not Found)*

### 4. Application Security: Unauthorized
- **Headers**: Missing or invalid token.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Unauthorized"
}
```
  *(Status: 401 Unauthorized)*
