# ENDPOINT: POST /api/rag/control

## Description
Submits text for **RAG Device Control** processing with automatic LLM provider fallback. This endpoint interprets natural language commands to control devices via Tuya or other integrations.

### Processing Flow
1. **Orion First**: System attempts to use Orion LLM API with health check.
2. **Gemini Fallback**: If Orion is unavailable or fails, automatically falls back to **Google Gemini API**.
3. **Ollama Fallback**: If Gemini is unavailable or fails, finally falls back to **Local Ollama** service.

The execution is **asynchronous** with automatic failover between LLM providers and returns a task ID.

## Authentication
- **Type**: BearerAuth
- **Header**: `Authorization: Bearer <token>`

## Request Body
- **Content-Type**: `application/json`
- **Required Fields**:
  - `text` (string): The query or text to process.

## Test Scenarios

### 1. Submit RAG Process (Success)
- **Method**: `POST`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: LLM engine/service is available.
- **Request Body**:
```json
{
  "text": "What is the policy for remote work?"
}
```
- **Expected Response**:
```json
{
  "status": true,
  "message": "RAG processing started",
  "data": {
    "task_id": "rag-xyz789",
    "task_status": {
      "status": "processing",
      "expires_at": "2026-02-10T11:00:00Z",
      "expires_in_seconds": 3600
    }
  }
}
```
  *(Status: 202 Accepted)*
- **Side Effects**: Task is initialized in cache and background processing starts.

### 2. Validation: Missing Text
- **Method**: `POST`
- **Request Body**: `{}`
- **Expected Response**:
```json
{
  "status": false,
  "message": "Text field is required"
}
```
  *(Status: 400 Bad Request)*

### 3. Security: Unauthorized
- **Headers**: No Bearer token provided.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Unauthorized"
}
```
  *(Status: 401 Unauthorized)*

### 4. Error: RAG Service Error
- **Pre-conditions**: External LLM service provider returns an error.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Failed to start RAG processing"
}
```
  *(Status: 500 Internal Server Error)*
