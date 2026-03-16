package com.example.whisperandroid.service

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.app.Service
import android.content.Intent
import android.os.Build
import android.os.IBinder
import android.util.Log
import androidx.core.app.NotificationCompat
import com.example.whisperandroid.MainActivity
import com.example.whisperandroid.data.di.NetworkModule
import com.example.whisperandroid.data.manager.MeetingProcessManager
import com.example.whisperandroid.domain.usecase.MeetingProcessState
import java.io.File
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.cancel
import kotlinx.coroutines.launch

class MeetingForegroundService : Service() {

    private val serviceScope = CoroutineScope(Dispatchers.IO + SupervisorJob())
    private val notificationId = 1001
    private val channelId = "meeting_processing_channel"

    companion object {
        const val ACTION_CANCEL = "com.example.whisperandroid.service.ACTION_CANCEL"

        // Track if service has an active processing job
        // Using a top-level volatile variable to avoid lifecycle race conditions
        @Volatile
        private var currentProcessingJob: kotlinx.coroutines.Job? = null

        // Service lifecycle state (is the service instance created/running)
        @Volatile
        private var isServiceCreated: Boolean = false

        // Service lifecycle: is the service instance alive (regardless of whether it's processing)
        fun isServiceRunning(): Boolean = isServiceCreated

        // Processing state: is there an active job running (subset of service running)
        fun isProcessingActive(): Boolean = currentProcessingJob != null
    }

    private var processingJob: kotlinx.coroutines.Job? = null
    private var currentAudioPath: String? = null

    override fun onBind(intent: Intent?): IBinder? = null

    override fun onCreate() {
        super.onCreate()
        isServiceCreated = true
        createNotificationChannel()
    }

    override fun onDestroy() {
        super.onDestroy()
        isServiceCreated = false
        // DO NOT cancel serviceScope here - this would trigger !isActive paths in ProcessMeetingUseCase
        // and clear the submission state, breaking restart safety.
        // The processing job will be cancelled separately if user initiates cancellation.
        // Coroutines in serviceScope will complete naturally or be cancelled via processingJob.cancel()
        Log.d("MeetingForegroundService", "Service destroyed - preserving submission state for resume capability")
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        when (intent?.action) {
            ACTION_CANCEL -> {
                handleCancel()
                return START_NOT_STICKY
            }
        }

        val audioPath = intent?.getStringExtra("AUDIO_PATH") ?: return START_NOT_STICKY
        val token = intent.getStringExtra("TOKEN") ?: return START_NOT_STICKY
        val targetLang = intent.getStringExtra("TARGET_LANG") ?: "English"
        val macAddress = intent.getStringExtra("MAC_ADDRESS")

        val audioFile = File(audioPath)
        if (!audioFile.exists()) {
            stopSelf()
            return START_NOT_STICKY
        }

        // Store the audio path for cancellation cleanup
        currentAudioPath = audioPath
        MeetingProcessManager.setCurrentAudioPath(audioPath)

        startForeground(notificationId, createNotification("Starting process..."))

        val job = serviceScope.launch {
            try {
                // Load existing submission state to check for resume scenario
                // Idempotency key must be stable for one logical submission lifecycle
                val prefs = getSharedPreferences("upload_sessions", MODE_PRIVATE)
                val submissionStateKey = "submission_" + audioFile.absolutePath
                val submissionStateJson = prefs.getString(submissionStateKey, null)

                val submissionIdempotencyKey = if (submissionStateJson != null) {
                    // Resume existing submission: reuse saved idempotency key
                    try {
                        val obj = org.json.JSONObject(submissionStateJson)
                        val savedKey = obj.getString("idempotencyKey")
                        Log.d("MeetingForegroundService", "Resuming submission with saved idempotency key for: $audioPath")
                        savedKey
                    } catch (e: Exception) {
                        Log.w("MeetingForegroundService", "Failed to parse submission state, generating new key: ${e.message}")
                        "meeting_${audioFile.lastModified()}_${System.currentTimeMillis()}"
                    }
                } else {
                    // Fresh submission: generate new idempotency key
                    "meeting_${audioFile.lastModified()}_${System.currentTimeMillis()}"
                }

                NetworkModule.processMeetingUseCase(
                    audioFile = audioFile,
                    token = token,
                    targetLang = targetLang,
                    macAddress = macAddress,
                    idempotencyKey = submissionIdempotencyKey
                ).collect { state ->
                    MeetingProcessManager.updateState(state)
                    updateNotification(state)
                    if (state is MeetingProcessState.Success ||
                        state is MeetingProcessState.Error ||
                        state is MeetingProcessState.Cancelled
                    ) {
                        showFinalNotification(state)
                        stopForeground(false)
                        stopSelf()
                    }
                }
            } catch (e: CancellationException) {
                // User-initiated cancellation - don't show error state
                Log.d("MeetingForegroundService", "Processing cancelled: ${e.message}")
                MeetingProcessManager.updateState(MeetingProcessState.Cancelled)
                stopForeground(false)
                stopSelf()
            } catch (e: Exception) {
                val errorState = MeetingProcessState.Error(e.message ?: "Unknown error")
                MeetingProcessManager.updateState(errorState)
                showFinalNotification(errorState)
                stopForeground(false)
                stopSelf()
            } finally {
                // Clear the job reference when processing completes
                processingJob = null
                currentProcessingJob = null
                currentAudioPath = null
            }
        }

        // Update job references after launch
        processingJob = job
        currentProcessingJob = job

        return START_NOT_STICKY
    }

    private fun handleCancel() {
        // Cancel the processing job
        processingJob?.cancel()
        processingJob = null
        currentProcessingJob = null

        // Clear the persisted session state to prevent resume of abandoned upload
        currentAudioPath?.let { audioPath ->
            NetworkModule.processMeetingUseCase.clearSessionState(audioPath)
        }
        currentAudioPath = null

        // Update state to Cancelled (not Error)
        MeetingProcessManager.cancel()

        // Stop the service
        stopForeground(false)
        stopSelf()
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
        // Don't show a notification for cancelled state - user initiated this
        if (state is MeetingProcessState.Cancelled) {
            return
        }

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
