# ENDPOINT: DELETE /api/cache/flush

## Description
Removes all data from the absolute persistent cache storage (Badger DB). This operation is destructive and clears all cached data across all domains (Speech, RAG, Tuya, etc).

## Authentication
- **Type**: BearerAuth
- **Header**: `Authorization: Bearer <token>`
- **Note**: Access is restricted to authorized administrative users.

## Test Scenarios

### 1. Flush Cache (Success)
- **Method**: `DELETE`
- **Headers**:
```json
{
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: Cache service is active and contains data.
- **Request Body**: None.
- **Expected Response**:
```json
{
  "status": true,
  "message": "Cache flushed successfully",
  "data": null
}
```
  *(Status: 200 OK)*
- **Side Effects**: 
  - All pending transcription results are deleted.
  - All RAG task data is removed.
  - All cached device status and sensor mappings are cleared.
  - Storage directory is purged.

### 2. Flush Empty Cache (Success)
- **Method**: `DELETE`
- **Pre-conditions**: Cache is already empty.
- **Expected Response**:
```json
{
  "status": true,
  "message": "Cache flushed successfully"
}
```
  *(Status: 200 OK)*

### 3. Security: Unauthorized
- **Headers**: Missing or invalid token.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Unauthorized"
}
```
  *(Status: 401 Unauthorized)*

### 4. Error: Internal Storage Failure
- **Pre-conditions**: Badger DB process is locked or filesystem is read-only.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Failed to flush cache"
}
```
  *(Status: 500 Internal Server Error)*
