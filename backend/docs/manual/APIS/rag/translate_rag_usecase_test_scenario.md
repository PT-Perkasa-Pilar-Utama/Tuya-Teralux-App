# ENDPOINT: POST /api/rag/translate

## Description

Directly translates input text to the target language using LLM with automatic provider fallback (Orion → Gemini → Ollama). Best for short phrases or command normalization. For meetings, use the Summary endpoint instead.

## Authentication

- **Type**: BearerAuth
- **Header**: `Authorization: Bearer <token>`

## Request Body

- **Content-Type**: `application/json`
- **Headers**:
  - `Idempotency-Key` (string, optional): A unique key to prevent duplicate processing. Duplicate requests with the same key will return the existing task ID.
- **Required Fields**:
  - `text` (string): The text to translate.

## Test Scenarios

### 1. Translate Text (Success)

- **Method**: `POST`
- **Headers**:

```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>",
  "Idempotency-Key": "my-trans-key-456"
}
```

- **Request Body**:

```json
{
  "text": "nyalakan lampu ruang tamu"
}
```

- **Expected Response**:

```json
{
  "status": true,
  "message": "Translation task queued",
  "data": {
    "task_id": "18616e7f-d77e-4b9a-bc7a-8b2ec8cd341e"
  }
}
```

_(Status: 202 Accepted)_

### 2. Validation: Missing Text

- **Method**: `POST`
- **Request Body**: `{}`
- **Expected Response**:

```json
{
  "status": false,
  "message": "Validation Error",
  "details": [{ "field": "text", "message": "text field is required" }]
}
```

_(Status: 400 Bad Request)_

### 3. Security: Unauthorized

- **Headers**: No Bearer token provided.
- **Expected Response**:

```json
{
  "status": false,
  "message": "Unauthorized"
}
```

_(Status: 401 Unauthorized)_
