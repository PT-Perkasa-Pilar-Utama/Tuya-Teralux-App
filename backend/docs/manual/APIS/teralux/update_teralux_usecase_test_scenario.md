# ENDPOINT: PUT /api/teralux/:id

## Description
Update an existing Teralux device's information.

## Test Scenarios

### 1. Update Teralux (Success - Name Only)
- **URL**: `http://localhost:8080/api/teralux/t1`
- **Method**: `PUT`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: Device `t1` exists.
- **Request Body**:
```json
{
  "name": "New Name"
}
```
- **Expected Response**:
```json
{
  "status": true,
  "message": "Updated successfully",
  "data": null
}
```
  *(Status: 200 OK)*
- **Side Effects**: Name updated, MAC and Room ID unchanged.

### 2. Update Teralux (Success - Move Room)
- **URL**: `http://localhost:8080/api/teralux/t1`
- **Method**: `PUT`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**:
  - Device `t1` exists (currently in `room-1`).
  - `room-2` exists.
- **Request Body**:
```json
{
  "room_id": "room-2"
}
```
- **Expected Response**:
```json
{
  "status": true,
  "message": "Updated successfully"
}
```
  *(Status: 200 OK)*
- **Side Effects**: Device moved to `room-2`.

### 3. Update Teralux (Not Found)
- **URL**: `http://localhost:8080/api/teralux/unknown`
- **Method**: `PUT`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: Device `unknown` does not exist.
- **Request Body**:
```json
{ "name": "Hack" }
```
- **Expected Response**:
```json
{
  "status": false,
  "message": "Teralux not found"
}
```
  *(Status: 404 Not Found)*

### 4. Validation: Invalid Room ID
- **URL**: `http://localhost:8080/api/teralux/t1`
- **Method**: `PUT`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**:
  - Device `t1` exists.
  - `room-999` does not exist.
- **Request Body**:
```json
{ "room_id": "room-999" }
```
- **Expected Response**:
```json
{
  "status": false,
  "message": "Validation Error",
  "details": [
    { "field": "room_id", "message": "room does not exist" }
  ]
}
```
  *(Status: 422 Unprocessable Entity)*

### 5. Validation: Empty Name (If Present)
- **URL**: `http://localhost:8080/api/teralux/t1`
- **Method**: `PUT`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: Device `t1` exists.
- **Request Body**:
```json
{ "name": "" }
```
- **Expected Response**:
```json
{
  "status": false,
  "message": "Validation Error",
  "details": [
    { "field": "name", "message": "name cannot be empty" }
  ]
}
```
  *(Status: 422 Unprocessable Entity)*

### 6. Conflict: Update to Duplicate MAC
- **URL**: `http://localhost:8080/api/teralux/t1`
- **Method**: `PUT`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**:
  - Device `t1` exists.
  - Device `t2` exists with MAC `MAC-2`.
- **Request Body**:
```json
{ "mac_address": "MAC-2" }
```
- **Expected Response**:
```json
{
  "status": false,
  "message": "Mac Address already in use"
}
```
  *(Status: 409 Conflict)*
- **Side Effects**: No changes made.

### 7. Security: Unauthorized
- **URL**: `http://localhost:8080/api/teralux/t1`
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
{
  "status": false,
  "message": "Unauthorized"
}
```
  *(Status: 401 Unauthorized)*
