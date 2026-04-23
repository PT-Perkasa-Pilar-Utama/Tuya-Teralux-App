package com.example.whisperandroid.domain.model.reminder

/**
 * Data class representing a persisted meeting reminder entity.
 *
 * @property id Unique identifier from the backend payload
 * @property publishAtEpochMillis Epoch timestamp when the reminder should fire
 * @property title Title text for the reminder overlay (from MQTT payload)
 * @property message Body message for the reminder (from MQTT payload)
 * @property eventType Event type: meeting_start or meeting_end
 * @property meetingId Optional meeting identifier
 * @property roomId Optional room identifier
 * @property severity Optional severity level
 * @property createdAtEpochMillis Epoch timestamp when the reminder was created
 * @property fired Whether the reminder has been fired
 */
data class MeetingReminderEntity(
    val id: String,
    val publishAtEpochMillis: Long,
    val title: String,
    val message: String,
    val eventType: String,
    val meetingId: String? = null,
    val roomId: String? = null,
    val severity: String? = null,
    val createdAtEpochMillis: Long,
    val fired: Boolean = false
) {
    companion object {
        /**
         * Generate a dedupe key from id (backend-provided).
         */
        fun generateId(backendId: String): String {
            return backendId
        }
    }
}
