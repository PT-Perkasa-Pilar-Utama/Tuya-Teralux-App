package com.example.whisperandroid.domain.model.reminder

/**
 * UI model for displaying a meeting reminder in the overlay.
 *
 * @property title Title text for the reminder overlay
 * @property message Body message showing remaining time
 * @property remainingMinutes Minutes remaining until the meeting starts
 */
data class MeetingReminderUiModel(
    val title: String,
    val message: String,
    val remainingMinutes: Int
) {
    companion object {
        /**
         * Create a UI model from a reminder entity.
         */
        fun fromEntity(entity: MeetingReminderEntity): MeetingReminderUiModel {
            return MeetingReminderUiModel(
                title = "Meeting Reminder",
                message = "Waktu meeting anda sisa ${entity.remainingMinutes} menit lagi",
                remainingMinutes = entity.remainingMinutes
            )
        }
    }
}
