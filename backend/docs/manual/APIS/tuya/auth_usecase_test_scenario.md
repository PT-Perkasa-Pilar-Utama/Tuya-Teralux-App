# ENDPOINT: GET /api/tuya/auth

## Description
Initiates the Tuya authentication process to retrieve an access token.

## Test Scenarios

### 1. Authenticate (Success)
- **URL**: `http://localhost:8080/api/tuya/auth`
- **Method**: `GET`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json",
    "Authorization": "Bearer <valid_token>"
  }
  ```
- **Pre-conditions**:
  - Tuya credentials (`TUYA_CLIENT_ID`, `TUYA_SECRET`) are correctly configured in `.env`.
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

### 2. Security: Unauthorized
- **URL**: `http://localhost:8080/api/tuya/auth`
- **Method**: `GET`
- **Headers**:
  ```json
  {
    "Content-Type": "application/json"
    // Missing Authorization
  }
  ```
- **Pre-conditions**: No Auth Token.
- **Expected Response**:
  ```json
  {
    "status": false,
    "message": "Unauthorized"
  }
  ```
  *(Status: 401 Unauthorized)*
