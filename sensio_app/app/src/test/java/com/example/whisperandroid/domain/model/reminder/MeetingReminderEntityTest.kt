package com.example.whisperandroid.domain.model.reminder

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotEquals
import org.junit.Test

/**
 * Unit tests for MeetingReminderEntity.
 */
class MeetingReminderEntityTest {

    @Test
    fun generateId_sameInputs_returnsSameId() {
        val backendId = "rem-123-abc"

        val id1 = MeetingReminderEntity.generateId(backendId)
        val id2 = MeetingReminderEntity.generateId(backendId)

        assertEquals(id1, id2)
    }

    @Test
    fun generateId_differentInputs_returnsDifferentId() {
        val id1 = MeetingReminderEntity.generateId("rem-123-abc")
        val id2 = MeetingReminderEntity.generateId("rem-456-def")

        assertNotEquals(id1, id2)
    }

    @Test
    fun entity_withDefaultFiredValue_isFalse() {
        val entity = MeetingReminderEntity(
            id = "test_id",
            publishAtEpochMillis = 1710655500000L,
            title = "Meeting Reminder",
            message = "Your meeting will start soon",
            eventType = "meeting_start",
            createdAtEpochMillis = System.currentTimeMillis()
        )

        assertEquals(false, entity.fired)
    }

    @Test
    fun entity_withOptionalFields() {
        val entity = MeetingReminderEntity(
            id = "test_id",
            publishAtEpochMillis = 1710655500000L,
            title = "Meeting Ending",
            message = "Your meeting is ending",
            eventType = "meeting_end",
            meetingId = "meet-123",
            roomId = "room-456",
            severity = "high",
            createdAtEpochMillis = System.currentTimeMillis()
        )

        assertEquals("meet-123", entity.meetingId)
        assertEquals("room-456", entity.roomId)
        assertEquals("high", entity.severity)
    }

    @Test
    fun uiModel_fromEntity_usesDynamicTitleMessage() {
        val entity = MeetingReminderEntity(
            id = "test_id",
            publishAtEpochMillis = 1710655500000L,
            title = "Custom Title",
            message = "Custom message from backend",
            eventType = "meeting_start",
            createdAtEpochMillis = System.currentTimeMillis()
        )

        val uiModel = MeetingReminderUiModel.fromEntity(entity)

        assertEquals("Custom Title", uiModel.title)
        assertEquals("Custom message from backend", uiModel.message)
        assertEquals("meeting_start", uiModel.eventType)
    }
}
