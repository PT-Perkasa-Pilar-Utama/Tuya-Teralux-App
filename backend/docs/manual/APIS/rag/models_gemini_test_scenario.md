# ENDPOINT: POST /api/rag/models/gemini

## Description
Sends a raw prompt directly to the Google Gemini LLM model configured in `GEMINI_API_KEY` and `GEMINI_MODEL_LOW`/`GEMINI_MODEL_HIGH`. This endpoint strictly returns the LLM response without going through the RAG Orchestrator or any RAG Skills.

## Authentication
- **Type**: BearerAuth
- **Header**: `Authorization: Bearer <token>`

## Request
- **Method**: `POST`
- **Content-Type**: `application/json`

### Body Example
```json
{
  "prompt": "Hello! How are you?"
}
```

## Responses

### 1. Success Response (200 OK)
```json
{
  "status": true,
  "message": "Query executed successfully",
  "data": {
    "status": "completed",
    "trigger": "/api/rag/models/gemini",
    "started_at": "2026-02-22T00:05:40+07:00",
    "duration_seconds": 2.123,
    "result": "I'm doing well, thank you for asking! How can I help you today?"
  }
}
```

### 2. Failure Response (500 Internal Server Error)
```json
{
  "status": false,
  "message": "Query execution failed",
  "data": {
    "status": "failed",
    "trigger": "/api/rag/models/gemini",
    "started_at": "2026-02-22T00:05:40+07:00",
    "duration_seconds": 1.25,
    "error": "failed to call gemini api: timeout"
  }
}
```
