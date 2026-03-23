package com.example.whisperandroid.service.reminder

import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.os.Build
import androidx.core.app.NotificationCompat
import com.example.whisperandroid.MainActivity
import com.example.whisperandroid.domain.model.reminder.MeetingReminderUiModel

/**
 * Notifier for meeting reminders.
 *
 * Creates notification channel and posts high-priority notifications.
 */
class MeetingReminderNotifier(private val context: Context) {
    private val notificationManager: NotificationManager =
        context.getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager

    init {
        createNotificationChannel()
    }

    /**
     * Create the notification channel for meeting reminders.
     */
    private fun createNotificationChannel() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val channel = NotificationChannel(
                CHANNEL_ID,
                "Meeting Reminders",
                NotificationManager.IMPORTANCE_HIGH
            ).apply {
                description = "Notifications for upcoming meeting reminders"
                enableVibration(true)
                lockscreenVisibility = NotificationCompat.VISIBILITY_PUBLIC
            }
            notificationManager.createNotificationChannel(channel)
        }
    }

    /**
     * Show a meeting reminder notification.
     *
     * @param uiModel The UI model containing reminder content
     */
    fun showNotification(uiModel: MeetingReminderUiModel) {
        val notificationId = uiModel.remainingMinutes.hashCode() + NOTIFICATION_ID_OFFSET

        val notification = NotificationCompat.Builder(context, CHANNEL_ID)
            .setContentTitle(uiModel.title)
            .setContentText(uiModel.message)
            .setSmallIcon(android.R.drawable.ic_dialog_alert)
            .setPriority(NotificationCompat.PRIORITY_HIGH)
            .setCategory(NotificationCompat.CATEGORY_REMINDER)
            .setAutoCancel(true)
            .setContentIntent(createContentIntent())
            .build()

        notificationManager.notify(notificationId, notification)
    }

    /**
     * Create a pending intent to open MainActivity when notification is tapped.
     */
    private fun createContentIntent(): PendingIntent {
        val intent = Intent(context, MainActivity::class.java).apply {
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK
        }

        return PendingIntent.getActivity(
            context,
            0,
            intent,
            PendingIntent.FLAG_IMMUTABLE
        )
    }

    companion object {
        const val CHANNEL_ID = "meeting_reminder_channel"
        private const val NOTIFICATION_ID_OFFSET = 2000
    }
}
