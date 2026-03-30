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
        val publishAt = 1710655500000L
        val remainingMinutes = 15

        val id1 = MeetingReminderEntity.generateId(publishAt, remainingMinutes)
        val id2 = MeetingReminderEntity.generateId(publishAt, remainingMinutes)

        assertEquals(id1, id2)
    }

    @Test
    fun generateId_differentPublishAt_returnsDifferentId() {
        val publishAt1 = 1710655500000L
        val publishAt2 = 1710655800000L
        val remainingMinutes = 15

        val id1 = MeetingReminderEntity.generateId(publishAt1, remainingMinutes)
        val id2 = MeetingReminderEntity.generateId(publishAt2, remainingMinutes)

        assertNotEquals(id1, id2)
    }

    @Test
    fun generateId_differentRemainingMinutes_returnsDifferentId() {
        val publishAt = 1710655500000L
        val remainingMinutes1 = 15
        val remainingMinutes2 = 30

        val id1 = MeetingReminderEntity.generateId(publishAt, remainingMinutes1)
        val id2 = MeetingReminderEntity.generateId(publishAt, remainingMinutes2)

        assertNotEquals(id1, id2)
    }

    @Test
    fun entity_withDefaultFiredValue_isFalse() {
        val entity = MeetingReminderEntity(
            id = "test_id",
            publishAtEpochMillis = 1710655500000L,
            remainingMinutes = 15,
            createdAtEpochMillis = System.currentTimeMillis()
        )

        assertEquals(false, entity.fired)
    }

    @Test
    fun uiModel_fromEntity_correctMessageFormat() {
        val entity = MeetingReminderEntity(
            id = "test_id",
            publishAtEpochMillis = 1710655500000L,
            remainingMinutes = 15,
            createdAtEpochMillis = System.currentTimeMillis()
        )

        val uiModel = MeetingReminderUiModel.fromEntity(entity)

        assertEquals("Meeting Reminder", uiModel.title)
        assertEquals("Waktu meeting anda sisa 15 menit lagi", uiModel.message)
        assertEquals(15, uiModel.remainingMinutes)
    }
}
