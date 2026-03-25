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
 */
object MeetingReminderParser {
    private val tag = "MeetingReminderParser"
    private val gson = Gson()

    // ISO 8601 date format parser
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
     * @return Parsed [MeetingReminderMessage] or null if invalid
     */
    fun parse(payload: String): MeetingReminderMessage? {
        return try {
            val message = gson.fromJson(payload, MeetingReminderMessage::class.java)

            // Validate required fields
            if (message.publishAt.isBlank()) {
                Log.w(tag, "Invalid payload: missing or empty publish_at")
                return null
            }

            if (message.remainingMinutes < 0) {
                Log.w(tag, "Invalid payload: remaining_minutes must be non-negative: ${message.remainingMinutes}")
                return null
            }

            // Validate publish_at format
            if (parseTimestamp(message.publishAt) == null) {
                Log.w(tag, "Invalid payload: cannot parse publish_at: ${message.publishAt}")
                return null
            }

            Log.d(tag, "Successfully parsed reminder: publishAt=${message.publishAt}, remainingMinutes=${message.remainingMinutes}")
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
            // Handle timezone offset in format like +07:00 or +0700
            val normalized = normalizeTimezone(isoString)
            val date: Date = iso8601Format.get()!!.parse(normalized)
            date.time
        } catch (e: ParseException) {
            Log.e(tag, "Failed to parse timestamp '$isoString': ${e.message}")
            null
        } catch (e: Exception) {
            Log.e(tag, "Unexpected error parsing timestamp '$isoString': ${e.message}")
            null
        }
    }

    /**
     * Normalize timezone offset from +07:00 to +0700 for SimpleDateFormat.
     * Also handles UTC Z suffix (converts Z to +0000).
     */
    private fun normalizeTimezone(isoString: String): String {
        // Handle UTC Z suffix: convert "2024-03-24T10:00:00Z" to "2024-03-24T10:00:00+0000"
        val withUtcOffset = isoString.replace(Regex("""Z$"""), "+0000")
        
        // If already in +0700 format, return as-is
        if (Regex("""\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[+-]\d{4}""").matches(withUtcOffset)) {
            return withUtcOffset
        }
        // Convert +07:00 to +0700
        return withUtcOffset.replace(Regex("""([+-]\d{2}):(\d{2})$"""), "$1$2")
    }
}
