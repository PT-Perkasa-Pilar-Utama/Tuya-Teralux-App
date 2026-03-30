package com.example.whisperandroid.service.reminder

import android.app.AlarmManager
import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.os.Build
import android.util.Log
import com.example.whisperandroid.domain.model.reminder.MeetingReminderEntity

/**
 * Scheduler for meeting reminder alarms using AlarmManager.
 *
 * Handles exact-time scheduling for reminder delivery.
 */
class MeetingReminderScheduler(private val context: Context) {
    private val alarmManager: AlarmManager = context.getSystemService(Context.ALARM_SERVICE) as AlarmManager
    private val tag = "MeetingReminderScheduler"

    /**
     * Schedule an exact alarm for a reminder.
     *
     * @param entity The reminder entity to schedule
     */
    fun scheduleReminder(entity: MeetingReminderEntity) {
        val currentTime = System.currentTimeMillis()

        // Check if reminder is in the past but within grace window
        val fireTime = entity.publishAtEpochMillis
        if (fireTime < currentTime) {
            if (currentTime - fireTime <= GRACE_WINDOW_MILLIS) {
                Log.i(tag, "Reminder ${entity.id} is late but within grace window, firing immediately")
                triggerImmediate(entity)
                return
            } else {
                Log.w(tag, "Reminder ${entity.id} is stale (${(currentTime - fireTime) / 1000}s late), discarding")
                return
            }
        }

        val pendingIntent = createAlarmPendingIntent(entity.id)

        try {
            // Check exact alarm permission on Android 12+ (SDK 31+)
            val canScheduleExact = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
                alarmManager.canScheduleExactAlarms()
            } else {
                true
            }

            if (canScheduleExact) {
                // Use setExactAndAllowWhileIdle on API 23+
                if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
                    alarmManager.setExactAndAllowWhileIdle(
                        AlarmManager.RTC_WAKEUP,
                        fireTime,
                        pendingIntent
                    )
                } else if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.KITKAT) {
                    alarmManager.setExact(
                        AlarmManager.RTC_WAKEUP,
                        fireTime,
                        pendingIntent
                    )
                } else {
                    @Suppress("DEPRECATION")
                    alarmManager.set(
                        AlarmManager.RTC_WAKEUP,
                        fireTime,
                        pendingIntent
                    )
                }
            } else {
                // Fallback to inexact alarm if exact alarm permission not granted
                Log.w(tag, "Exact alarms not permitted, using setAndAllowWhileIdle fallback for ${entity.id}")
                if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
                    alarmManager.setAndAllowWhileIdle(
                        AlarmManager.RTC_WAKEUP,
                        fireTime,
                        pendingIntent
                    )
                } else {
                    alarmManager.set(
                        AlarmManager.RTC_WAKEUP,
                        fireTime,
                        pendingIntent
                    )
                }
            }
            Log.i(tag, "Scheduled reminder ${entity.id} for ${entity.publishAtEpochMillis}")
        } catch (e: SecurityException) {
            Log.e(tag, "SecurityException scheduling ${entity.id}: ${e.message}, falling back to non-exact alarm")
            // Fallback to non-exact alarm as last resort
            try {
                alarmManager.set(AlarmManager.RTC_WAKEUP, fireTime, pendingIntent)
                Log.i(tag, "Fallback non-exact alarm scheduled for ${entity.id}")
            } catch (e2: Exception) {
                Log.e(tag, "Failed to schedule fallback alarm for ${entity.id}: ${e2.message}")
            }
        } catch (e: Exception) {
            Log.e(tag, "Failed to schedule reminder ${entity.id}: ${e.message}")
        }
    }

    /**
     * Cancel a scheduled alarm.
     *
     * @param reminderId The reminder ID to cancel
     */
    fun cancelReminder(reminderId: String) {
        val pendingIntent = createAlarmPendingIntent(reminderId)
        alarmManager.cancel(pendingIntent)
        Log.d(tag, "Cancelled reminder alarm: $reminderId")
    }

    /**
     * Reschedule all pending reminders.
     *
     * Called on app startup or boot to restore alarms.
     */
    fun rescheduleAll(entities: List<MeetingReminderEntity>) {
        val currentTime = System.currentTimeMillis()
        var scheduled = 0
        var skipped = 0

        entities.forEach { entity ->
            if (!entity.fired) {
                val fireTime = entity.publishAtEpochMillis
                if (fireTime + GRACE_WINDOW_MILLIS >= currentTime) {
                    scheduleReminder(entity)
                    scheduled++
                } else {
                    skipped++
                }
            }
        }

        Log.i(tag, "Rescheduled $scheduled reminders, skipped $skipped stale ones")
    }

    /**
     * Trigger an immediate alarm (for late reminders within grace window).
     */
    private fun triggerImmediate(entity: MeetingReminderEntity) {
        val intent = Intent(context, MeetingReminderAlarmReceiver::class.java).apply {
            action = ACTION_IMMEDIATE_TRIGGER
            putExtra(EXTRA_REMINDER_ID, entity.id)
            putExtra(EXTRA_PUBLISH_AT, entity.publishAtEpochMillis)
            putExtra(EXTRA_REMAINING_MINUTES, entity.remainingMinutes)
        }

        val pendingIntent = PendingIntent.getBroadcast(
            context,
            entity.id.hashCode(),
            intent,
            PendingIntent.FLAG_ONE_SHOT or PendingIntent.FLAG_IMMUTABLE
        )

        try {
            val canScheduleExact = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
                alarmManager.canScheduleExactAlarms()
            } else {
                true
            }

            if (canScheduleExact) {
                if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
                    alarmManager.setExactAndAllowWhileIdle(
                        AlarmManager.RTC_WAKEUP,
                        System.currentTimeMillis(),
                        pendingIntent
                    )
                } else if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.KITKAT) {
                    alarmManager.setExact(
                        AlarmManager.RTC_WAKEUP,
                        System.currentTimeMillis(),
                        pendingIntent
                    )
                } else {
                    @Suppress("DEPRECATION")
                    alarmManager.set(
                        AlarmManager.RTC_WAKEUP,
                        System.currentTimeMillis(),
                        pendingIntent
                    )
                }
            } else {
                Log.w(tag, "Exact alarms not permitted for immediate trigger, using setAndAllowWhileIdle fallback for ${entity.id}")
                if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
                    alarmManager.setAndAllowWhileIdle(
                        AlarmManager.RTC_WAKEUP,
                        System.currentTimeMillis(),
                        pendingIntent
                    )
                } else {
                    alarmManager.set(
                        AlarmManager.RTC_WAKEUP,
                        System.currentTimeMillis(),
                        pendingIntent
                    )
                }
            }
        } catch (e: SecurityException) {
            Log.e(tag, "SecurityException calling triggerImmediate for ${entity.id}: ${e.message}, falling back to non-exact alarm")
            try {
                alarmManager.set(AlarmManager.RTC_WAKEUP, System.currentTimeMillis(), pendingIntent)
            } catch (e2: Exception) {
                Log.e(tag, "Failed to schedule fallback immediate alarm for ${entity.id}: ${e2.message}")
            }
        } catch (e: Exception) {
            Log.e(tag, "Failed to schedule immediate reminder ${entity.id}: ${e.message}")
        }
    }

    private fun createAlarmPendingIntent(reminderId: String): PendingIntent {
        val intent = Intent(context, MeetingReminderAlarmReceiver::class.java).apply {
            putExtra(EXTRA_REMINDER_ID, reminderId)
        }

        return PendingIntent.getBroadcast(
            context,
            reminderId.hashCode(),
            intent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )
    }

    companion object {
        const val ACTION_IMMEDIATE_TRIGGER = "com.example.whisperandroid.action.IMMEDIATE_TRIGGER"
        const val EXTRA_REMINDER_ID = "reminder_id"
        const val EXTRA_PUBLISH_AT = "publish_at"
        const val EXTRA_REMAINING_MINUTES = "remaining_minutes"

        private val GRACE_WINDOW_MILLIS = 2 * 60 * 1000L // 2 minutes
    }
}
