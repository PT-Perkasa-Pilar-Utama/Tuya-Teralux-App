# ENDPOINT: DELETE /api/teralux/:id/scenes/:scene_id

## Description
Delete a Scene configuration for a specific Teralux device.

## Test Scenarios

### 1. Delete Scene (Success)
- **URL**: `http://localhost:8080/api/teralux/:id/scenes/:scene_id`
- **Method**: `DELETE`
- **Path Parameters**:
    - `id`: The UUID of the scene to delete.
- **Headers**:
```json
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Expected Response**:
```json
{
  "status": true,
  "message": "Scene deleted successfully"
}
```
  *(Status: 200 OK)*

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
