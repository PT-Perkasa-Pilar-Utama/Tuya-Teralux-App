package com.example.whisperandroid.domain.model.reminder

/**
 * UI model for displaying a meeting reminder in the overlay.
 *
 * @property title Title text for the reminder overlay (from MQTT payload)
 * @property message Body message for the reminder (from MQTT payload)
 * @property eventType Event type: meeting_start or meeting_end
 */
data class MeetingReminderUiModel(
    val title: String,
    val message: String,
    val eventType: String
) {
    companion object {
        /**
         * Create a UI model from a reminder entity.
         * Uses dynamic title/message from MQTT payload.
         */
        fun fromEntity(entity: MeetingReminderEntity): MeetingReminderUiModel {
            return MeetingReminderUiModel(
                title = entity.title,
                message = entity.message,
                eventType = entity.eventType
            )
        }
    }
}
