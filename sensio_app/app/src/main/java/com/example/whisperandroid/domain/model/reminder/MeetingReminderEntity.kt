package com.example.whisperandroid.domain.model.reminder

/**
 * Data class representing a persisted meeting reminder entity.
 *
 * @property id Unique identifier for deduplication (publishAt + remainingMinutes)
 * @property publishAtEpochMillis Epoch timestamp when the reminder should fire
 * @property remainingMinutes Minutes remaining until the meeting starts
 * @property createdAtEpochMillis Epoch timestamp when the reminder was created
 * @property fired Whether the reminder has been fired
 */
data class MeetingReminderEntity(
    val id: String,
    val publishAtEpochMillis: Long,
    val remainingMinutes: Int,
    val createdAtEpochMillis: Long,
    val fired: Boolean = false
) {
    companion object {
        /**
         * Generate a dedupe key from publishAt and remainingMinutes.
         */
        fun generateId(publishAtEpochMillis: Long, remainingMinutes: Int): String {
            return "${publishAtEpochMillis}_$remainingMinutes"
        }
    }
}
