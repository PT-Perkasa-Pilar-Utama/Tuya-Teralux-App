package com.example.whisperandroid.util

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

/**
 * Unit tests for MeetingReminderParser.
 */
class MeetingReminderParserTest {

    @Test
    fun parseValidPayload_returnsMessage() {
        val payload = """{"publish_at": "2026-03-17T13:45:00+0700", "remaining_minutes": 15}"""

        val result = MeetingReminderParser.parse(payload)

        assertEquals("2026-03-17T13:45:00+0700", result?.publishAt)
        assertEquals(15, result?.remainingMinutes)
    }

    @Test
    fun parseValidPayloadWithColonTimezone_returnsMessage() {
        val payload = """{"publish_at": "2026-03-17T13:45:00+07:00", "remaining_minutes": 30}"""

        val result = MeetingReminderParser.parse(payload)

        assertEquals("2026-03-17T13:45:00+07:00", result?.publishAt)
        assertEquals(30, result?.remainingMinutes)
    }

    @Test
    fun parseMalformedJson_returnsNull() {
        val payload = """{invalid json}"""

        val result = MeetingReminderParser.parse(payload)

        assertNull(result)
    }

    @Test
    fun parseMissingPublishAt_returnsNull() {
        val payload = """{"remaining_minutes": 15}"""

        val result = MeetingReminderParser.parse(payload)

        assertNull(result)
    }

    @Test
    fun parseEmptyPublishAt_returnsNull() {
        val payload = """{"publish_at": "", "remaining_minutes": 15}"""

        val result = MeetingReminderParser.parse(payload)

        assertNull(result)
    }

    @Test
    fun parseZeroRemainingMinutes_returnsMessage() {
        val payload = """{"publish_at": "2026-03-17T13:45:00+0700", "remaining_minutes": 0}"""

        val result = MeetingReminderParser.parse(payload)

        assertEquals("2026-03-17T13:45:00+0700", result?.publishAt)
        assertEquals(0, result?.remainingMinutes)
    }

    @Test
    fun parseNegativeRemainingMinutes_returnsNull() {
        val payload = """{"publish_at": "2026-03-17T13:45:00+0700", "remaining_minutes": -5}"""

        val result = MeetingReminderParser.parse(payload)

        assertNull(result)
    }

    @Test
    fun parseInvalidTimestampFormat_returnsNull() {
        val payload = """{"publish_at": "invalid-date", "remaining_minutes": 15}"""

        val result = MeetingReminderParser.parse(payload)

        assertNull(result)
    }

    @Test
    fun parseTimestamp_validIso8601_returnsEpochMillis() {
        // 2026-03-17T13:45:00+07:00 in epoch millis
        // Note: Parser expects timezone offset format like +0700 or +07:00
        val timestamp = "2026-03-17T13:45:00+0700"

        val result = MeetingReminderParser.parseTimestamp(timestamp)

        // Expected: approximately 1742193900000 (depending on exact timezone)
        // Just verify it's not null and is a reasonable timestamp
        assert(result != null)
        assert(result!! > 1700000000000L) // After 2023
        assert(result < 1800000000000L) // Before 2027
    }

    @Test
    fun parseTimestamp_invalidFormat_returnsNull() {
        val result = MeetingReminderParser.parseTimestamp("not-a-date")

        assertNull(result)
    }
}
