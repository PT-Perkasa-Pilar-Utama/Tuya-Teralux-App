# AI Assistant Chat Test Scenario

## 1. Overview
The AI Assistant Chat endpoint (`/api/rag/chat`) serves as the primary entry point for user interactions. It classifies the user's input as either a general conversation (CHAT) or a device control command (CONTROL).

## 2. API Endpoint
- **URL**: `/api/rag/chat`
- **Method**: `POST`
- **Auth**: Required (Bearer Token)

## 3. Test Cases

### 3.1 General Conversation (CHAT)
**Request Body**:
```json
{
    "prompt": "Halo, siapa kamu?",
    "language": "id"
}
```

**Expected Response**:
```json
{
    "status": true,
    "message": "Chat processed successfully",
    "data": {
        "response": "Halo! Saya adalah Sensio AI Assistant, asisten rumah pintar Anda. Ada yang bisa saya bantu hari ini?",
        "is_control": false
    }
}
```

### 3.2 Device Control Request (CONTROL)
**Request Body**:
```json
{
    "prompt": "Nyalakan AC di kamar",
    "language": "id"
}
```

**Expected Response**:
```json
{
    "status": true,
    "message": "Chat processed successfully",
    "data": {
        "response": "Processing your command...",
        "is_control": true,
        "redirect": {
            "endpoint": "/api/rag/control",
            "method": "POST",
            "body": {
                "prompt": "Nyalakan AC di kamar"
            }
        }
    }
}
```

### 3.3 Validation: Missing Prompt
**Request Body**:
```json
{
    "prompt": "",
    "teralux_id": "tx-1"
}
```

**Expected Response**:
```json
{
    "status": false,
    "message": "Validation Error",
    "details": [
        { "field": "prompt", "message": "prompt is required" }
    ]
}
```
*(Status: 400 Bad Request)*

## 4. Verification Steps
1. Send the CHAT request using `curl` or Postman.
2. Verify that `is_control` is `false` and the `response` is conversational.
3. Send the CONTROL request.
4. Verify that `is_control` is `true` and the `redirect` object points to `/api/rag/control`.
