# ENDPOINT: GET /api/teralux/:id/scenes

## Description
Retrieve all Scenes configured for a specific Teralux device.

## Test Scenarios

### 1. Get All Scenes (Success)
- **URL**: `http://localhost:8080/api/teralux/:id/scenes`
- **Method**: `GET`
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
  "message": "Scenes retrieved successfully",
  "data": [
    {
      "id": "uuid-1",
      "name": "Movie Night"
    },
    {
      "id": "uuid-2",
      "name": "Romantic Evening"
    },
    {
      "id": "uuid-3",
      "name": "Morning Wakeup"
    }
  ]
}
```
  *(Status: 200 OK)*

### 2. Application Security: Unauthorized
- **Headers**: Missing or invalid token.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Unauthorized"
}
```
  *(Status: 401 Unauthorized)*
