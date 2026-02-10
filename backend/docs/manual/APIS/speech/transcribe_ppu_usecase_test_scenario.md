# ENDPOINT: POST /api/speech/transcribe/ppu

## Description
Submits an audio file for transcription via the **Outsystems Proxy (PPU)**. This endpoint explicitly tries the proxy service and is processed **asynchronously**.

## Authentication
- **Type**: BearerAuth
- **Header**: `Authorization: Bearer <token>`

## Request Body
- **Content-Type**: `multipart/form-data`
- **Parameters**:
  - `audio` (file, required): Audio file. Supported formats: `.mp3`, `.wav`, `.m4a`, `.aac`, `.ogg`, `.flac`.

## Test Scenarios

### 1. Transcribe via PPU (Success)
- **Method**: `POST`
- **Headers**:
```json
{
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: Audio file is valid and PPU service is reachable.
- **Request**: Upload `audio.m4a`.
- **Expected Response**:
```json
{
  "status": true,
  "message": "PPU transcription started",
  "data": {
    "task_id": "ppu-abc123-xyz789",
    "task_status": {
      "status": "processing",
      "expires_at": "2026-02-10T10:30:00Z",
      "expires_in_seconds": 3600,
      "result": null
    }
  }
}
```
  *(Status: 202 Accepted)*
- **Side Effects**: File is queued and forwarded to the Outsystems PPU service.

### 2. Validation: Unsupported Media Type
- **Method**: `POST`
- **Request**: Upload `unsupported.rar`.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Unsupported media type. Supported formats: .mp3, .wav, .m4a, .aac, .ogg, .flac"
}
```
  *(Status: 415 Unsupported Media Type)*

### 3. Error: PPU Proxy Unavailable
- **Pre-conditions**: The Outsystems Proxy service is down or returning errors.
- **Expected Response**:
```json
{
  "status": false,
  "message": "PPU service unavailable"
}
```
  *(Status: 500 Internal Server Error)*

### 4. Security: Unauthorized
- **Headers**: Invalid or missing token.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Unauthorized"
}
```
  *(Status: 401 Unauthorized)*
