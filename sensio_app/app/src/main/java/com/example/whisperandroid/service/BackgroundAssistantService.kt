package com.example.whisperandroid.service

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
import com.example.whisperandroid.presentation.assistant.BackgroundAssistantCoordinator
import com.example.whisperandroid.presentation.assistant.SensioWakeWordManager
import com.example.whisperandroid.util.AppLog
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
    private var wakeWordManager: SensioWakeWordManager? = null
    private lateinit var coordinator: BackgroundAssistantCoordinator
    private var overlayController: BackgroundAssistantOverlayController? = null
    private val TAG = "Service"

    companion object {
        const val ACTION_START_ASSISTANT = "ACTION_START_ASSISTANT"
        const val ACTION_STOP_ASSISTANT = "ACTION_STOP_ASSISTANT"
    }

    private val alertChannelId = "background_assistant_alerts"
    private val alertNotificationId = 1003

    override fun onBind(intent: Intent?): IBinder? = null

    override fun onCreate() {
        super.onCreate()
        AppLog.i(TAG, "Service onCreate")

        createNotificationChannels()
        startForeground(notificationId, createNotification())

        // Initialize dependencies
        coordinator = NetworkModule.backgroundAssistantCoordinator
        coordinator.start(serviceScope)

        overlayController = BackgroundAssistantOverlayController(
            context = this,
            coordinator = coordinator,
            onError = {
                AppLog.e(TAG, "Overlay error occurred")
                // Post high-priority notification if needed
            }
        )

        wakeWordManager = SensioWakeWordManager(this) {
            AppLog.d(TAG, "Wake word detected")
            handleWakeDetected()
        }

        coordinator.onDismissed = {
            AppLog.d(TAG, "Coordinator dismissed, hiding overlay")
            overlayController?.hide()
            if (NetworkModule.backgroundAssistantModeStore.isEnabled.value) {
                checkPermissionsAndStart(isPeriodic = false)
            }
        }

        // Observe mode store to self-terminate if disabled
        serviceScope.launch {
            NetworkModule.backgroundAssistantModeStore.isEnabled.collectLatest { enabled ->
                if (!enabled) {
                    AppLog.i(TAG, "Background mode disabled, stopping service")
                    stopSelf()
                }
            }
        }

        // Watch for sync readiness
        serviceScope.launch {
            NetworkModule.isTuyaSyncReady.collectLatest { ready ->
                if (ready && NetworkModule.backgroundAssistantModeStore.isEnabled.value) {
                    if (wakeWordManager?.isListeningRequested() != true) {
                        AppLog.i(TAG, "Tuya sync ready, starting assistant")
                        checkPermissionsAndStart(isPeriodic = false)
                    }
                }
            }
        }

        serviceScope.launch {
            while (true) {
                kotlinx.coroutines.delay(5000)
                if (NetworkModule.backgroundAssistantModeStore.isEnabled.value && !coordinator.isSessionActive) {
                    checkPermissionsAndStart(isPeriodic = true)
                }
            }
        }
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        val action = intent?.action
        AppLog.d(TAG, "onStartCommand action: $action")

        when (action) {
            ACTION_START_ASSISTANT -> {
                checkPermissionsAndStart()
            }
            ACTION_STOP_ASSISTANT -> {
                performCleanup()
                stopForeground(true)
                stopSelf()
            }
        }

        return START_STICKY
    }

    private fun checkPermissionsAndStart(isPeriodic: Boolean = false) {
        if (coordinator.isSessionActive) {
            AppLog.d(TAG, "Skipping wake start: Session is active")
            return
        }

        if (!NetworkModule.isTuyaSyncReady.value) {
            if (!isPeriodic) {
                AppLog.i(TAG, "Tuya sync not ready, deferring start")
                showAlertNotification(
                    "Syncing Devices",
                    "Assistant will start once device synchronization is complete."
                )
            }
            return
        }

        val hasMic = androidx.core.content.ContextCompat.checkSelfPermission(
            this,
            android.Manifest.permission.RECORD_AUDIO
        ) == android.content.pm.PackageManager.PERMISSION_GRANTED

        if (hasMic) {
            dismissAlertNotification()
            // Only start if not periodic OR if not already listening
            if (!isPeriodic || wakeWordManager?.isListeningRequested() != true) {
                wakeWordManager?.startListening()
            }
        } else {
            AppLog.w(TAG, "Missing RECORD_AUDIO permission, auto-disabling")
            disableBackgroundModeDueToMicPermission()
        }
    }

    private fun disableBackgroundModeDueToMicPermission() {
        AppLog.i(TAG, "disableBackgroundModeDueToMicPermission called")
        NetworkModule.backgroundAssistantModeStore.setEnabled(false)
        showAlertNotification(
            "Background Assistant Disabled",
            "Microphone permission was removed. Re-enable permission to turn it on again."
        )
        performCleanup()
        stopForeground(true)
        stopSelf()
    }

    private fun handleWakeDetected() {
        serviceScope.launch {
            if (android.provider.Settings.canDrawOverlays(this@BackgroundAssistantService)) {
                dismissAlertNotification()
                overlayController?.show()
                coordinator.onWakeDetected()
            } else {
                AppLog.w(TAG, "Missing overlay permission")
                showAlertNotification(
                    "Overlay Required",
                    "Please grand 'Appear on top' permission to see the assistant UI."
                )
            }
        }
    }

    private fun createNotificationChannels() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val notificationManager = getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager

            // Primary Channel
            val primaryChannel = NotificationChannel(
                channelId,
                "Sensio Background Assistant",
                NotificationManager.IMPORTANCE_LOW
            ).apply {
                description = "Ensures wake-word detection is active"
            }
            notificationManager.createNotificationChannel(primaryChannel)

            // Alert Channel
            val alertChannel = NotificationChannel(
                alertChannelId,
                "Assistant Alerts",
                NotificationManager.IMPORTANCE_HIGH
            ).apply {
                description = "Notifications for missing permissions or errors"
            }
            notificationManager.createNotificationChannel(alertChannel)
        }
    }

    private fun showAlertNotification(title: String, content: String) {
        val notification = NotificationCompat.Builder(this, alertChannelId)
            .setContentTitle(title)
            .setContentText(content)
            .setSmallIcon(android.R.drawable.stat_sys_warning)
            .setPriority(NotificationCompat.PRIORITY_HIGH)
            .setAutoCancel(true)
            .setContentIntent(
                PendingIntent.getActivity(
                    this,
                    1,
                    Intent(this, MainActivity::class.java),
                    PendingIntent.FLAG_IMMUTABLE
                )
            )
            .build()

        val notificationManager = getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager
        notificationManager.notify(alertNotificationId, notification)
    }

    private fun dismissAlertNotification() {
        val notificationManager = getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager
        notificationManager.cancel(alertNotificationId)
    }

    private fun createNotification() = NotificationCompat.Builder(this, channelId)
        .setContentTitle("Sensio Assistant Active")
        .setContentText("Listening for wake word...")
        .setSmallIcon(android.R.drawable.ic_btn_speak_now) // Use a better icon in production
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
        wakeWordManager?.destroy()
        overlayController?.destroy()
        coordinator.stop()
        serviceScope.cancel()
    }

    override fun onDestroy() {
        super.onDestroy()
        AppLog.i(TAG, "Service onDestroy")
        performCleanup()
    }
}
