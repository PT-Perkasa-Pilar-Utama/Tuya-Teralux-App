# ENDPOINT: POST /api/speech/transcribe

## Description
Starts transcription of an audio file with automatic provider fallback. This endpoint automatically refines the output (KBBI for Indonesian, Grammar Fix for English).

### Processing Flow
1. **Orion First**: System attempts to send audio to the Outsystems Orion proxy with health check.
2. **Local Fallback**: If Orion is unavailable or fails, finally falls back to **Local Whisper.cpp** engine.

Processing is **asynchronous** with automatic failover between providers.

## Authentication
- **Type**: BearerAuth
- **Header**: `Authorization: Bearer <token>`

## Request Body
- **Content-Type**: `multipart/form-data`
- **Parameters**:
  - `audio` (file, required): Audio file. Supported formats: `.mp3`, `.wav`, `.m4a`, `.aac`, `.ogg`, `.flac`.

## Test Scenarios

### 1. Transcribe Audio File (Success)
- **Method**: `POST`
- **Headers**:
```json
{
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: Valid audio file, valid Bearer token.
- **Request**: Upload `audio.mp3`.
- **Expected Response**:
```json
{
  "status": true,
  "message": "Transcription task submitted successfully",
  "data": {
    "task_id": "abc123-def456-ghi789",
    "task_status": "pending"
  }
}
```
  *(Status: 202 Accepted)*
- **Side Effects**: 
  - Task entry created in cache storage.
  - Background processing started.

### 2. Validation: Missing Audio File
- **Method**: `POST`
- **Request**: No file uploaded.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Validation Error",
  "details": [
    { "field": "audio", "message": "audio file is required" }
  ]
}
```
  *(Status: 400 Bad Request)*

### 3. Validation: Unsupported File Type
- **Method**: `POST`
- **Request**: Upload `image.png`.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Validation Error",
  "details": [
    { "field": "audio", "message": "Unsupported media type. Supported formats: .mp3, .wav, .m4a, .aac, .ogg, .flac" }
  ]
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
  "message": "File size exceeds maximum limit"
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
- **Expected Behavior**: The transcription process completes successfully.
- **Expected Result**: `transcription: ""` and `refined_text: ""` because no speech was detected.

### 7. Validation: Wrong Extension / Corrupt Header
- **Request**: Upload a `.txt` file renamed to `.mp3`.
- **Expected Behavior**: The file is accepted at the API layer (due to extension check).
- **Processing Outcome**: The background transcription engine (Whisper) will fail to decode the audio.
- **Expected Status**: Task status becomes `failed` after processing.

### 8. Error: Internal Server Error
- **Pre-conditions**: Both Orion and Local Whisper engines are failing or system resources are exhausted.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Failed to start transcription"
}
```
*(Status: 500 Internal Server Error)*
