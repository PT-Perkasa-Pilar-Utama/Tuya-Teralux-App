# ENDPOINT: POST /api/rag/models/orion

## Description
Sends a raw prompt directly to the private Orion Platform LLM model configured in `ORION_API_KEY`, `ORION_BASE_URL`, and `ORION_MODEL`. This endpoint strictly returns the LLM response without going through the RAG Orchestrator or any RAG Skills.

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
    "trigger": "/api/rag/models/orion",
    "started_at": "2026-02-22T00:05:40+07:00",
    "duration_seconds": 2.51,
    "result": "Hello! As an AI residing on the Orion platform, I am fully operational."
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
    "trigger": "/api/rag/models/orion",
    "started_at": "2026-02-22T00:05:40+07:00",
    "duration_seconds": 0.3,
    "error": "orion api returned status 500: Server Error"
  }
}
```
