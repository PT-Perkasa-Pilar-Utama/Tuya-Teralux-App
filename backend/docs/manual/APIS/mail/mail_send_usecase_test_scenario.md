# ENDPOINT: POST /api/email/send

## Description
Submits an email sending task asynchronously. Returns a `task_id` immediately (HTTP 202 Accepted). The actual email is sent in the background. Use the status endpoint to poll for completion.

## Authentication
- **Type**: BearerAuth
- **Header**: `Authorization: Bearer <token>`
- **Note**: Requires a valid JWT token.

## Request Body
- **Content-Type**: `application/json`
- **Required Fields**:
  - `to` ([]string): List of recipient email addresses.
  - `subject` (string): The subject line of the email.
- **Optional Fields**:
  - `template` (string): The name of the server-side template to use (defaults to `"test"`).
  - `attachment_path` (string): Server-side path to a PDF file to attach (e.g. `/uploads/reports/summary_123.pdf`).

## Test Scenarios

### 1. Success - Submit Task with Default Template
- **URL**: `http://localhost:8080/api/email/send`
- **Method**: `POST`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**:
  - Valid SMTP configuration in `.env`.
  - `test.html` template exists in `domain/mail/templates`.
- **Request Body**:
```json
{
  "to": ["user@example.com"],
  "subject": "System Notification"
}
```
- **Expected Response** *(Status: 202 Accepted)*:
```json
{
  "status": true,
  "message": "Email task submitted successfully",
  "data": {
    "task_id": "mail-abc-123xyz",
    "task_status": "pending"
  }
}
```
- **Side Effects**: Task is queued. Poll `/api/email/status/{task_id}` for the result.

### 2. Success - Submit Task with PDF Attachment
- **URL**: `http://localhost:8080/api/email/send`
- **Method**: `POST`
- **Request Body**:
```json
{
  "to": ["boss@example.com"],
  "subject": "Weekly Progress Summary",
  "template": "summary",
  "attachment_path": "/uploads/reports/summary_abc.pdf"
}
```
- **Expected Response** *(Status: 202 Accepted)*:
```json
{
  "status": true,
  "message": "Email task submitted successfully",
  "data": {
    "task_id": "mail-xyz-456abc",
    "task_status": "pending"
  }
}
```

### 3. Validation: Missing Required Fields
- **Request Body**:
```json
{
  "to": [],
  "subject": ""
}
```
- **Expected Response** *(Status: 400 Bad Request)*:
```json
{
  "status": false,
  "message": "Validation Error",
  "details": [
    { "field": "to", "message": "to field is required" }
  ]
}
```

### 4. Error: Template Not Found
- **Request Body**:
```json
{
  "to": ["user@example.com"],
  "subject": "Test",
  "template": "ghost_template"
}
```
- **Expected Response** *(Status: 500 Internal Server Error)*:
```json
{
  "status": false,
  "message": "Internal Server Error"
}
```

### 5. Security: Unauthorized
- **Headers**: No Bearer token provided.
- **Expected Response** *(Status: 401 Unauthorized)*:
```json
{
  "status": false,
  "message": "Unauthorized"
}
```
