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
  "message": "RAG status retrieved",
  "data": {
    "task_id": "rag-xyz789",
    "task_status": {
      "status": "completed",
      "expires_at": "2026-02-10T11:00:00Z",
      "expires_in_seconds": 2400,
      "result": "The remote work policy allows up to 3 days per week..."
    }
  }
}
```
*(Status: 200 OK)*

### 2. Get Failed Status (Success)
- **Method**: `GET`
- **Pre-conditions**: The external service returned an error during execution.
- **Expected Response**:
```json
{
  "status": true,
  "message": "RAG status retrieved",
  "data": {
    "task_id": "rag-xyz789",
    "task_status": {
      "status": "failed",
      "expires_at": "2026-02-10T11:00:00Z",
      "expires_in_seconds": 1800,
      "result": "Error: Upstream service timeout (Code 504)"
    }
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
