# Smart Door Lock Test Results

## Date: April 17, 2026

## Device Information

### Current Device
- **Name:** Smart Doorlock (Lite Version)
- **Device ID:** `a328e47b067a5c5b6bc8k6`
- **Category:** `ms`
- **Status:** ✅ Online
- **Model:** DL-WH-HV
- **Product:** Smart Doorlock (Lite Version)

### Available Functions
Based on device specifications, this device supports:

**Status Monitoring:**
- ✅ `unlock_fingerprint` - Track fingerprint unlocks
- ✅ `unlock_password` - Track password unlocks
- ✅ `unlock_temporary` - Track temporary password unlocks
- ✅ `unlock_dynamic` - Track dynamic password unlocks
- ✅ `unlock_card` - Track card unlocks
- ✅ `unlock_face` - Track face unlocks
- ✅ `unlock_key` - Track key unlocks
- ✅ `unlock_request` - Remote unlock request (0-90 seconds)
- ✅ `residual_electricity` - Battery level
- ✅ `unlock_app` - Track app unlocks
- ✅ `hijack` - Duress alarm
- ✅ `open_inside` - Inside open status
- ✅ `closed_opened` - Door open/close status
- ✅ `doorbell` - Doorbell status
- ✅ `anti_lock_outside` - Anti-lock status

**Commands:**
- ✅ `reply_unlock_request` - Reply to unlock request

**NOT Supported:**
- ❌ Create temporary password via API
- ❌ Create dynamic password via API
- ❌ Password ticket generation

## API Test Results

| Test | Status | Response |
|------|--------|----------|
| Get Access Token | ✅ PASS | Token obtained successfully |
| Get Device Info | ✅ PASS | Device found and online |
| Get Device Specifications | ✅ PASS | Category: `ms` |
| Create Temp Password (v1) | ❌ FAIL | `param is illegal (code: 1109)` |
| Create Temp Password (v2 with ticket) | ❌ FAIL | API not supported for this device |
| Create Dynamic Password | ❌ FAIL | API not supported for this device |

## Root Cause

The **Smart Doorlock (Lite Version)** with category `ms` does NOT support the Door Lock Password API endpoints:
- `/v1.0/devices/{id}/door-lock/temp-password`
- `/v1.0/devices/{id}/door-lock/dynamic-password`
- `/v1.0/devices/{id}/door-lock/password-ticket`

These APIs are only available for:
- Category: `jtmspro` (Smart Door Lock Pro)
- Category: `dlq` with specific product models that support password management

## Recommendations

### Option 1: Use Smart Life App (Manual) ✅ RECOMMENDED
Continue using the Smart Life / Tuya app to create temporary passwords manually:
1. Open Smart Life app
2. Select your door lock
3. Go to "Password" or "Temporary Password" section
4. Create temporary password with desired duration

### Option 2: Upgrade to Pro Device
If API automation is required, consider upgrading to a door lock that supports the full API:
- Look for devices with category `jtmspro`
- Example: U688S-WiFi-Pro (mentioned in old README)

### Option 3: Local WiFi API (Advanced)
Some door locks support local WiFi API calls without going through Tuya cloud. This would require:
1. Device must be on same WiFi network
2. Use device's local IP and port
3. Encrypt commands with `local_key`
4. Reverse engineer the Smart Life app protocol

## Current Device Status

```
Device: Smart Doorlock (Lite Version)
ID: a328e47b067a5c5b6bc8k6
Online: YES
Battery: Available
Last Activity: Recent
```

## Alternative Automation

If you need automation, you can:
1. **Monitor unlock events** - Track when passwords/fingerprints are used
2. **Remote unlock** - Use `unlock_request` command (if supported)
3. **Get battery status** - Monitor `residual_electricity`
4. **Get door status** - Monitor `closed_opened`

## Files Updated

- `/home/farismnrr/Documents/shared/teralux_app/backend/.env` - Updated with correct device ID
- `/home/farismnrr/Documents/shared/teralux_app/backend/services/smart-door-lock-test/.env` - Updated with correct device ID

## Test Scripts Created

- `test_apis.py` - Comprehensive API testing
- `test_device_specs.py` - Device specification checker
- `create_temp_password_simple.py` - Simple password creation (for compatible devices)
- `create_temp_password_5min.py` - Password creation with ticket encryption

---

**Summary:** Your current door lock device is a "Lite Version" that doesn't support temporary password creation via Tuya cloud API. You'll need to use the Smart Life app manually or upgrade to a Pro version device for API automation.
