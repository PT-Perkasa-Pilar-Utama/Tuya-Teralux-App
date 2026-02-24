# ENDPOINT: POST /api/rag/models/openai

## Description
Sends a raw prompt directly to the OpenAI LLM model configured in `OPENAI_API_KEY` and `OPENAI_MODEL`. This endpoint strictly returns the LLM response without going through the RAG Orchestrator or any RAG Skills.

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
    "trigger": "/api/rag/models/openai",
    "started_at": "2026-02-22T00:05:40+07:00",
    "duration_seconds": 1.551,
    "result": "I am just a computer program, but I'm functioning perfectly."
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
    "trigger": "/api/rag/models/openai",
    "started_at": "2026-02-22T00:05:40+07:00",
    "duration_seconds": 0.5,
    "error": "openai api returned status 401: unauthorized"
  }
}
```
