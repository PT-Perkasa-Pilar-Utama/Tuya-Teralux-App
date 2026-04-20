package com.example.whisperandroid.util

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

/**
 * Unit tests for MeetingReminderParser.
 * Tests the new payload contract: id, publish_at, title, message, event_type
 */
class MeetingReminderParserTest {

    @Test
    fun parseValidPayload_returnsMessage() {
        val payload = """{"id": "rem-123", "publish_at": "2026-03-17T13:45:00+0700", "title": "Meeting Reminder", "message": "Your meeting will start soon", "event_type": "meeting_start"}"""

        val result = MeetingReminderParser.parse(payload)

        assertEquals("rem-123", result?.id)
        assertEquals("2026-03-17T13:45:00+0700", result?.publishAt)
        assertEquals("Meeting Reminder", result?.title)
        assertEquals("Your meeting will start soon", result?.message)
        assertEquals("meeting_start", result?.eventType)
    }

    @Test
    fun parseValidPayloadWithOptionalFields_returnsMessage() {
        val payload = """{"id": "rem-456", "publish_at": "2026-03-17T13:45:00+07:00", "title": "Meeting Ending", "message": "Your meeting is ending", "event_type": "meeting_end", "meeting_id": "meet-123", "room_id": "room-456", "severity": "high", "ttl_seconds": 3600}"""

        val result = MeetingReminderParser.parse(payload)

        assertEquals("rem-456", result?.id)
        assertEquals("2026-03-17T13:45:00+07:00", result?.publishAt)
        assertEquals("Meeting Ending", result?.title)
        assertEquals("Your meeting is ending", result?.message)
        assertEquals("meeting_end", result?.eventType)
        assertEquals("meet-123", result?.meetingId)
        assertEquals("room-456", result?.roomId)
        assertEquals("high", result?.severity)
        assertEquals(3600, result?.ttlSeconds)
    }

    @Test
    fun parseMalformedJson_returnsNull() {
        val payload = """{invalid json}"""

        val result = MeetingReminderParser.parse(payload)

        assertNull(result)
    }

    @Test
    fun parseMissingId_returnsNull() {
        val payload = """{"publish_at": "2026-03-17T13:45:00+0700", "title": "Meeting Reminder", "message": "Your meeting", "event_type": "meeting_start"}"""

        val result = MeetingReminderParser.parse(payload)

        assertNull(result)
    }

    @Test
    fun parseEmptyId_returnsNull() {
        val payload = """{"id": "", "publish_at": "2026-03-17T13:45:00+0700", "title": "Meeting Reminder", "message": "Your meeting", "event_type": "meeting_start"}"""

        val result = MeetingReminderParser.parse(payload)

        assertNull(result)
    }

    @Test
    fun parseMissingPublishAt_returnsNull() {
        val payload = """{"id": "rem-123", "title": "Meeting Reminder", "message": "Your meeting", "event_type": "meeting_start"}"""

        val result = MeetingReminderParser.parse(payload)

        assertNull(result)
    }

    @Test
    fun parseEmptyPublishAt_returnsNull() {
        val payload = """{"id": "rem-123", "publish_at": "", "title": "Meeting Reminder", "message": "Your meeting", "event_type": "meeting_start"}"""

        val result = MeetingReminderParser.parse(payload)

        assertNull(result)
    }

    @Test
    fun parseMissingTitle_returnsNull() {
        val payload = """{"id": "rem-123", "publish_at": "2026-03-17T13:45:00+0700", "message": "Your meeting", "event_type": "meeting_start"}"""

        val result = MeetingReminderParser.parse(payload)

        assertNull(result)
    }

    @Test
    fun parseEmptyTitle_returnsNull() {
        val payload = """{"id": "rem-123", "publish_at": "2026-03-17T13:45:00+0700", "title": "", "message": "Your meeting", "event_type": "meeting_start"}"""

        val result = MeetingReminderParser.parse(payload)

        assertNull(result)
    }

    @Test
    fun parseMissingMessage_returnsNull() {
        val payload = """{"id": "rem-123", "publish_at": "2026-03-17T13:45:00+0700", "title": "Meeting Reminder", "event_type": "meeting_start"}"""

        val result = MeetingReminderParser.parse(payload)

        assertNull(result)
    }

    @Test
    fun parseEmptyMessage_returnsNull() {
        val payload = """{"id": "rem-123", "publish_at": "2026-03-17T13:45:00+0700", "title": "Meeting Reminder", "message": "", "event_type": "meeting_start"}"""

        val result = MeetingReminderParser.parse(payload)

        assertNull(result)
    }

    @Test
    fun parseMissingEventType_returnsNull() {
        val payload = """{"id": "rem-123", "publish_at": "2026-03-17T13:45:00+0700", "title": "Meeting Reminder", "message": "Your meeting"}"""

        val result = MeetingReminderParser.parse(payload)

        assertNull(result)
    }

    @Test
    fun parseEmptyEventType_returnsNull() {
        val payload = """{"id": "rem-123", "publish_at": "2026-03-17T13:45:00+0700", "title": "Meeting Reminder", "message": "Your meeting", "event_type": ""}"""

        val result = MeetingReminderParser.parse(payload)

        assertNull(result)
    }

    @Test
    fun parseInvalidTimestampFormat_returnsNull() {
        val payload = """{"id": "rem-123", "publish_at": "invalid-date", "title": "Meeting Reminder", "message": "Your meeting", "event_type": "meeting_start"}"""

        val result = MeetingReminderParser.parse(payload)

        assertNull(result)
    }

    @Test
    fun parseTimestamp_validIso8601_returnsEpochMillis() {
        val timestamp = "2026-03-17T13:45:00+0700"

        val result = MeetingReminderParser.parseTimestamp(timestamp)

        assert(result != null)
        assert(result!! > 1700000000000L)
        assert(result < 1800000000000L)
    }

    @Test
    fun parseTimestamp_invalidFormat_returnsNull() {
        val result = MeetingReminderParser.parseTimestamp("not-a-date")

        assertNull(result)
    }
}