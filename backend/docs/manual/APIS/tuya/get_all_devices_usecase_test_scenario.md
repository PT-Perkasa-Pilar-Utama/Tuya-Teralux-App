# ENDPOINT: GET /api/tuya/devices

## Description
Retrieves a list of all devices in a **Merged View** (Smart IR remotes merged with Hubs). Devices are sorted alphabetically by name. For `infrared_ac` devices, the status array is automatically populated with saved device state (power, temp, mode, wind) or default values if no state exists.

## Authentication
- **Type**: BearerAuth
- **Header**: `Authorization: Bearer <token>`

## Query Parameters
- `page` (integer, optional) - Page number for pagination (default: 1)
- `limit` (integer, optional) - Items per page (default: all)
- `per_page` (integer, optional) - Alias for `limit`
- `category` (string, optional) - Filter by device category (e.g., "dj", "kg", "infrared_ac")

## Special Features
- **Merged View**: Smart IR remotes are merged with their parent Hub devices in the `collections` array
- **Alphabetical Sorting**: Devices are sorted by name (A-Z)
- **IR AC Auto-Population**: For `infrared_ac` devices, the `status` array contains saved state or defaults
- **Nested Collections**: Hub devices include child devices in the `collections` field

## Test Scenarios

### 1. Get All Devices (Success - No Pagination)
- **URL**: `http://localhost:8080/api/tuya/devices`
- **Method**: `GET`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**:
  - Tuya account is linked and devices are synchronized.
- **Request Body**: None.
- **Expected Response**:
```json
{
  "status": true,
  "message": "Devices retrieved successfully",
  "data": {
    "devices": [
      {
        "id": "bf0d2...",
        "name": "AC Living Room",
        "category": "infrared_ac",
        "product_name": "Smart IR AC",
        "online": true,
        "status": [
          {"code": "power", "value": true},
          {"code": "temp", "value": 24},
          {"code": "mode", "value": 1},
          {"code": "wind", "value": 2}
        ],
        "remote_id": "remote_123",
        "remote_category": "infrared_ac",
        "remote_product_name": "AC Remote",
        "gateway_id": "hub_456",
        "collections": [],
        "icon": "https://...",
        "create_time": 1678888000,
        "update_time": 1678888000
      },
      {
        "id": "hub_456",
        "name": "Smart Hub",
        "category": "wg2",
        "product_name": "Zigbee Hub",
        "online": true,
        "status": [],
        "collections": [
          {
            "id": "remote_123",
            "name": "IR Remote",
            "category": "wnykq",
            "product_name": "IR Controller",
            "online": true
          }
        ],
        "icon": "https://...",
        "ip": "192.168.1.100",
        "local_key": "abc123...",
        "create_time": 1678888000,
        "update_time": 1678888000
      }
    ],
    "page": 1,
    "per_page": 100,
    "total": 2,
    "total_devices": 2,
    "current_page_count": 2
  }
}
```
  *(Status: 200 OK)*
- **Side Effects**: None (read-only).

### 2. Get Devices with Pagination
- **URL**: `http://localhost:8080/api/tuya/devices?page=1&limit=10`
- **Method**: `GET`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: Multiple devices exist.
- **Expected Response**:
```json
{
  "status": true,
  "message": "Devices retrieved successfully",
  "data": {
    "devices": [/* first 10 devices */],
    "page": 1,
    "per_page": 10,
    "total": 50,
    "total_devices": 50,
    "current_page_count": 10
  }
}
```
  *(Status: 200 OK)*
- **Side Effects**: None.

### 3. Filter by Category
- **URL**: `http://localhost:8080/api/tuya/devices?category=infrared_ac`
- **Method**: `GET`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: Devices with category "infrared_ac" exist.
- **Expected Response**:
```json
{
  "status": true,
  "message": "Devices retrieved successfully",
  "data": {
    "devices": [/* only infrared_ac devices */],
    "page": 1,
    "per_page": 100,
    "total": 5,
    "total_devices": 5,
    "current_page_count": 5
  }
}
```
  *(Status: 200 OK)*
- **Side Effects**: None.

### 4. Combine Pagination and Filter
- **URL**: `http://localhost:8080/api/tuya/devices?page=2&limit=5&category=dj`
- **Method**: `GET`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: More than 5 devices with category "dj" exist.
- **Expected Response**:
```json
{
  "status": true,
  "message": "Devices retrieved successfully",
  "data": {
    "devices": [/* devices 6-10 of category dj */],
    "page": 2,
    "per_page": 5,
    "total": 15,
    "total_devices": 15,
    "current_page_count": 5
  }
}
```
  *(Status: 200 OK)*
- **Side Effects**: None.

### 5. Security: Unauthorized
- **URL**: `http://localhost:8080/api/tuya/devices`
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

### 6. Error: Server Error
- **URL**: `http://localhost:8080/api/tuya/devices`
- **Method**: `GET`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: Tuya Cloud service is down or database unavailable.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Failed to retrieve devices"
}
```
  *(Status: 500 Internal Server Error)*
