# ENDPOINT: DELETE /api/terminal/:id

## Description
Delete an existing Terminal device by its ID.

## Test Scenarios

### 1. Delete Terminal (Success Condition)
- **URL**: `http://localhost:8080/api/terminal/tx-1`
- **Method**: `DELETE`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**:
  - Device `tx-1` exists.
- **Request Body**: None.
- **Expected Response**:
```json
{
  "status": true,
  "message": "Terminal deleted successfully",
  "data": null
}
```
  *(Status: 200 OK)*
- **Side Effects**:
  - Device `tx-1` removed from database.

### 2. Delete Terminal (Not Found)
- **URL**: `http://localhost:8080/api/terminal/tx-999`
- **Method**: `DELETE`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**:
  - Device `tx-999` does not exist.
- **Request Body**: None.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Terminal not found",
  "data": null
}
```
  *(Status: 404 Not Found)*

### 3. Validation: Invalid ID Format
- **URL**: `http://localhost:8080/api/terminal/INVALID-UUID`
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
- **URL**: `http://localhost:8080/api/terminal/tx-1`
- **Method**: `DELETE`
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
