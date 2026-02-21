# ENDPOINT: GET /api/speech/transcribe/{transcribe_id}

## Description
Retrieves the current status and result of any transcription task. This is a **consolidated** endpoint that works for any task ID returned by standard or model-specific transcription requests.

## Authentication
- **Type**: BearerAuth
- **Header**: `Authorization: Bearer <token>`

## Path Parameters
- `transcribe_id` (string, required): The Task ID returned by any of the transcription initiation endpoints.

## Test Scenarios

### 1. Get Processing Status (Success)
- **Method**: `GET`
- **Pre-conditions**: Task ID is valid and the task is still being processed.
- **Expected Response**:
```json
{
  "status": true,
  "message": "Task status retrieved successfully",
  "data": {
    "status": "pending",
    "started_at": "2026-02-21T11:00:00Z",
    "expires_at": "2026-02-21T12:00:00Z",
    "expires_in_seconds": 3600
  }
}
```
  *(Status: 200 OK)*

### 2. Get Completed Result (Success)
- **Method**: `GET`
- **Pre-conditions**: Task is completed.
- **Expected Response**:
```json
{
  "status": true,
  "message": "Task status retrieved successfully",
  "data": {
    "status": "completed",
    "result": {
      "transcription": "halo apa kabar",
      "detected_language": "id"
    },
    "trigger": "/api/speech/models/gemini",
    "started_at": "2026-02-21T11:00:04Z",
    "duration_seconds": 1.2,
    "expires_at": "2026-02-21T12:00:04Z",
    "expires_in_seconds": 3600
  }
}
```
  *(Status: 200 OK)*

### 3. Get Failed Status (Failure)
- **Method**: `GET`
- **Pre-conditions**: Task failed during processing.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Task failed",
  "data": {
    "status": "failed",
    "error": "openai whisper returned status 401: Unauthorized",
    "trigger": "/api/speech/models/openai",
    "started_at": "2026-02-21T11:10:00Z",
    "duration_seconds": 0.3,
    "expires_at": "2026-02-21T12:10:00Z",
    "expires_in_seconds": 3600
  }
}
```
  *(Status: 401 Unauthorized)*

### 4. Validation: Task Not Found
- **Method**: `GET`
- **Pre-conditions**: Invalid or expired task ID.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Transcription task not found"
}
```
  *(Status: 404 Not Found)*

### 5. Security: Unauthorized
- **Headers**: No Authorization header.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Unauthorized"
}
```
  *(Status: 401 Unauthorized)*
