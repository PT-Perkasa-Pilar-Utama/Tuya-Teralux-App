package com.example.whisperandroid.domain.model.reminder

import com.google.gson.annotations.SerializedName

/**
 * Data class representing an incoming MQTT meeting reminder message.
 *
 * New payload contract (single source of truth from backend publisher):
 * - Required fields: id, publish_at, title, message, event_type
 * - event_type allowed values: meeting_start, meeting_end (extensible string enum)
 * - Optional metadata fields: meeting_id, room_id, severity, ttl_seconds
 *
 * @property id Unique identifier for deduplication
 * @property publishAt ISO 8601 timestamp when the reminder should be shown
 * @property title Title text for the reminder overlay
 * @property message Body message for the reminder
 * @property eventType Event type: meeting_start or meeting_end
 * @property meetingId Optional meeting identifier
 * @property roomId Optional room identifier
 * @property severity Optional severity level
 * @property ttlSeconds Optional TTL in seconds
 */
data class MeetingReminderMessage(
    @SerializedName("id")
    val id: String,
    @SerializedName("publish_at")
    val publishAt: String,
    @SerializedName("title")
    val title: String,
    @SerializedName("message")
    val message: String,
    @SerializedName("event_type")
    val eventType: String,
    @SerializedName("meeting_id")
    val meetingId: String? = null,
    @SerializedName("room_id")
    val roomId: String? = null,
    @SerializedName("severity")
    val severity: String? = null,
    @SerializedName("ttl_seconds")
    val ttlSeconds: Int? = null
)