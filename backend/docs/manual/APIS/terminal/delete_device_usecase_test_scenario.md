# ENDPOINT: DELETE /api/devices/:id

## Description
Remove a device from the system.

## Test Scenarios

### 1. Delete Device (Success)
- **URL**: `http://localhost:8080/api/devices/dev-1`
- **Method**: `DELETE`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: Device `dev-1` exists.
- **Request Body**: None.
- **Expected Response**:
```json
{
  "status": true,
  "message": "Device deleted successfully"
}
```
  *(Status: 200 OK)*
- **Side Effects**:
  - Device record removed.
  - Associated logs or statuses might be archived or deleted (depending on policy).

### 2. Delete Device (Not Found)
- **URL**: `http://localhost:8080/api/devices/dev-999`
- **Method**: `DELETE`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: Device `dev-999` does not exist.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Device not found"
}
```
  *(Status: 404 Not Found)*

### 3. Validation: Invalid ID Format
- **URL**: `http://localhost:8080/api/devices/INVALID`
- **Method**: `DELETE`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Expected Response**:
```json
{
  "status": false,
  "message": "Invalid ID format"
}
```
  *(Status: 400 Bad Request)*

### 4. Security: Unauthorized
- **URL**: `http://localhost:8080/api/devices/dev-1`
- **Method**: `DELETE`
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
