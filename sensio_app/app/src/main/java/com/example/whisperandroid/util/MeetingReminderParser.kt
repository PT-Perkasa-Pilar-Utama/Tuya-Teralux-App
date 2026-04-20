package com.example.whisperandroid.util

import android.util.Log
import com.example.whisperandroid.domain.model.reminder.MeetingReminderMessage
import com.google.gson.Gson
import com.google.gson.JsonSyntaxException
import java.text.ParseException
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale
import java.util.TimeZone

/**
 * Parser for meeting reminder MQTT payloads.
 *
 * Validates and parses incoming JSON messages into typed models.
 * Strict validation: rejects payloads with missing or blank title, message, event_type, or invalid timestamp.
 */
object MeetingReminderParser {
    private val tag = "MeetingReminderParser"
    private val gson = Gson()

    private val iso8601Format = object : ThreadLocal<SimpleDateFormat>() {
        override fun initialValue(): SimpleDateFormat {
            return SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ssZ", Locale.US).apply {
                timeZone = TimeZone.getTimeZone("UTC")
            }
        }
    }

    /**
     * Parse and validate an incoming MQTT reminder payload.
     *
     * Required fields: id, publish_at, title, message, event_type
     * event_type allowed values: meeting_start, meeting_end
     *
     * @return Parsed [MeetingReminderMessage] or null if invalid
     */
    fun parse(payload: String): MeetingReminderMessage? {
        return try {
            val message = gson.fromJson(payload, MeetingReminderMessage::class.java)

            if (message.id.isBlank()) {
                Log.w(tag, "Invalid payload: missing or empty id")
                return null
            }

            if (message.publishAt.isBlank()) {
                Log.w(tag, "Invalid payload: missing or empty publish_at")
                return null
            }

            if (message.title.isBlank()) {
                Log.w(tag, "Invalid payload: missing or empty title")
                return null
            }

            if (message.message.isBlank()) {
                Log.w(tag, "Invalid payload: missing or empty message")
                return null
            }

            if (message.eventType.isBlank()) {
                Log.w(tag, "Invalid payload: missing or empty event_type")
                return null
            }

            if (parseTimestamp(message.publishAt) == null) {
                Log.w(tag, "Invalid payload: cannot parse publish_at: ${message.publishAt}")
                return null
            }

            Log.d(tag, "Successfully parsed reminder: id=${message.id}, eventType=${message.eventType}, title=${message.title}")
            message
        } catch (e: JsonSyntaxException) {
            Log.e(tag, "Invalid JSON payload: ${e.message}")
            null
        } catch (e: Exception) {
            Log.e(tag, "Unexpected error parsing payload: ${e.message}")
            null
        }
    }

    /**
     * Parse ISO 8601 timestamp to epoch millis.
     *
     * @return Epoch millis or null if parsing fails
     */
    fun parseTimestamp(isoString: String): Long? {
        return try {
            val normalized = normalizeTimezone(isoString)
            val date = iso8601Format.get()!!.parse(normalized) ?: return null
            date.time
        } catch (e: ParseException) {
            Log.e(tag, "Failed to parse timestamp '$isoString': ${e.message}")
            null
        } catch (e: Exception) {
            Log.e(tag, "Failed to parse timestamp '$isoString': ${e.message}")
            null
        }
    }

    /**
     * Normalize timezone offset from +07:00 to +0700 for SimpleDateFormat.
     * Also handles UTC Z suffix (converts Z to +0000).
     */
    private fun normalizeTimezone(isoString: String): String {
        val withUtcOffset = isoString.replace(Regex("""Z$"""), "+0000")

        if (Regex("""\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[+-]\d{4}""").matches(withUtcOffset)) {
            return withUtcOffset
        }
        return withUtcOffset.replace(Regex("""([+-]\d{2}):(\d{2})$"""), "$1$2")
    }
}