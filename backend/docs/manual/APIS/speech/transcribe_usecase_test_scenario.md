# ENDPOINT: POST /api/speech/transcribe

## Description
Starts transcription of an audio file. This is the standard entry point for short audio transcription. It follows a fallback logic:
1. Attempt **PPU (Outsystems Proxy)** transcription first.
2. If PPU fails or is unavailable, fall back to **Local Whisper** transcription.
The processing is **asynchronous**.

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
  "message": "Transcription started",
  "data": {
    "task_id": "abc123-def456-ghi789",
    "task_status": {
      "status": "processing",
      "expires_at": "2026-02-10T10:00:00Z",
      "expires_in_seconds": 3600,
      "result": null
    }
  }
}
```
  *(Status: 202 Accepted)*
- **Side Effects**: 
  - Task entry created in cache storage.
  - Background processing started (PPU attempted first).

### 2. Validation: Missing Audio File
- **Method**: `POST`
- **Request**: No file uploaded.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Audio file is required"
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
  "message": "Unsupported media type. Supported formats: .mp3, .wav, .m4a, .aac, .ogg, .flac"
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

### 6. Error: Internal Server Error
- **Pre-conditions**: Cache service or transcription engine is down.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Failed to start transcription"
}
```
  *(Status: 500 Internal Server Error)*
