# ENDPOINT: GET /api/teralux/:id/scenes/:scene_id/control

## Description
Trigger/Apply a Scene for a specific Teralux device.

## Test Scenarios

### 1. Control Scene (Success)
- **URL**: `http://localhost:8080/api/teralux/:id/scenes/:scene_id/control`
- **Method**: `GET`
- **Path Parameters**:
    - `teralux_id`: The UUID of the Teralux device.
    - `id`: The UUID of the scene to control.
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
  "status": true,
  "message": "Scene applied successfully"
}
```
  *(Status: 200 OK)*
- **Side Effects**: All devices associated with the scene update their state.

### 2. Error: Scene Not Found
- **Path Parameters**: `id` = `non-existent-id`
- **Expected Response**:
```json
{
  "status": false,
  "message": "Scene not found"
}
```
  *(Status: 404 Not Found)*

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
