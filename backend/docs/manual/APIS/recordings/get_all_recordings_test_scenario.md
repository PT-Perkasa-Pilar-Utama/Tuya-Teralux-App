# ENDPOINT: GET /api/recordings

## Description
Retrieves a paginated list of audio recordings stored in the system. Recordings include both the original uploads and processed UUID-named files.

## Authentication
- **Type**: BearerAuth (Optional depending on global middleware, but recommended)
- **Header**: `Authorization: Bearer <token>`

## Query Parameters
- `page` (int, optional): Page number (default: 1)
- `limit` (int, optional): Items per page (default: 10)

## Test Scenarios

### 1. Get Recordings (Success)
- **Method**: `GET`
- **Query**: `?page=1&limit=2`
- **Expected Response**:
```json
{
  "status": true,
  "message": "Recordings retrieved successfully",
  "data": {
    "recordings": [
      {
        "id": "abc-123",
        "filename": "abc-123.wav",
        "original_name": "meeting.mp3",
        "audio_url": "http://localhost:8081/uploads/recordings/abc-123.wav",
        "created_at": "2026-02-11T10:00:00Z"
      },
      {
        "id": "def-456",
        "filename": "def-456.wav",
        "original_name": "voice_note.m4a",
        "audio_url": "http://localhost:8081/uploads/recordings/def-456.wav",
        "created_at": "2026-02-11T11:00:00Z"
      }
    ],
    "total": 5,
    "page": 1,
    "limit": 2
  }
}
```
  *(Status: 200 OK)*

### 2. Validation: Empty Result
- **Query**: `?page=999`
- **Expected Response**:
```json
{
  "status": true,
  "data": { "recordings": [], "total": 5, "page": 999, "limit": 2 }
}
```
  *(Status: 200 OK)*

### 3. Validation: Invalid Parameters
- **Query**: `?page=-1&limit=abc`
- **Expected Behavior**: System should default to `page=1` and `limit=10`.
- **Expected Response**:
```json
{
  "status": true,
  "data": {
    "recordings": [...],
    "total": 5,
    "page": 1,
    "limit": 10
  }
}
```
  *(Status: 200 OK)*

### 4. Serving Static File
- **Method**: `GET`
- **URL**: `/uploads/recordings/abc-123.wav`
- **Expected Response**: Binary audio stream.
  *(Status: 200 OK)*
- **Header**: `Content-Type: audio/wav` (or appropriate format)
