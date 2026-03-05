package com.example.whisperandroid.service

import android.app.Notification
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
import com.example.whisperandroid.data.manager.MeetingProcessManager
import com.example.whisperandroid.domain.usecase.MeetingProcessState
import java.io.File
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.launch

class MeetingForegroundService : Service() {

    private val serviceScope = CoroutineScope(Dispatchers.IO + SupervisorJob())
    private val notificationId = 1001
    private val channelId = "meeting_processing_channel"

    override fun onBind(intent: Intent?): IBinder? = null

    override fun onCreate() {
        super.onCreate()
        createNotificationChannel()
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        val audioPath = intent?.getStringExtra("AUDIO_PATH") ?: return START_NOT_STICKY
        val token = intent.getStringExtra("TOKEN") ?: return START_NOT_STICKY
        val targetLang = intent.getStringExtra("TARGET_LANG") ?: "English"
        val macAddress = intent.getStringExtra("MAC_ADDRESS")

        val audioFile = File(audioPath)
        if (!audioFile.exists()) {
            stopSelf()
            return START_NOT_STICKY
        }

        startForeground(notificationId, createNotification("Starting process..."))

        serviceScope.launch {
            try {
                NetworkModule.processMeetingUseCase(
                    audioFile = audioFile,
                    token = token,
                    targetLang = targetLang,
                    macAddress = macAddress,
                    waitSignal = { taskName ->
                        // If we need to wait for MQTT signals, we do it here
                        // For simplicity in background, we might assume MqttHelper handles it
                        // but since ProcessMeetingUseCase needs a waitSignal block,
                        // we can reuse the logic or simplify for service.
                        // Currently, ProcessMeetingUseCase.invoke needs this.
                        android.util.Log.d("MeetingService", "Waiting for signal: $taskName")
                        // Implementation of waitSignal logic for service
                        val signalChannel = kotlinx.coroutines.channels.Channel<String>(1)
                        val job = launch {
                            NetworkModule.mqttHelper.messages.collect { (topic, msg) ->
                                val taskTopic = NetworkModule.mqttHelper.getTaskTopic()
                                if (topic == taskTopic) {
                                    val json = com.google.gson.JsonParser.parseString(msg).asJsonObject
                                    if (json.has("event") && json.get("event").asString == "stop") {
                                        val receivedTask = json.get("task").asString
                                        signalChannel.trySend(receivedTask)
                                    }
                                }
                            }
                        }

                        NetworkModule.mqttHelper.publishTaskMessage("start", taskName)
                        while (true) {
                            if (signalChannel.receive() == taskName) break
                        }
                        job.cancel()
                    }
                ).collect { state ->
                    MeetingProcessManager.updateState(state)
                    updateNotification(state)
                    if (state is MeetingProcessState.Success || state is MeetingProcessState.Error) {
                        showFinalNotification(state)
                        stopForeground(false)
                        stopSelf()
                    }
                }
            } catch (e: Exception) {
                val errorState = MeetingProcessState.Error(e.message ?: "Unknown error")
                MeetingProcessManager.updateState(errorState)
                showFinalNotification(errorState)
                stopForeground(false)
                stopSelf()
            }
        }

        return START_NOT_STICKY
    }

    private fun createNotificationChannel() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val channel = NotificationChannel(
                channelId,
                "Meeting Processing",
                NotificationManager.IMPORTANCE_LOW
            )
            val manager = getSystemService(NotificationManager::class.java)
            manager.createNotificationChannel(channel)
        }
    }

    private fun createNotification(content: String): Notification {
        val intent = Intent(this, MainActivity::class.java)
        val pendingIntent = PendingIntent.getActivity(
            this,
            0,
            intent,
            PendingIntent.FLAG_IMMUTABLE or PendingIntent.FLAG_UPDATE_CURRENT
        )

        return NotificationCompat.Builder(this, channelId)
            .setContentTitle("Processing Meeting")
            .setContentText(content)
            .setSmallIcon(android.R.drawable.stat_notify_sync)
            .setContentIntent(pendingIntent)
            .setOngoing(true)
            .build()
    }

    private fun updateNotification(state: MeetingProcessState) {
        val message = when (state) {
            is MeetingProcessState.Uploading -> "Uploading audio..."
            is MeetingProcessState.Transcribing -> "Transcribing..."
            is MeetingProcessState.Translating -> "Translating..."
            is MeetingProcessState.Summarizing -> "Summarizing..."
            else -> return
        }
        val notificationManager = getSystemService(NotificationManager::class.java)
        notificationManager.notify(notificationId, createNotification(message))
    }

    private fun showFinalNotification(state: MeetingProcessState) {
        val intent = Intent(this, MainActivity::class.java)
        val pendingIntent = PendingIntent.getActivity(
            this,
            0,
            intent,
            PendingIntent.FLAG_IMMUTABLE or PendingIntent.FLAG_UPDATE_CURRENT
        )

        val builder = NotificationCompat.Builder(this, channelId)
            .setSmallIcon(android.R.drawable.stat_notify_sync)
            .setAutoCancel(true)
            .setContentIntent(pendingIntent)

        if (state is MeetingProcessState.Success) {
            builder.setContentTitle("Process Complete")
                .setContentText("Meeting summary is ready.")
        } else if (state is MeetingProcessState.Error) {
            builder.setContentTitle("Process Failed")
                .setContentText(state.message)
        }

        val notificationManager = getSystemService(NotificationManager::class.java)
        notificationManager.notify(notificationId + 1, builder.build())
    }
}
