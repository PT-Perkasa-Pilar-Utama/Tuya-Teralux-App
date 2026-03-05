package com.sensio.notification

import android.app.NotificationChannel
import android.app.NotificationManager
import android.content.Context
import android.content.Intent
import android.os.Build
import androidx.core.app.NotificationCompat

lateinit var appContext: Context

actual fun showNotification(
    title: String,
    message: String,
) {
    // Launch Modal Activity
    val intent =
        Intent(appContext, NotificationActivity::class.java).apply {
            putExtra("title", title)
            putExtra("message", message)
            addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
        }
    appContext.startActivity(intent)

    val notificationManager = appContext.getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager
    val channelId = "sensio_notification_channel"

    if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
        val channel =
            NotificationChannel(
                channelId,
                "Meeting Reminder",
                NotificationManager.IMPORTANCE_DEFAULT,
            )
        notificationManager.createNotificationChannel(channel)
    }

    val notification =
        NotificationCompat.Builder(appContext, channelId)
            .setSmallIcon(android.R.drawable.ic_dialog_info) // Fallback small icon
            .setContentTitle(title)
            .setContentText(message)
            .setPriority(NotificationCompat.PRIORITY_DEFAULT)
            .setAutoCancel(true)
            .build()

    notificationManager.notify(1, notification)
}
