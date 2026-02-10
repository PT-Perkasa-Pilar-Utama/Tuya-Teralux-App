# ENDPOINT: POST /api/speech/mqtt/publish

## Description
Publishes a message to the pre-configured MQTT topic: `users/teralux/whisper`. Access is restricted to authorized clients via Bearer token.

## Authentication
- **Type**: BearerAuth
- **Header**: `Authorization: Bearer <token>`

## Request Body
- **Content-Type**: `application/json`
- **Required Fields**:
  - `message` (string): The text content to publish.

## Test Scenarios

### 1. Publish Message (Success)
- **Method**: `POST`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: MQTT broker is connected and client is authorized.
- **Request Body**:
```json
{
  "message": "Transcription completed: Hello world"
}
```
- **Expected Response**:
```json
{
  "status": true,
  "message": "Message published to MQTT successfully"
}
```
  *(Status: 200 OK)*
- **Side Effects**: Message is received by all clients subscribed to `users/teralux/whisper`.

### 2. Validation: Missing Message
- **Method**: `POST`
- **Request Body**: `{}`
- **Expected Response**:
```json
{
  "status": false,
  "message": "Message field is required"
}
```
  *(Status: 400 Bad Request)*

### 3. Error: MQTT Broker Disconnected
- **Pre-conditions**: The backend service has lost connection to the MQTT broker.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Failed to publish message to MQTT"
}
```
  *(Status: 500 Internal Server Error)*

### 4. Security: Unauthorized
- **Headers**: Missing or invalid Bearer token.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Unauthorized"
}
```
  *(Status: 401 Unauthorized)*
