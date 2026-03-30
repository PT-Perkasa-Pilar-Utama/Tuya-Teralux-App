package com.example.whisperandroid.service.reminder

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.os.Build
import android.util.Log
import com.example.whisperandroid.data.di.NetworkModule

/**
 * Broadcast receiver for restoring meeting reminders on device boot.
 *
 * Reschedules all pending reminders and restarts the reminder service after BOOT_COMPLETED.
 */
class MeetingReminderBootReceiver : BroadcastReceiver() {
    private val tag = "MeetingReminderBoot"

    override fun onReceive(context: Context, intent: Intent) {
        if (intent.action != Intent.ACTION_BOOT_COMPLETED && intent.action != "android.intent.action.QUICKBOOT_POWERON") {
            return
        }

        Log.i(tag, "Boot completed received, restoring reminders...")

        // Ensure NetworkModule is initialized
        NetworkModule.ensureInitialized(context)

        val store = NetworkModule.meetingReminderStore
        val scheduler = NetworkModule.meetingReminderScheduler

        // Prune stale reminders first
        store.pruneStale(System.currentTimeMillis())

        // Get all valid pending reminders and reschedule
        val pendingReminders = store.getValidPendingReminders(System.currentTimeMillis())

        if (pendingReminders.isNotEmpty()) {
            scheduler.rescheduleAll(pendingReminders)
            Log.i(tag, "Restored ${pendingReminders.size} reminders after boot")
        } else {
            Log.d(tag, "No pending reminders to restore")
        }

        // Start the reminder service to resume MQTT listening
        Log.i(tag, "Starting MeetingReminderService after boot")
        val serviceIntent = Intent(context, MeetingReminderService::class.java).apply {
            action = MeetingReminderService.ACTION_START_REMINDER
        }

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            context.startForegroundService(serviceIntent)
        } else {
            context.startService(serviceIntent)
        }
    }
}
