# ENDPOINT: POST /api/rag/models/groq

## Description
Sends a raw prompt directly to the Groq LLM model configured in `GROQ_API_KEY` and `GROQ_MODEL`. This endpoint strictly returns the LLM response without going through the RAG Orchestrator or any RAG Skills.

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
    "trigger": "/api/rag/models/groq",
    "started_at": "2026-02-22T00:05:40+07:00",
    "duration_seconds": 0.441,
    "result": "Grateful to help! What do you need today?"
  }
}
```

### 2. Failure Response (429 Too Many Requests)
```json
{
  "status": false,
  "message": "Query execution failed",
  "data": {
    "status": "failed",
    "trigger": "/api/rag/models/groq",
    "started_at": "2026-02-22T00:05:40+07:00",
    "duration_seconds": 0.1,
    "error": "groq api returned status 429: Rate limit exceeded"
  }
}
```
