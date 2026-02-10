# ENDPOINT: PUT /api/devices/:id

## Description
Update a device's information (Name, Type).

## Test Scenarios

### 1. Update Device (Success)
- **URL**: `http://localhost:8080/api/devices/dev-1`
- **Method**: `PUT`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: Device `dev-1` exists.
- **Request Body**:
```json
{
  "name": "Kitchen Sink Light"
}
```
- **Expected Response**:
```json
{
  "status": true,
  "message": "Device updated successfully"
}
```
  *(Status: 200 OK)*
- **Side Effects**: Name updated.

### 2. Update Device (Not Found)
- **URL**: `http://localhost:8080/api/devices/dev-unknown`
- **Method**: `PUT`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Request Body**:
```json
{ "name": "New Name" }
```
- **Expected Response**:
```json
{
  "status": false,
  "message": "Device not found"
}
```
  *(Status: 404 Not Found)*

### 3. Validation: Empty Name
- **URL**: `http://localhost:8080/api/devices/dev-1`
- **Method**: `PUT`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Request Body**:
```json
{ "name": "" }
```
- **Expected Response**:
```json
{
  "status": false,
  "message": "Validation Error: name cannot be empty"
}
```
  *(Status: 422 Unprocessable Entity)*

### 4. Security: Unauthorized
- **URL**: `http://localhost:8080/api/devices/dev-1`
- **Method**: `PUT`
- **Headers**:
```json
{
  "Content-Type": "application/json"
  // Missing Authorization
}
```
- **Expected Response**:
```json
{ "status": false, "message": "Unauthorized" }
```
  *(Status: 401 Unauthorized)*
