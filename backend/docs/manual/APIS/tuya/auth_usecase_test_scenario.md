# ENDPOINT: GET /api/tuya/auth

## Description
Initiates the Tuya authentication process to retrieve an access token. This endpoint uses API Key authentication (not Bearer token) to obtain a Tuya access token for subsequent API calls.

## Authentication
- **Type**: ApiKeyAuth
- **Header**: `X-API-KEY: <api_key>`
- **Note**: This endpoint is used to GET a Tuya access token, not to validate a user Bearer token

## Test Scenarios

### 1. Authenticate (Success)
- **URL**: `http://localhost:8080/api/tuya/auth`
- **Method**: `GET`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "X-API-KEY": "<api_key>"
}
```
- **Pre-conditions**:
  - Tuya credentials (`TUYA_CLIENT_ID`, `TUYA_SECRET`) are correctly configured in `.env`.
  - Valid API key provided in `X-API-KEY` header.
- **Request Body**: None.
- **Expected Response**:
```json
{
  "status": true,
  "message": "Authenticated successfully",
  "data": {
    "access_token": "<tuya_access_token>",
    "expire_time": 7200,
    "refresh_token": "<tuya_refresh_token>",
    "uid": "<tuya_uid>"
  }
}
```
  *(Status: 200 OK)*
- **Side Effects**:
  - Tuya tokens might be cached or refreshed.

### 2. Security: Unauthorized (Missing API Key)
- **URL**: `http://localhost:8080/api/tuya/auth`
- **Method**: `GET`
- **Headers**:
```json
{
  "Content-Type": "application/json"
  // Missing X-API-KEY
}
```
- **Pre-conditions**: No API Key provided.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Unauthorized"
}
```
  *(Status: 401 Unauthorized)*

### 3. Error: Invalid API Key
- **URL**: `http://localhost:8080/api/tuya/auth`
- **Method**: `GET`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "X-API-KEY": "invalid_key_12345"
}
```
- **Pre-conditions**: Invalid API key provided.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Unauthorized"
}
```
  *(Status: 401 Unauthorized)*

### 4. Error: Tuya Service Unavailable
- **URL**: `http://localhost:8080/api/tuya/auth`
- **Method**: `GET`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "X-API-KEY": "<valid_api_key>"
}
```
- **Pre-conditions**: Valid API key, but Tuya Cloud service is down or credentials are invalid.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Failed to authenticate with Tuya",
  "details": "error: connection refused"
}
```
  *(Status: 500 Internal Server Error)*
