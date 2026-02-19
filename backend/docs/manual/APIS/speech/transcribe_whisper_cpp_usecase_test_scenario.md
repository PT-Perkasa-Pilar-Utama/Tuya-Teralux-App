# ENDPOINT: POST /api/speech/transcribe/whisper/cpp

## Description
Starts transcription of an audio file using **local Whisper.cpp only** (no provider fallback). This endpoint is optimized for longer recordings and provides **background execution**. It automatically refines the output (KBBI for Indonesian, Grammar Fix for English) using LLM with automatic provider fallback (Orion → Gemini → Ollama).

## Authentication
- **Type**: BearerAuth
- **Header**: `Authorization: Bearer <token>`

## Request Body
- **Content-Type**: `multipart/form-data`
- **Parameters**:
  - `audio` (file, required): Audio file. Supported formats: `.mp3`, `.wav`, `.m4a`, `.aac`, `.ogg`, `.flac`.
  - `language` (string, required): Language code (e.g., `id`, `en`, `es`, `fr`).

## Test Scenarios

### 1. Transcribe Audio (Success)
- **Method**: `POST`
- **Headers**:
```json
{
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: Valid audio file, valid language code.
- **Request**: Upload `recording.wav` and `language="id"`.
- **Expected Response**:
```json
{
  "status": true,
  "message": "Transcription task submitted successfully",
  "data": {
    "task_id": "whisper-xyz789-abc123",
    "task_status": "pending"
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
  "message": "Validation Error",
  "details": [
    { "field": "language", "message": "language is required" }
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
  "message": "Unsupported file type"
}
```
  *(Status: 415 Unsupported Media Type)*

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

### 5. Error: Transcription Engine Error
- **Pre-conditions**: Local Whisper service fails to initialize or process the file.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Failed to start transcription task"
}
```
  *(Status: 500 Internal Server Error)*
