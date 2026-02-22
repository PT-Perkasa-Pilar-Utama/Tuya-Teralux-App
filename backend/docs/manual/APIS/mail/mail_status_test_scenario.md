# ENDPOINT: GET /api/email/status/:task_id

## Description
Retrieves the current status and result of an asynchronous email task. Poll this endpoint after receiving a `task_id` from `POST /api/email/send` or `POST /api/email/send/mac/:mac_address`.

## Authentication
- **Type**: BearerAuth
- **Header**: `Authorization: Bearer <token>`

## Parameters
- **Path Parameter**:
  - `task_id` (string): The task ID returned by the send endpoint.

## Status Values
| Status | Meaning |
|--------|---------|
| `pending` | Task queued, not yet started |
| `sending` | Email is being sent |
| `completed` | Email sent successfully |
| `failed` | An error occurred |

## Test Scenarios

### 1. Success - Task Completed
- **URL**: `http://localhost:8080/api/email/status/mail-abc-123xyz`
- **Method**: `GET`
- **Headers**: `Authorization: Bearer <valid_token>`
- **Expected Response** *(Status: 200 OK)*:
```json
{
  "status": true,
  "message": "Task status retrieved successfully",
  "data": {
    "status": "completed",
    "result": "Email sent to user@example.com",
    "started_at": "2026-02-22T07:00:00Z",
    "duration_seconds": 1.23,
    "expires_at": "2026-02-22T08:00:00Z",
    "expires_in_seconds": 3540
  }
}
```

### 2. Task Still Pending
- **Expected Response** *(Status: 200 OK)*:
```json
{
  "status": true,
  "message": "Task status retrieved successfully",
  "data": {
    "status": "pending"
  }
}
```

### 3. Task Failed
- **Expected Response** *(Status: 200 OK)*:
```json
{
  "status": false,
  "message": "Task failed",
  "data": {
    "status": "failed",
    "error": "smtp: 535 5.7.8 Authentication credentials invalid"
  }
}
```

### 4. Task Not Found
- **URL**: `http://localhost:8080/api/email/status/invalid-id`
- **Expected Response** *(Status: 404 Not Found)*:
```json
{
  "status": false,
  "message": "Task not found"
}
```

### 5. Security: Unauthorized
- **Headers**: No Bearer token.
- **Expected Response** *(Status: 401 Unauthorized)*:
```json
{
  "status": false,
  "message": "Unauthorized"
}
```

## Polling Recommendation
- Poll every **2 seconds** with a maximum of **30 attempts** (60 seconds total).
- Stop polling when status is `completed` or `failed`.
