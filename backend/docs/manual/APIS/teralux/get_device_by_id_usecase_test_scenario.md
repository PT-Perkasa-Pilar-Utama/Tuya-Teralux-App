# ENDPOINT: GET /api/devices/:id

## Description
Retrieve a single device by its ID.

## Test Scenarios

### 1. Get Device By ID (Success)
- **URL**: `http://localhost:8080/api/devices/dev-1`
- **Method**: `GET`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: Device `dev-1` exists.
- **Expected Response**:
```json
{
  "status": true,
  "message": "Device retrieved successfully",
  "data": {
    "device": {
      "id": "dev-1",
      "name": "Kitchen Switch",
      "teralux_id": "tx-1",
      "remote_id": "rem-1",
      "category": "sw",
      "remote_category": "switch",
      "product_name": "Tuya Switch",
      "remote_product_name": "Smart Switch",
      "icon": "icon_url",
      "custom_name": "Kitchen",
      "model": "T1",
      "ip": "192.168.1.1",
      "local_key": "key123",
      "gateway_id": "gw-1",
      "create_time": 1600000000,
      "update_time": 1600000000,
      "collections": "[]",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  }
}
```
  *(Status: 200 OK)*

### 2. Get Device By ID (Not Found)
- **URL**: `http://localhost:8080/api/devices/dev-unknown`
- **Method**: `GET`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: Device does not exist.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Device not found"
}
```
  *(Status: 404 Not Found)*

### 3. Security: Unauthorized
- **URL**: `http://localhost:8080/api/devices/dev-1`
- **Method**: `GET`
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
