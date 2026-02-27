package com.sensio.notification

import android.app.NotificationChannel
import android.app.NotificationManager
import android.content.Context
import android.os.Build
import android.widget.Toast
import androidx.core.app.NotificationCompat

lateinit var appContext: Context

actual fun showNotification(
    title: String,
    message: String,
) {
    // Also show toast as it doesn't require runtime permission and gives immediate feedback
    Toast.makeText(appContext, "$title: $message", Toast.LENGTH_LONG).show()

    val notificationManager = appContext.getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager
    val channelId = "sensio_notification_channel"

    if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
        val channel =
            NotificationChannel(
                channelId,
                "Sensio Notifications",
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
