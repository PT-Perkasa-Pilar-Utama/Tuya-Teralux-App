package com.example.whisperandroid.domain.model.reminder

import com.google.gson.annotations.SerializedName

/**
 * Data class representing an incoming MQTT meeting reminder message.
 *
 * @property publishAt ISO 8601 timestamp when the reminder should be shown
 * @property remainingMinutes Minutes remaining until the meeting starts
 */
data class MeetingReminderMessage(
    @SerializedName("publish_at")
    val publishAt: String,
    @SerializedName("remaining_minutes")
    val remainingMinutes: Int
)
