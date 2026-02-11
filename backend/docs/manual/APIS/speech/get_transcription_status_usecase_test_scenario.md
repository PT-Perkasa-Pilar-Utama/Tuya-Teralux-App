# ENDPOINT: GET /api/speech/transcribe/{transcribe_id}

## Description
Retrieves the current status and result of any transcription task. This is a **consolidated** endpoint that works for standard, long, and PPU transcription tasks.

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
  "message": "Transcription status retrieved",
  "data": {
    "task_id": "short-abc123",
    "task_status": {
      "status": "processing",
      "expires_at": "2026-02-10T11:00:00Z",
      "expires_in_seconds": 3500,
      "result": null
    }
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
  "message": "Transcription status retrieved",
  "data": {
    "task_id": "short-abc123",
    "task_status": {
      "status": "completed",
      "expires_at": "2026-02-10T11:00:00Z",
      "expires_in_seconds": 3200,
      "result": {
        "transcription": "halo apa kabar",
        "refined_text": "Halo, apa kabar?",
        "detected_language": "id"
      }
    }
  }
}
```
  *(Status: 200 OK)*

### 3. Get Failed Status (Success)
- **Method**: `GET`
- **Pre-conditions**: Task failed during processing.
- **Expected Response**:
```json
{
  "status": true,
  "message": "Transcription status retrieved",
  "data": {
    "task_id": "short-abc123",
    "task_status": {
      "status": "failed",
      "expires_at": "2026-02-10T11:00:00Z",
      "expires_in_seconds": 3000,
      "result": null
    }
  }
}
```
  *(Status: 200 OK)*

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
