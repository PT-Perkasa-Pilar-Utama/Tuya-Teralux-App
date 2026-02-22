# ENDPOINT: POST /api/mail/send

## Description
Sends an email using a server-side HTML template. This endpoint allows specifying recipients and the subject, while the content is rendered from predefined templates.

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
  - `template` (string): The name of the server-side template to use (defaults to "test").

## Test Scenarios

### 1. Success - Send with Default Template
- **URL**: `http://localhost:8080/api/mail/send`
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
  - "test.html" template exists in `domain/mail/templates`.
- **Request Body**:
```json
{
  "to": ["user@example.com"],
  "subject": "System Notification"
}
```
- **Expected Response**:
```json
{
  "status": true,
  "message": "Email sent successfully"
}
```
  *(Status: 200 OK)*
- **Side Effects**: Email sent to recipient.

### 2. Success - Send with Specific Template
- **URL**: `http://localhost:8080/api/mail/send`
- **Method**: `POST`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**:
  - "summary.html" template exists in `domain/mail/templates`.
- **Request Body**:
```json
{
  "to": ["boss@example.com"],
  "subject": "Weekly Progress Summary",
  "template": "summary"
}
```
- **Expected Response**:
```json
{
  "status": true,
  "message": "Email sent successfully"
}
```
  *(Status: 200 OK)*

### 3. Validation: Missing Required Fields
- **URL**: `http://localhost:8080/api/mail/send`
- **Method**: `POST`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: Missing `to` or `subject` field.
- **Request Body**:
```json
{
  "to": [],
  "subject": ""
}
```
- **Expected Response**:
```json
{
  "status": false,
  "message": "Validation Error",
  "details": [
    { "field": "to", "message": "to field is required" }
  ]
}
```
  *(Status: 400 Bad Request)*

### 4. Error: Template Not Found
- **URL**: `http://localhost:8080/api/mail/send`
- **Method**: `POST`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: Template file does not exist on disk.
- **Request Body**:
```json
{
  "to": ["user@example.com"],
  "subject": "Test",
  "template": "ghost_template"
}
```
- **Expected Response**:
```json
{
  "status": false,
  "message": "Internal Server Error"
}
```
  *(Status: 500 Internal Server Error)*

### 5. Security: Unauthorized
- **URL**: `http://localhost:8080/api/mail/send`
- **Headers**: No Bearer token provided.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Unauthorized"
}
```
*(Status: 401 Unauthorized)*
