# ENDPOINT: POST /api/speech/models/openai

## Description
Submits an audio file for transcription via the **OpenAI Whisper** model. This endpoint uses the OpenAI API and is processed **asynchronously**.

## Authentication
- **Type**: BearerAuth
- **Header**: `Authorization: Bearer <token>`

## Request Body
- **Content-Type**: `multipart/form-data`
- **Parameters**:
  - `audio` (file, required): Audio file. Supported formats: `.mp3`, `.wav`, `.m4a`, `.aac`, `.ogg`, `.flac`.
  - `language` (string, optional): Language code (e.g., `id`, `en`).

## Test Scenarios

### 1. Transcribe via OpenAI (Success)
- **Method**: `POST`
- **Headers**:
```json
{
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: Audio file is valid and OpenAI API is accessible.
- **Request**: Upload `audio.mp3`.
- **Expected Response**:
```json
{
  "status": true,
  "message": "OpenAI transcription task submitted",
  "data": {
    "task_id": "openai-abc123-xyz789",
    "task_status": "pending",
    "recording_id": "uuid-v4"
  }
}
```
  *(Status: 202 Accepted)*
- **Side Effects**: 
  - Task entry created in cache storage.
  - Background processing started via OpenAI.

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
- **Request**: Upload `image.png`.
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
- **Expected Behavior**: File accepted at API layer, but OpenAI API will return an error during background processing.
- **Expected Status**: Task status becomes `failed`.

### 8. Error: Internal Server Error
- **Pre-conditions**: File system error or database failure.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Internal Server Error"
}
```
*(Status: 500 Internal Server Error)*
