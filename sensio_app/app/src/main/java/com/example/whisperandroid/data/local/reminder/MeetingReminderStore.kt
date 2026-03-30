package com.example.whisperandroid.data.local.reminder

import android.content.Context
import android.content.SharedPreferences
import android.util.Log
import com.example.whisperandroid.domain.model.reminder.MeetingReminderEntity
import com.google.gson.Gson
import com.google.gson.reflect.TypeToken

/**
 * SharedPreferences-based storage for meeting reminders.
 *
 * Handles persistence, deduplication, and retrieval of pending reminders.
 */
class MeetingReminderStore(context: Context) {
    private val prefs: SharedPreferences = context.getSharedPreferences(
        PREFS_NAME,
        Context.MODE_PRIVATE
    )
    private val gson = Gson()

    private val tag = "MeetingReminderStore"

    /**
     * Save a pending reminder. If a reminder with the same ID exists, it will be replaced.
     */
    fun savePending(entity: MeetingReminderEntity) {
        val pending = getPendingReminders().toMutableList()

        // Remove existing with same ID (dedupe)
        val existingIndex = pending.indexOfFirst { it.id == entity.id }
        if (existingIndex >= 0) {
            Log.d(tag, "Replacing existing reminder: ${entity.id}")
            pending.removeAt(existingIndex)
        }

        pending.add(entity)
        savePendingList(pending)
        Log.i(tag, "Saved reminder: ${entity.id}, fireAt=${entity.publishAtEpochMillis}")
    }

    /**
     * Get all pending (unfired) reminders.
     */
    fun getPendingReminders(): List<MeetingReminderEntity> {
        val json = prefs.getString(KEY_PENDING_REMINDERS, null) ?: return emptyList()
        val type = object : TypeToken<List<MeetingReminderEntity>>() {}.type
        return try {
            gson.fromJson(json, type) ?: emptyList()
        } catch (e: Exception) {
            Log.e(tag, "Error parsing pending reminders: ${e.message}")
            emptyList()
        }
    }

    /**
     * Mark a reminder as fired.
     */
    fun markFired(id: String) {
        val pending = getPendingReminders().toMutableList()
        val reminder = pending.find { it.id == id }
        if (reminder != null) {
            pending.remove(reminder)
            pending.add(reminder.copy(fired = true))
            savePendingList(pending)
            Log.i(tag, "Marked reminder as fired: $id")
        }
    }

    /**
     * Prune fired and stale reminders.
     */
    fun pruneStale(currentTimeMillis: Long) {
        val pending = getPendingReminders()
        val valid = pending.filter { !it.fired && it.publishAtEpochMillis + GRACE_WINDOW_MILLIS >= currentTimeMillis }

        if (valid.size != pending.size) {
            savePendingList(valid)
            Log.i(tag, "Pruned ${pending.size - valid.size} stale reminders")
        }
    }

    /**
     * Get all valid pending reminders (not fired, not stale).
     */
    fun getValidPendingReminders(currentTimeMillis: Long): List<MeetingReminderEntity> {
        return getPendingReminders().filter { !it.fired && it.publishAtEpochMillis + GRACE_WINDOW_MILLIS >= currentTimeMillis }
    }

    /**
     * Clear all reminders (for testing or reset).
     */
    fun clearAll() {
        prefs.edit().remove(KEY_PENDING_REMINDERS).apply()
        Log.i(tag, "Cleared all reminders")
    }

    private fun savePendingList(list: List<MeetingReminderEntity>) {
        val json = gson.toJson(list)
        prefs.edit().putString(KEY_PENDING_REMINDERS, json).apply()
    }

    companion object {
        private const val PREFS_NAME = "meeting_reminder_prefs"
        private const val KEY_PENDING_REMINDERS = "pending_reminders"

        /**
         * Grace window: reminders within 2 minutes of publish_at will still fire.
         */
        val GRACE_WINDOW_MILLIS = 2 * 60 * 1000L // 2 minutes
    }
}
