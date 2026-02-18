# ENDPOINTS: Device Control (Switch & IR)

## Description
Controls Tuya devices by sending commands. This documentation covers two types of controls:
1. **Standard Switch Control**: For standard devices like lights, switches, and breakers.
2. **Infrared (IR) Control**: For IR-controlled devices like ACs and TVs via an IR Blaster hub.

---

## 1. Standard Switch Control
**URL**: `POST /api/tuya/devices/{id}/commands/switch`

### Authentication
- **Type**: BearerAuth
- **Header**: `Authorization: Bearer <token>`

### Path Parameters
- `id` (string, required) - Target device ID

### Test Scenarios

#### 1.1 Control Switch (Success)
- **Method**: `POST`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: Device supports switch commands.
- **Request Body**:
```json
{
  "code": "switch_1",
  "value": true
}
```
- **Expected Response**:
```json
{
  "status": true,
  "message": "Command sent successfully",
  "data": true
}
```
  *(Status: 200 OK)*
- **Side Effects**: Physical device turns ON.

#### 1.2 Validation: Missing Body
- **Method**: `POST`
- **Request Body**: `{}`
- **Expected Response**:
```json
{
  "status": false,
  "message": "Validation Error",
  "details": [
    { "field": "body", "message": "invalid request body" }
  ]
}
```
  *(Status: 400 Bad Request)*

---

## 2. Infrared (IR) Control
**URL**: `POST /api/tuya/devices/{id}/commands/ir`

### Authentication
- **Type**: BearerAuth
- **Header**: `Authorization: Bearer <token>`

### Path Parameters
- `id` (string, required) - IR Blaster (Hub) device ID

### Test Scenarios

#### 2.1 Control IR Device (AC/TV) (Success)
- **Method**: `POST`
- **Headers**:
```json
{
  "Content-Type": "application/json",
  "Authorization": "Bearer <valid_token>"
}
```
- **Pre-conditions**: Device is an IR Blaster and target `remote_id` is configured.
- **Request Body**:
```json
{
  "remote_id": "remote_123",
  "code": "PowerOn",
  "value": "1"
}
```
- **Expected Response**:
```json
{
  "status": true,
  "message": "IR command sent successfully"
}
```
  *(Status: 200 OK)*
- **Side Effects**: IR Blaster emits the specific "PowerOn" signal to the remote device.

#### 2.2 Validation: Missing Remote ID
- **Method**: `POST`
- **Request Body**: 
```json
{
  "code": "PowerOn",
  "value": "1"
}
```
- **Expected Response**:
```json
{
  "status": false,
  "message": "Validation Error",
  "details": [
    { "field": "remote_id", "message": "remote_id is required" }
  ]
}
```
  *(Status: 400 Bad Request)*

---

## Common Security & Error Scenarios (Both Endpoints)

### 3.1 Security: Unauthorized
- **Headers**: No Authorization header.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Unauthorized"
}
```
  *(Status: 401 Unauthorized)*

### 3.2 Error: Device Not Found
- **Pre-conditions**: Invalid device ID provided.
- **Expected Response**:
```json
{
  "status": false,
  "message": "Device not found"
}
```
  *(Status: 500 Internal Server Error)*

### 3.3 Error: Tuya Cloud Error
- **Pre-conditions**: Valid request but Tuya Cloud returns an error (e.g., device offline).
- **Expected Response**:
```json
{
  "status": false,
  "message": "Failed to send command to Tuya"
}
```
  *(Status: 500 Internal Server Error)*
