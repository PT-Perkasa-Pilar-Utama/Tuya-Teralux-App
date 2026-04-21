# Smart Door Lock Command Reference

## Quick Start

```bash
# Edit the script
nano send_command.py

# Run it
python3 send_command.py
```

## Available Commands

### Integer Commands (0-999)

| Command | Min | Max | Description |
|---------|-----|-----|-------------|
| `unlock_fingerprint` | 0 | 999 | Fingerprint unlock counter (read-only) |
| `unlock_password` | 0 | 999 | Password unlock counter (read-only) |
| `unlock_temporary` | 0 | 999 | Temporary password counter (read-only) |
| `unlock_dynamic` | 0 | 999 | Dynamic password counter (read-only) |
| `unlock_card` | 0 | 999 | Card unlock counter (read-only) |
| `unlock_face` | 0 | 999 | Face unlock counter (read-only) |
| `unlock_key` | 0 | 999 | Physical key counter (read-only) |
| `unlock_request` | 0 | 90 | Remote unlock request in seconds ⚠️ |
| `unlock_app` | 0 | 999 | App unlock counter (read-only) |
| `residual_electricity` | 0 | 100 | Battery level percentage (read-only) |

⚠️ **Note:** `unlock_request` may not work on all devices

### Boolean Commands

| Command | Values | Description |
|---------|--------|-------------|
| `hijack` | true/false | Duress alarm status |
| `open_inside` | true/false | Inside handle status |
| `doorbell` | true/false | Doorbell pressed status |
| `anti_lock_outside` | true/false | Anti-lock status |
| `reply_unlock_request` | true/false | Reply to unlock request ✅ |

✅ **Tested:** `reply_unlock_request` works on this device

### Enum Commands

| Command | Values | Description |
|---------|--------|-------------|
| `closed_opened` | "unknown", "open", "closed" | Door position status |

### List Commands

| Command | Format | Description |
|---------|--------|-------------|
| `update_all_finger` | [] | Update all fingerprints |
| `update_all_password` | [] | Update all passwords |
| `update_all_card` | [] | Update all cards |
| `update_all_face` | [] | Update all faces |

---

## Usage Examples

### Example 1: Reply to Unlock Request

```python
COMMANDS = [
    {"code": "reply_unlock_request", "value": True},
]
```

### Example 2: Check Door Status

```python
COMMANDS = [
    {"code": "closed_opened", "value": "open"},
]
```

### Example 3: Trigger Doorbell

```python
COMMANDS = [
    {"code": "doorbell", "value": True},
]
```

### Example 4: Clear All Passwords

```python
COMMANDS = [
    {"code": "update_all_password", "value": []},
]
```

### Example 5: Multiple Commands

```python
COMMANDS = [
    {"code": "doorbell", "value": True},
    {"code": "closed_opened", "value": "closed"},
]
```

---

## Expected Responses

### Success Response
```json
{
  "result": true,
  "success": true,
  "t": 1776408410680,
  "tid": "366fc5c73a2911f1a1dff2a53f931937"
}
```

### Error Response - Command Not Supported
```json
{
  "code": 2008,
  "msg": "command or value not support",
  "success": false,
  "t": 1776408405046,
  "tid": "3314fdcd3a2911f1a1dff2a53f931937"
}
```

### Error Response - Permission Denied
```json
{
  "code": 1106,
  "msg": "permission deny",
  "success": false,
  "t": 1776408405046,
  "tid": "3314fdcd3a2911f1a1dff2a53f931937"
}
```

### Error Response - Device Offline
```json
{
  "code": 2001,
  "msg": "device is offline",
  "success": false,
  "t": 1776408405046,
  "tid": "3314fdcd3a2911f1a1dff2a53f931937"
}
```

---

## Error Codes Reference

| Code | Message | Solution |
|------|---------|----------|
| 2008 | command or value not support | Command not available on this device |
| 1106 | permission deny | Device not linked to cloud project |
| 2001 | device is offline | Check WiFi connection |
| 1109 | param is illegal | Invalid parameter value |
| 1004 | sign invalid | Check API credentials |

---

## Device: Smart Doorlock (Lite Version)

**Device ID:** `a328e47b067a5c5b6bc8k6`  
**Category:** `ms`  
**Model:** DL-WH-HV

### Tested Commands ✅

| Command | Value | Status |
|---------|-------|--------|
| `reply_unlock_request` | `true` | ✅ Works |
| `reply_unlock_request` | `false` | ✅ Works |

### Failed Commands ❌

| Command | Value | Error |
|---------|-------|-------|
| `unlock_request` | `1-90` | 2008 - not support |
| `add_password` | `"123456"` | 2008 - not support |
| `update_all_password` | `[...]` | 2008 - not support |
| `closed_opened` | `"open"` | 2008 - not support |
| `doorbell` | `true` | 2008 - not support |

---

## Limitations

This device (category `ms`) is **read-only** for most commands:

❌ **Cannot:**
- Create temporary passwords via API
- Add/remove passwords via API
- Remote unlock via API
- Control door lock directly

✅ **Can:**
- Monitor unlock events (counters)
- Monitor battery status
- Monitor door open/close status
- Reply to unlock requests (but doesn't actually unlock)

---

## Alternative Solutions

### 1. Use Smart Life App (Manual)
For creating temporary passwords, use the Smart Life app directly.

### 2. Automate via Playwright
Use browser automation to interact with Smart Life web interface.

### 3. Upgrade Device
Get a `jtmspro` category device (e.g., U688S-WiFi-Pro) for full API support.

---

**Last Updated:** April 17, 2026
