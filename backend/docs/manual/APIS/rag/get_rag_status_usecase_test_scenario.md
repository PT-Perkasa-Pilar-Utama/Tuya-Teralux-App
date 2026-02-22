# ENDPOINT: GET /api/rag/{task_id}

## Description
Retrieves the status and execution result of a RAG processing task.

## Authentication
- **Type**: BearerAuth
- **Header**: `Authorization: Bearer <token>`

## Path Parameters
- `task_id` (string, required): The Task ID returned by the POST /api/rag endpoint.

## Test Scenarios

### 1. Get Completed RAG Result (Success)
- **Method**: `GET`
- **Pre-conditions**: Task is completed successfully.
- **Expected Response**:
```json
{
  "status": true,
  "message": "Task status retrieved successfully",
  "data": {
    "status": "completed",
    "result": "The remote work policy allows up to 3 days per week...",
    "trigger": "/api/rag/summary",
    "started_at": "2026-02-10T11:00:00Z",
    "duration_seconds": 1.5,
    "expires_at": "2026-02-10T12:00:00Z",
    "expires_in_seconds": 3600
  }
}
```
*(Status: 200 OK)*

### 2. Get Failed Status (Failure)
- **Method**: `GET`
- **Pre-conditions**: The external service returned an error during execution.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Task failed",
  "data": {
    "status": "failed",
    "error": "Gemini API returned status 503: Service Unavailable",
    "trigger": "/api/rag/summary",
    "started_at": "2026-02-10T11:00:00Z",
    "duration_seconds": 0.5,
    "expires_at": "2026-02-10T12:00:00Z",
    "expires_in_seconds": 3600
  }
}
```
*(Status: 200 OK)*

### 3. Validation: Task Not Found
- **Method**: `GET`
- **Expected Response**:
```json
{
  "status": false,
  "message": "RAG task not found"
}
```
*(Status: 404 Not Found)*

### 4. Security: Unauthorized
- **Headers**: Invalid token.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Unauthorized"
}
```
*(Status: 401 Unauthorized)*
