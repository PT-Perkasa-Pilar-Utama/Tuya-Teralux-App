# ENDPOINT: GET /api/teralux/:id/scenes

## Description
Retrieve all scenes configured for a specific Teralux device. Each scene includes its full list of actions.

## Authentication
- **Type**: BearerAuth
- **Header**: `Authorization: Bearer <token>`

## Parameters
- `id` (path, required): Teralux device UUID

## Test Scenarios

### 1. Success - Get All Scenes with Actions
- **URL**: `http://localhost:8081/api/teralux/<id>/scenes`
- **Method**: `GET`
- **Expected Response** *(Status: 200 OK)*:
```json
{
  "status": true,
  "message": "Scenes retrieved successfully",
  "data": [
    {
      "id": "776cfc67-ab77-4bfa-a60e-8bed7611f921",
      "teralux_id": "1",
      "name": "Morning Scene",
      "actions": [
        { "device_id": "abc123", "code": "switch_led", "value": true },
        { "topic": "home/ac", "value": "on" }
      ]
    }
  ]
}
```

### 2. No Scenes
- **Expected Response** *(Status: 200 OK)*:
```json
{ "status": true, "message": "Scenes retrieved successfully", "data": [] }
```

### 3. Security: Unauthorized
- **Expected Response** *(Status: 401 Unauthorized)*:
```json
{ "status": false, "message": "Unauthorized" }
```
