# MQTT: Whisper Audio Transcription Scenario

## 1. Overview
Transcription can be triggered by sending raw audio bytes to the `users/terminal/whisper` topic.

## 2. MQTT Topic
- **Subscription Topic**: `users/terminal/whisper`
- **Response Topic**: `users/terminal/whisper/answer`
- **Payload Format**: JSON

**Payload**:
```json
{
    "audio": "BASE64_ENCODED_AUDIO_DATA",
    "terminal_id": "tx-1",
    "language": "id"
}
```

## 3. Expected Response (on `users/terminal/whisper/answer`)
```json
{
  "status": true,
  "message": "Transcription task submitted successfully",
  "data": {
    "task_id": "abc-123",
    "task_status": "pending"
  }
}
```

## 4. Behavior
- The backend receives the JSON payload and decodes the `audio` field.
- It saves the data to a file in `uploads/audio/mqtt/`.
- It triggers the internal transcription pipeline.
- **Note**: This process does not create a record in the `recordings` database table.

## 5. Verification Steps
1. Prepare a small audio file and convert it to Base64.
2. Publish the JSON payload to `users/terminal/whisper`.
3. Verify the response on `users/terminal/whisper/answer`.

