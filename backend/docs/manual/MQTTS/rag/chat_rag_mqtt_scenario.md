# MQTT: AI Assistant Chat Scenario

## 1. Overview
The AI assistant chat can be triggered via MQTT on the `users/teralux/chat` topic. Responses are published back to `users/teralux/chat/answer`.

## 2. MQTT Topic
- **Subscription Topic**: `users/teralux/chat`
- **Response Topic**: `users/teralux/chat/answer`
- **QoS**: 0

## 3. Test Cases

### 3.1 General Conversation (CHAT)
**Payload Format**: JSON (Strictly follows REST API)

**JSON Payload**:
```json
{
    "prompt": "Halo, siapa kamu?",
    "teralux_id": "tx-1",
    "language": "id"
}
```

**Expected Response (on `users/teralux/chat/answer`)**:
```json
{
  "status": true,
  "message": "Chat processed successfully",
  "data": {
    "response": "Halo! Saya adalah Sensio AI Assistant...",
    "is_control": false
  }
}
```

### 3.2 Device Control Request (CONTROL)
**Payload**:
```json
{
    "prompt": "Nyalakan AC",
    "teralux_id": "tx-1",
    "language": "id"
  }
```

**Expected Response (on `users/teralux/chat/answer`)**:
```json
{
  "status": true,
  "message": "Chat processed successfully",
  "data": {
    "response": "AC telah dinyalakan",
    "is_control": true,
    "redirect": {
      "endpoint": "/api/rag/control",
      "method": "POST",
      "body": {
        "prompt": "Nyalakan AC",
        "teralux_id": "tx-1"
      }
    }
  }
}
```

## 4. Verification Steps
1. Use an MQTT client (e.g., MQTT Explorer or `mosquitto_sub`) to subscribe to `users/teralux/chat/answer`.
2. Publish the CHAT payload to `users/teralux/chat`.
3. Verify the response message matches the expected format.
