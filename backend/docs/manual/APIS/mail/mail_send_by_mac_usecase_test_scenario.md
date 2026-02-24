# ENDPOINT: POST /api/email/send/mac/:mac_address

## Description
Looks up customer information via MAC address and submits an email task asynchronously. Returns a `task_id` immediately (HTTP 202 Accepted). Poll `/api/email/status/{task_id}` to track completion.

## Authentication
- **Type**: BearerAuth
- **Header**: `Authorization: Bearer <token>`

## Parameters
- **Path Parameter**:
  - `mac_address` (string): The MAC address (or UUID) of the Teralux device.

## Request Body
- **Content-Type**: `application/json`
- **Required Fields**:
  - `subject` (string): The subject line of the email.
- **Optional Fields**:
  - `template` (string): The name of the server-side template (defaults to `"test"`).
  - `attachment_path` (string): Server-side path to a PDF file to attach.

## Test Scenarios

### 1. Success - Submit Task for Valid Device
- **URL**: `http://localhost:8080/api/email/send/mac/db329671-96bb-368b-95d3-53a3a3712563`
- **Method**: `POST`
- **Pre-conditions**: Device with that MAC exists in the external system with a valid customer email.
- **Request Body**:
```json
{
  "subject": "Upcoming Booking Notification",
  "template": "test"
}
```
- **Expected Response** *(Status: 202 Accepted)*:
```json
{
  "status": true,
  "message": "Email task submitted successfully",
  "data": {
    "task_id": "mail-mac-abc123",
    "task_status": "pending"
  }
}
```

### 2. Error: Customer Email Not Found
- **URL**: `http://localhost:8080/api/email/send/mac/NON_EXISTENT_MAC`
- **Request Body**: `{ "subject": "Test" }`
- **Expected Response** *(Status: 404 Not Found)*:
```json
{
  "status": false,
  "message": "Customer email not found for this device"
}
```

### 3. Validation: Empty MAC Address
- **URL**: `http://localhost:8080/api/email/send/mac/`
- **Expected Response** *(Status: 400 Bad Request)*:
```json
{
  "status": false,
  "message": "mac_address is required"
}
```

### 4. Security: Unauthorized
- **Headers**: No Bearer token.
- **Expected Response** *(Status: 401 Unauthorized)*:
```json
{
  "status": false,
  "message": "Unauthorized"
}
```
