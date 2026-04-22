# Notification Publish API

## Overview

Publishes a notification to all terminals in a room via MQTT, with optional WhatsApp reminders. Notifications are scheduled via `scheduled_at` (ISO 8601 timestamp) and can include WhatsApp reminders to the meeting organizer.

### Endpoint

```
POST /api/notification/publish
```

### Headers

| Header | Value |
|--------|-------|
| Content-Type | application/json |
| Authorization | Bearer {token} |

### Request Body

```json
{
  "room_id": "room-001",
  "scheduled_at": "2026-04-20T23:00:00+07:00",
  "phone_numbers": ["+6281234567890"],
  "template": "end_meeting"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| room_id | string | Yes | The room ID to publish notification to |
| scheduled_at | string | No* | RFC3339 timestamp (e.g., `2026-04-20T23:00:00+07:00`). When to trigger the MQTT publish and WA reminder. |
| phone_numbers | array[string] | Yes | Non-empty array of phone numbers to send WhatsApp reminder |
| template | string | No | Notification template key. Defaults to `end_meeting` if omitted |

*If `scheduled_at` is omitted, the system attempts to derive the scheduled time from the device info booking end time. If that derivation also fails, MQTT is **not** published and a 400 error is returned.

### Response (Success - 200)

```json
{
  "status": true,
  "message": "Notification published successfully",
  "data": {
    "room_id": "room-001",
    "publish_at": "2026-04-20T22:45:00+07:00",
    "published_count": 1,
    "published_topics": [
      "users/db329671-96bb-368b-95d3-53a3a3712563/PRODUCTION/notification"
    ],
    "wa_notification_id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| status | boolean | Always `true` on success |
| message | string | Human-readable status message |
| data.room_id | string | The room ID |
| data.publish_at | datetime | When the notification will be triggered |
| data.published_count | integer | Number of terminals that received the MQTT message |
| data.published_topics | array[string] | MQTT topics the message was published to |
| data.wa_notification_id | string | UUID of the scheduled WA notification (only if phone_numbers was provided) |

### Response (Error - 400)

```json
{
  "status": false,
  "message": "room_id is required"
}
```

```json
{
  "status": false,
  "message": "phone_numbers must be a non-empty array"
}
```

```json
{
  "status": false,
  "message": "scheduled_at must be a valid RFC3339 timestamp"
}
```

```json
{
  "status": false,
  "message": "could not determine notification time: no scheduled_at provided and no booking end time available"
}
```

### Response (Error - 404)

```json
{
  "status": false,
  "message": "No terminals found for RoomID room-001"
}
```

## WhatsApp Notification Flow

1. When `phone_numbers` is provided in the request:
   - System looks up MAC address from terminal(s) in the room
   - Fetches booking information from Big API (`aplikasi-big.com`)
   - Schedules a WA job in the database with `status: pending`

2. Background Worker:
   - Polls database every 30 seconds
   - Executes WA send when `scheduled_at <= now` and `status = pending`
   - Updates status to `sent` on success, or `failed` on error

### WhatsApp Message Format

```
[PENGINGAT JADWAL PERTEMUAN]

Yth. {customer_name},

Melalui pesan ini, kami ingin mengingatkan bahwa jadwal pertemuan Anda akan dimulai dalam {remaining_minutes} menit. Berikut adalah rincian pertemuan tersebut:

🏢 Lokasi: {building_name}
🚪 Ruang: {room_name}
📅 Tanggal: {date}
⏰ Waktu: {booking_time}
🔐 Kata Sandi (Password): {password}

Kami mohon kesediaan {customer_name} untuk bersiap sebelum waktu pertemuan dimulai. Terima kasih atas perhatian dan kerja samanya.

Salam hangat,
[Nama Perusahaan/Tim Anda]
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| WA_NOTIFICATION_BASE_URL | `http://10.10.3.24:3000/api/v1/send` | WhatsApp API endpoint |