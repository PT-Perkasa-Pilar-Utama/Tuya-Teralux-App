# ENDPOINT: POST /api/rag/models/llama/cpp

## Description
Sends a raw prompt directly to the Local Llama.cpp backend configured via `LLAMA_CPP_BASE_URL`. This endpoint strictly returns the LLM response without going through the RAG Orchestrator or any RAG Skills.

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
    "trigger": "/api/rag/models/llama/cpp",
    "started_at": "2026-02-22T00:05:40+07:00",
    "duration_seconds": 3.441,
    "result": "I am a local AI assistant. I am functioning normally."
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
    "trigger": "/api/rag/models/llama/cpp",
    "started_at": "2026-02-22T00:05:40+07:00",
    "duration_seconds": 12.1,
    "error": "failed to call local llama api: connection refused"
  }
}
```
