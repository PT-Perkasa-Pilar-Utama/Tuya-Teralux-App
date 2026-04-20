# Smart Door Lock - Quick Unlock Script

## Usage

```bash
# Unlock (reply to unlock request)
./unlock.sh true

# Cancel/Reject
./unlock.sh false
```

## What it does

This script sends a `reply_unlock_request` command to your Smart Door Lock (Lite Version).

**Note:** This device (category `ms`) is a Lite Version that:
- ✅ Can acknowledge unlock requests
- ❌ Cannot directly unlock the door via API
- ❌ Cannot create passwords via API

## How it works

1. Gets access token from Tuya API
2. Sends command to device
3. Returns success/failure status

## Requirements

- Bash shell
- Python 3
- `requests` library (`pip3 install requests`)
- Valid `.env` file with Tuya credentials

## Example Output

```
==============================================================
  SMART DOOR LOCK - REPLY UNLOCK REQUEST
==============================================================

  Device: a328e47b067a5c5b6bc8k6
  Value:  true

=== Step 1: Get Access Token ===
✅ Access Token: ffa5ddc839b5b68831bb...

=== Step 2: Send Command ===
Response: {
  "result": true,
  "success": true,
  "t": 1776412408695,
  "tid": "856e7e3e3a3211f1a1dff2a53f931937"
}

==============================================================
  RESULT
==============================================================
✅ Command sent successfully!

   Reply Unlock Request: true

💡 Note: This command acknowledges an unlock request.
   It may not actually unlock the door depending on device.
```

## Error Codes

| Code | Meaning | Solution |
|------|---------|----------|
| 2008 | Command not supported | Device doesn't support this command |
| 1106 | Permission denied | Check device linkage to cloud project |
| 2001 | Device offline | Check WiFi connection |

## Setup

1. Make sure `.env` file exists with credentials:
   ```
   TUYA_CLIENT_ID=your_client_id
   TUYA_ACCESS_SECRET=your_secret
   TUYA_DEVICE_ID=your_device_id
   ```

2. Make script executable:
   ```bash
   chmod +x unlock.sh
   ```

3. Run:
   ```bash
   ./unlock.sh true
   ```

---

**Device:** Smart Doorlock (Lite Version)  
**Category:** `ms`  
**Device ID:** `a328e47b067a5c5b6bc8k6`
