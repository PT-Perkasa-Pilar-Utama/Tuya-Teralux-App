# ENDPOINT: POST /api/speech/transcribe/long

## Description
Starts transcription of a long audio file using **local Whisper**. This endpoint is optimized for long duration recordings. It provides **background execution** and does **not** perform translation or RAG processing.

## Authentication
- **Type**: BearerAuth
- **Header**: `Authorization: Bearer <token>`

## Request Body
- **Content-Type**: `multipart/form-data`
- **Parameters**:
  - `audio` (file, required): Audio file. Supported formats: `.mp3`, `.wav`, `.m4a`, `.aac`, `.ogg`, `.flac`.
  - `language` (string, required): Language code (e.g., `id`, `en`, `es`, `fr`).

## Test Scenarios

### 1. Transcribe Long Audio (Success)
- **Method**: `POST`
- **Headers**:
```json
{
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: Valid long audio file, valid language code.
- **Request**: Upload `long_recording.wav` and `language="id"`.
- **Expected Response**:
```json
{
  "status": true,
  "message": "Long transcription started",
  "data": {
    "task_id": "long-xyz789-abc123",
    "task_status": {
      "status": "processing",
      "expires_at": "2026-02-10T12:00:00Z",
      "expires_in_seconds": 7200,
      "result": null
    }
  }
}
```
  *(Status: 202 Accepted)*
- **Side Effects**: Background transcription task started using local Whisper.

### 2. Validation: Missing Language
- **Method**: `POST`
- **Request**: Audio file uploaded but `language` parameter is missing.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Language parameter is required"
}
```
  *(Status: 400 Bad Request)*

### 3. Validation: Invalid Language Code
- **Method**: `POST`
- **Request**: `language="xyz"`.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Invalid language code"
}
```
  *(Status: 400 Bad Request)*

### 4. Security: Unauthorized
- **Headers**: No Authorization header.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Unauthorized"
}
```
  *(Status: 401 Unauthorized)*

### 5. Error: Whisper Engine Error
- **Pre-conditions**: Local Whisper service fails to initialize or process the file.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Transcription engine failure"
}
```
  *(Status: 500 Internal Server Error)*
