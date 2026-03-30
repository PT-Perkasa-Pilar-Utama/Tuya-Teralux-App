package com.example.whisperandroid.service.reminder

import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.app.Service
import android.content.Context
import android.content.Intent
import android.os.Build
import android.os.IBinder
import androidx.core.app.NotificationCompat
import com.example.whisperandroid.MainActivity
import com.example.whisperandroid.data.di.NetworkModule
import com.example.whisperandroid.util.AppLog
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel

/**
 * Foreground service for meeting reminder runtime.
 *
 * Runs independently from background assistant mode to ensure reminders
 * are always received and processed.
 */
class MeetingReminderService : Service() {

    private val serviceScope = CoroutineScope(Dispatchers.Main + SupervisorJob())
    private val notificationId = 2001
    private val channelId = "meeting_reminder_service_channel"
    private var reminderCoordinator: MeetingReminderRuntimeCoordinator? = null
    private val TAG = "MeetingReminderService"

    companion object {
        const val ACTION_START_REMINDER = "ACTION_START_REMINDER"
        const val ACTION_STOP_REMINDER = "ACTION_STOP_REMINDER"
    }

    override fun onBind(intent: Intent?): IBinder? = null

    override fun onCreate() {
        super.onCreate()
        AppLog.i(TAG, "Service onCreate")

        // Ensure NetworkModule is initialized
        NetworkModule.ensureInitialized(applicationContext)

        createNotificationChannel()
        startForeground(notificationId, createNotification())

        // Initialize and start reminder coordinator
        try {
            reminderCoordinator = NetworkModule.meetingReminderRuntimeCoordinator
            reminderCoordinator?.start(serviceScope)
            AppLog.i(TAG, "Meeting reminder service started")
        } catch (e: Exception) {
            AppLog.e(TAG, "Failed to initialize reminder coordinator", e)
        }
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        val action = intent?.action
        AppLog.d(TAG, "onStartCommand action: $action")

        when (action) {
            ACTION_START_REMINDER -> {
                // Already started in onCreate
                AppLog.d(TAG, "Service already running")
            }
            ACTION_STOP_REMINDER -> {
                performCleanup()
                stopForeground(true)
                stopSelf()
            }
        }

        return START_STICKY
    }

    private fun createNotificationChannel() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val notificationManager = getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager

            val channel = NotificationChannel(
                channelId,
                "Meeting Reminder Service",
                NotificationManager.IMPORTANCE_LOW
            ).apply {
                description = "Ensures meeting reminders are always received"
            }
            notificationManager.createNotificationChannel(channel)
        }
    }

    private fun createNotification() = NotificationCompat.Builder(this, channelId)
        .setContentTitle("Meeting Reminders Active")
        .setContentText("Listening for meeting reminders...")
        .setSmallIcon(android.R.drawable.ic_dialog_info)
        .setPriority(NotificationCompat.PRIORITY_LOW)
        .setOngoing(true)
        .setContentIntent(
            PendingIntent.getActivity(
                this,
                0,
                Intent(this, MainActivity::class.java),
                PendingIntent.FLAG_IMMUTABLE
            )
        )
        .build()

    private fun performCleanup() {
        AppLog.i(TAG, "Performing cleanup")
        reminderCoordinator?.stop()
        serviceScope.cancel()
    }

    override fun onDestroy() {
        super.onDestroy()
        AppLog.i(TAG, "Service onDestroy")
        performCleanup()
    }
}
