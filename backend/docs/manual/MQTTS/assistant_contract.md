# MQTT Assistant E2E Contract

## Overview

This contract defines the canonical message structures for end-to-end communication between the Teralux Client App and the Backend via MQTT for Voice/Chat Assistant flows.

The standard response pattern requires all `answer` messages to include the correlation ID (`request_id`) and processing metadata (`source`), so the client can reliably map asynchronous responses to their origin requests.

### Client Topics
1. **Chat Submission**: `users/{TerminalID}/{ENV}/chat`
2. **Whisper Submission**: `users/{TerminalID}/{ENV}/whisper`

### Server Topics
1. **Chat Response**: `users/{TerminalID}/{ENV}/chat/answer`
2. **Whisper Response**: `users/{TerminalID}/{ENV}/whisper/answer`
3. *(Optional)* **Task Signal**: `users/{TerminalID}/{ENV}/task`

---

## 1. Chat Flow

### 1.1 Request (FE -> BE)
- **Topic**: `users/{TerminalID}/{env}/chat`
- **Payload**:
```json
{
  "uid": "123456",
  "terminal_id": "ABCDEF123456",
  "prompt": "Nyalakan lampu",
  "language": "id",
  "request_id": "req-uuid-1234"
}
```

### 1.2 Response (BE -> FE)
- **Topic**: `users/{TerminalID}/{env}/chat/answer`
- **Payload**:
```json
{
  "status": true,
  "message": "Chat processed successfully",
  "data": {
    "request_id": "req-uuid-1234",
    "response": "Lampu telah dinyalakan",
    "is_control": true,
    "http_status_code": 200,
    "source": "MQTT_SUBSCRIBER" 
  }
}
```
*Note: `source` can be `MQTT_SUBSCRIBER`, `IDEMPOTENCY_CACHED`, `IDEMPOTENCY_IN_PROGRESS`, etc.*

---

## 2. Whisper (Voice) Flow

### 2.1 Request (FE -> BE)
- **Topic**: `users/{TerminalID}/{env}/whisper`
- **Payload**:
```json
{
  "uid": "123456",
  "terminal_id": "ABCDEF123456",
  "audio": "base64_encoded_audio_bytes...",
  "language": "id",
  "diarize": false,
  "request_id": "req-uuid-5678"
}
```

### 2.2 Ack / In-Progress Response (BE -> FE)
*(Immediate ACK indicating voice is received and transcription logic started)*
- **Topic**: `users/{TerminalID}/{env}/whisper/answer`
- **Payload**:
```json
{
  "status": true,
  "message": "Transcription task submitted successfully (Ephemeral)",
  "data": {
    "request_id": "req-uuid-5678",
    "task_id": "task-uuid-8888",
    "task_status": "pending",
    "source": "MQTT_ACK"
  }
}
```

### 2.3 Final Transcription Response (BE -> FE)
*(Emitted when transcription finishes. Should echo `request_id` and contain `response` as the transcribed text.)*
- **Topic**: `users/{TerminalID}/{env}/whisper/answer`
- **Payload**:
```json
{
  "status": true,
  "message": "Transcription processed successfully",
  "data": {
    "request_id": "req-uuid-5678",
    "response": "Tolong matikan AC",
    "task_id": "task-uuid-8888",
    "task_status": "completed",
    "source": "WHISPER_WORKER"
  }
}
```

---

## 3. Client State Machine Rules

1. **Request Tracking**: FE must generate a unique `request_id` for every Chat or Whisper request and store it as the active `request_id`.
2. **Whisper Submission**:
   - Sends to `/whisper`.
   - Transitions to `Processing` state.
   - Ignores `whisper/answer` payloads where `data.task_status == "pending"`.
   - On final `whisper/answer` (contains `data.response` and `data.request_id` matches), FE does **NOT** resolve the conversation. It triggers a chained `Chat` flow automatically with the transcribed text using the SAME `request_id`.
3. **Chat Submission**:
   - Sends to `/chat`.
   - Transitions to `Processing` state.
   - On `chat/answer`, checks `data.request_id`. If it matches active `request_id`, resolves the conversation and shows `data.response`.
4. **Duplicate Dropping**: If BE replies with `source: MQTT_SYNC_DROP` or `IDEMPOTENCY_IN_PROGRESS`, FE must ignore it to avoid premature state transition. FE must only resolve on final payloads.
