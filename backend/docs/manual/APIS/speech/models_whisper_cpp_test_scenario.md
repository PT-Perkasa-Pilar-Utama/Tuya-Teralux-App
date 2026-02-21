# ENDPOINT: POST /api/speech/models/whisper/cpp

## Description
Starts transcription of an audio file using **local Whisper.cpp** model. This endpoint provides **background execution** and is processed **asynchronously**.

## Authentication
- **Type**: BearerAuth
- **Header**: `Authorization: Bearer <token>`

## Request Body
- **Content-Type**: `multipart/form-data`
- **Parameters**:
  - `audio` (file, required): Audio file. Supported formats: `.mp3`, `.wav`, `.m4a`, `.aac`, `.ogg`, `.flac`.
  - `language` (string, optional): Language code (e.g., `id`, `en`).

## Test Scenarios

### 1. Transcribe via Whisper.cpp (Success)
- **Method**: `POST`
- **Headers**:
```json
{
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: Valid audio file.
- **Request**: Upload `recording.wav` and `language="id"`.
- **Expected Response**:
```json
{
  "status": true,
  "message": "Whisper.cpp transcription task submitted",
  "data": {
    "task_id": "whisper-xyz789-abc123",
    "task_status": "pending",
    "recording_id": "uuid-v4"
  }
}
```
  *(Status: 202 Accepted)*
- **Side Effects**: 
  - Task entry created in cache storage.
  - Background transcription task started using local Whisper.cpp.

### 2. Validation: Missing Audio File
- **Method**: `POST`
- **Request**: No file uploaded.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Validation Error",
  "details": [
    { "field": "audio", "message": "Audio file is required: http: no such file" }
  ]
}
```
  *(Status: 400 Bad Request)*

### 3. Validation: Unsupported File Type
- **Method**: `POST`
- **Request**: Upload `video.mp4`.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Unsupported Media Type"
}
```
  *(Status: 415 Unsupported Media Type)*

### 4. Validation: File Too Large
- **Method**: `POST`
- **Pre-conditions**: Upload file exceeding the configured maximum size.
- **Expected Response**:
```json
{
  "status": false,
  "message": "File too large"
}
```
  *(Status: 413 Request Entity Too Large)*

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

### 6. Scenario: Silent Audio
- **Request**: Upload 5 seconds of absolute silence.
- **Expected Behavior**: Transcription task completes successfully but result text will be empty.

### 7. Validation: Wrong Extension / Corrupt Header
- **Request**: Upload a `.txt` file renamed to `.mp3`.
- **Expected Behavior**: File accepted at API layer, but local Whisper engine will fail to process it in background.
- **Expected Status**: Task status becomes `failed`.

### 8. Error: Internal Server Error
- **Pre-conditions**: Local Whisper service fails to initialize or process the file saving.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Internal Server Error"
}
```
*(Status: 500 Internal Server Error)*
