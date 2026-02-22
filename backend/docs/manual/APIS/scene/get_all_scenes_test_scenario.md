# ENDPOINT: GET /api/scenes

## Description
Retrieve all scenes across all Teralux devices. Each element in `data` wraps a `teralux` object that contains `teralux_id` and its `scenes`.

## Authentication
- **Type**: BearerAuth
- **Header**: `Authorization: Bearer <token>`

## Test Scenarios

### 1. Success
- **URL**: `http://localhost:8081/api/scenes`
- **Method**: `GET`
- **Expected Response** *(Status: 200 OK)*:
```json
{
  "status": true,
  "message": "Scenes retrieved successfully",
  "data": [
    {
      "teralux": {
        "teralux_id": "1",
        "scenes": [
          {
            "id": "776cfc67-ab77-4bfa-a60e-8bed7611f921",
            "name": "Morning Scene",
            "actions": [
              { "device_id": "abc123", "code": "switch_led", "value": true },
              { "topic": "home/ac", "value": "on" }
            ]
          }
        ]
      }
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
