package com.sensio.notification

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.Service
import android.content.Context
import android.content.Intent
import android.os.Build
import android.os.IBinder
import androidx.core.app.NotificationCompat
import com.sensio.notification.logic.MeetingMonitor
import com.sensio.notification.model.MeetingSession
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.delay
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import kotlinx.datetime.Clock
import kotlin.time.Duration.Companion.minutes
import kotlin.time.Duration.Companion.seconds

class MeetingMonitorService : Service() {
    private val serviceScope = CoroutineScope(Dispatchers.Default + SupervisorJob())
    private lateinit var monitor: MeetingMonitor

    override fun onCreate() {
        super.onCreate()
        appContext = applicationContext

        // Mock session: ends in 6 minutes
        val mockSession =
            MeetingSession(
                id = "1",
                title = "Project Sync",
                endTime = Clock.System.now() + 6.minutes,
            )
        monitor = MeetingMonitor(mockSession)

        startForeground(NOTIFICATION_ID, createForegroundNotification())

        serviceScope.launch {
            while (isActive) {
                monitor.checkAndTrigger()
                delay(10.seconds)
            }
        }
    }

    private fun createForegroundNotification(): Notification {
        val channelId = "meeting_monitor_channel"
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val channel =
                NotificationChannel(
                    channelId,
                    "Meeting Monitor",
                    NotificationManager.IMPORTANCE_LOW,
                )
            val manager = getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager
            manager.createNotificationChannel(channel)
        }

        return NotificationCompat.Builder(this, channelId)
            .setContentTitle("Sensio Meeting Monitor")
            .setContentText("Monitoring active meetings...")
            .setSmallIcon(android.R.drawable.ic_dialog_info)
            .build()
    }

    override fun onBind(intent: Intent?): IBinder? = null

    override fun onDestroy() {
        super.onDestroy()
        serviceScope.cancel()
    }

    companion object {
        const val NOTIFICATION_ID = 1001
    }
}
