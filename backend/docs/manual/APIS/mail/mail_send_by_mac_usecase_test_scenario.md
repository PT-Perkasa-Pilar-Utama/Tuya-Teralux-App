# ENDPOINT: POST /api/mail/send/mac/:mac_address

## Description
Looks up customer information (including their email address) based on the provided Teralux MAC address and sends an email using a server-side HTML template.

## Authentication
- **Type**: BearerAuth
- **Header**: `Authorization: Bearer <token>`

## Parameters
- **Path Parameter**:
  - `mac_address` (string): The MAC address of the Teralux device. (Any valid identifier used by the external system).

## Request Body
- **Content-Type**: `application/json`
- **Required Fields**:
  - `subject` (string): The subject line of the email.
- **Optional Fields**:
  - `template` (string): The name of the server-side template to use (defaults to "test").

## Test Scenarios

### 1. Success - Send Email to Customer
- **URL**: `http://localhost:8080/api/mail/send/mac/db329671-96bb-368b-95d3-53a3a3712563`
- **Method**: `POST`
- **Pre-conditions**:
  - Device with MAC `db329671-96bb-368b-95d3-53a3a3712563` exists in the external system with a valid customer email.
- **Request Body**:
```json
{
  "subject": "Upcoming Booking Notification",
  "template": "test"
}
```
- **Expected Response**:
```json
{
  "status": true,
  "message": "Email sent successfully to customer",
  "data": {
    "customer_email": "customer@example.com"
  }
}
```
*(Status: 200 OK)*

### 2. Error: Customer Email Not Found
- **URL**: `http://localhost:8080/api/mail/send/mac/NON_EXISTENT_MAC`
- **Method**: `POST`
- **Pre-conditions**: Device exists but has no associated customer email or MAC doesn't exist.
- **Request Body**:
```json
{
  "subject": "Test"
}
```
- **Expected Response**:
```json
{
  "status": false,
  "message": "Customer email not found for this device"
}
```
*(Status: 404 Not Found)*

### 4. Security: Unauthorized
- **URL**: `http://localhost:8080/api/mail/send/mac/db329671-96bb-368b-95d3-53a3a3712563`
- **Headers**: No Bearer token.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Unauthorized"
}
```
*(Status: 401 Unauthorized)*
