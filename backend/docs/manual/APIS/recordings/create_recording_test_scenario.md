# ENDPOINT: POST /api/recordings

## Description
Uploads an audio recording to the system. The recording is saved to disk and metadata is stored in the database.

## Authentication
- **Type**: BearerAuth (Optional depending on global middleware, but recommended)
- **Header**: `Authorization: Bearer <token>`

## Form Parameters
- `audio` (file, required): The audio file to upload.
- `mac_address` (string, optional): The device's MAC address to associate with the recording. If provided, it will trigger an external API update to the BIG system.

## Test Scenarios

### 1. Create Recording (Success)
- **Method**: `POST`
- **Form Data**:
    - `audio`: (selection of a .wav or .mp3 file)
    - `mac_address`: `AA:BB:CC:DD:EE:FF`
- **Expected Response**:
```json
{
  "status": true,
  "message": "Recording saved successfully",
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "filename": "123e4567-e89b-12d3-a456-426614174000.wav",
    "original_name": "meeting_recording.wav",
    "audio_url": "http://localhost:8081/uploads/audio/123e4567-e89b-12d3-a456-426614174000.wav",
    "mac_address": "AA:BB:CC:DD:EE:FF",
    "created_at": "2026-02-11T10:00:00Z"
  }
}
```
  *(Status: 201 Created)*

### 2. Validation: Missing Audio File
- **Method**: `POST`
- **Form Data**:
    - `mac_address`: `AA:BB:CC:DD:EE:FF`
- **Expected Response**:
```json
{
  "status": false,
  "message": "http: no such file"
}
```
  *(Status: 400 Bad Request)*
