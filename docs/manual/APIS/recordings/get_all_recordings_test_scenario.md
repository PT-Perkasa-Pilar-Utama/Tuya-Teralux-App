# Get All Recordings - Test Scenario

## Objective
Verify that the `GET /api/recordings` endpoint correctly retrieves a paginated list of recordings sorted by `created_at` descending.

## Pre-requisites
- Backend server running.
- At least one recording uploaded (via `POST /api/speech/transcribe`).

## Test Steps

1. **Upload Audio**:
   Use `curl` or Postman to upload an audio file.
   ```bash
   curl -X POST http://localhost:8080/api/speech/transcribe \
     -H "Authorization: Bearer <YOUR_TOKEN>" \
     -F "file=@/path/to/test_audio.wav"
   ```

2. **Get All Recordings**:
   Retrieve the list of recordings.
   ```bash
   curl -X GET "http://localhost:8080/api/recordings?page=1&limit=5" \
     -H "Authorization: Bearer <YOUR_TOKEN>"
   ```

3. **Verify Response**:
   The response should be JSON with `recordings`, `total`, `page`, and `limit`.
   ```json
   {
       "recordings": [
           {
               "id": "uuid-string",
               "filename": "uuid-string.wav",
               "original_name": "test_audio.wav",
               "audio_url": "/api/static/audio/uuid-string.wav",
               "created_at": "2023-11-20T10:00:00Z"
           }
       ],
       "total": 1,
       "page": 1,
       "limit": 5
   }
   ```

4. **Verify Audio Access**:
   Copy the `audio_url` from the response (e.g., `/api/static/audio/uuid-string.wav`) and open it in a browser or fetch it via curl prefixing the base URL.
   ```bash
   curl -I http://localhost:8080/api/static/audio/uuid-string.wav
   ```
   Should return `HTTP/1.1 200 OK`.

## Edge Cases
- **No Recordings**: Should return empty list `[]` and `total: 0`.
- **Invalid Page/Limit**: specific behavior (default to 1/10).
