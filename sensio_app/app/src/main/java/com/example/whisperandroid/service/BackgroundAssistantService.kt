package com.example.whisperandroid.service

import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.app.Service
import android.content.Intent
import android.os.Build
import android.os.IBinder
import androidx.core.app.NotificationCompat
import com.example.whisperandroid.MainActivity
import com.example.whisperandroid.data.di.NetworkModule
import com.example.whisperandroid.presentation.assistant.SensioWakeWordManager
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.flow.collectLatest
import kotlinx.coroutines.launch

class BackgroundAssistantService : Service() {

    private val serviceScope = CoroutineScope(Dispatchers.Main + SupervisorJob())
    private val notificationId = 1002
    private val channelId = "background_assistant_channel"
    private val alertChannelId = "background_assistant_alerts"
    private var wakeWordManager: SensioWakeWordManager? = null
    private lateinit var coordinator: com.example.whisperandroid.presentation.assistant.BackgroundAssistantCoordinator
    private var overlayController: BackgroundAssistantOverlayController? = null
    private val fallbackNotificationId = 1003
    private var hasShownAlertInCurrentCycle = false
    override fun onBind(intent: Intent?): IBinder? = null

    override fun onCreate() {
        super.onCreate()
        android.util.Log.d("SensioService", "Service started but immediately stopping for foreground-only phase")
        stopSelf()
    }

    /* 
    private fun performCleanup() {
        wakeWordManager?.destroy()
        // ... rest of the code ...
    }
    */
}
