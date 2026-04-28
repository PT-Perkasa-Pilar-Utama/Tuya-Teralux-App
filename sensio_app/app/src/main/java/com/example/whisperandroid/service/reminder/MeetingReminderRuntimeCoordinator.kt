package com.example.whisperandroid.service.reminder

import android.content.Context
import android.util.Log
import com.example.whisperandroid.BuildConfig
import com.example.whisperandroid.data.di.NetworkModule
import com.example.whisperandroid.data.local.reminder.MeetingReminderStore
import com.example.whisperandroid.domain.model.reminder.MeetingReminderEntity
import com.example.whisperandroid.domain.model.reminder.MeetingReminderMessage
import com.example.whisperandroid.util.MeetingReminderParser
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.collectLatest
import kotlinx.coroutines.launch

/**
 * Central coordinator for meeting reminder runtime.
 *
 * Listens to MQTT notification messages, persists reminders, and schedules alarms.
 */
class MeetingReminderRuntimeCoordinator(
    private val context: Context,
    private val store: MeetingReminderStore,
    private val scheduler: MeetingReminderScheduler,
    private val notifier: MeetingReminderNotifier,
    private val overlayController: MeetingReminderOverlayController,
    private val arbiter: OverlayArbiter
) {
    private val mqttHelper = NetworkModule.mqttHelper
    private val tag = "ReminderCoordinator"

    private var notificationTopic: String? = null
    private var isStarted = false

    /**
     * Start the reminder coordinator.
     *
     * Subscribes to notification topic and begins listening for messages.
     */
    fun start(scope: CoroutineScope) {
        if (isStarted) {
            Log.w(tag, "Already started, skipping")
            return
        }

        isStarted = true

        // Build notification topic with environment segment
        val username = getUsername()
        val env = BuildConfig.APPLICATION_ENVIRONMENT
        notificationTopic = "users/$username/$env/notification"

        Log.i(tag, "Starting reminder coordinator, topic: $notificationTopic")

        // Restore pending reminders on startup
        restorePendingReminders()

        // MQTT connection is now managed by Foreground Service
        // Check connection status and log warning if not connected
        if (mqttHelper.connectionStatus.value != com.example.whisperandroid.util.MqttHelper.MqttConnectionStatus.CONNECTED) {
            Log.w(tag, "MQTT not yet connected, topic subscription will activate when connection is established")
        }

        // Listen to MQTT messages
        scope.launch {
            // Subscribe to the topic (MQTT should already be connected via service)
            mqttHelper.subscribe(notificationTopic!!)

            // Collect messages
            mqttHelper.messages.collectLatest { (topic, payload) ->
                if (topic == notificationTopic) {
                    Log.d(tag, "Notification message received: $payload")
                    onNotificationMessage(payload)
                }
            }
        }
    }

    /**
     * Wait for MQTT connection with retry logic and timeout.
     *
     * @return true if connected successfully, false if timeout or unrecoverable failure
     */
    private suspend fun waitForMqttConnection(timeoutMs: Long = 30_000): Boolean {
        val startTime = System.currentTimeMillis()

        while (System.currentTimeMillis() - startTime < timeoutMs) {
            when (val status = mqttHelper.connectionStatus.value) {
                com.example.whisperandroid.util.MqttHelper.MqttConnectionStatus.CONNECTED -> {
                    Log.d(tag, "MQTT connection established")
                    return true
                }
                com.example.whisperandroid.util.MqttHelper.MqttConnectionStatus.FAILED,
                com.example.whisperandroid.util.MqttHelper.MqttConnectionStatus.NO_INTERNET -> {
                    Log.w(tag, "Connection status: $status, retrying in 5s...")
                    delay(5000)
                    mqttHelper.connect()
                }
                com.example.whisperandroid.util.MqttHelper.MqttConnectionStatus.DISCONNECTED -> {
                    Log.w(tag, "Connection status: $status, connecting...")
                    mqttHelper.connect()
                    delay(1000)
                }
                com.example.whisperandroid.util.MqttHelper.MqttConnectionStatus.CONNECTING -> {
                    Log.d(tag, "Connection status: $status, waiting...")
                    delay(1000)
                }
                else -> {
                    Log.d(tag, "Unknown connection status: $status, waiting...")
                    delay(1000)
                }
            }
        }

        Log.e(tag, "MQTT connection timeout after ${timeoutMs}ms")
        return false
    }

    /**
     * Stop the reminder coordinator.
     */
    fun stop() {
        if (!isStarted) return

        Log.i(tag, "Stopping reminder coordinator")
        isStarted = false

        notificationTopic?.let { topic ->
            // Note: MqttHelper doesn't have unsubscribe, but that's okay for shutdown
        }
    }

    /**
     * Handle incoming notification message.
     */
    private fun onNotificationMessage(payload: String) {
        val message = MeetingReminderParser.parse(payload)
            ?: run {
                Log.w(tag, "Invalid notification payload, ignoring")
                return
            }

        processReminderMessage(message)
    }

    /**
     * Process a parsed reminder message.
     */
    private fun processReminderMessage(message: MeetingReminderMessage) {
        val publishAtMillis = MeetingReminderParser.parseTimestamp(message.publishAt)
            ?: run {
                Log.w(tag, "Cannot parse publish_at timestamp, ignoring")
                return
            }

        val currentTime = System.currentTimeMillis()
        val id = MeetingReminderEntity.generateId(publishAtMillis, message.remainingMinutes)

        val entity = MeetingReminderEntity(
            id = id,
            publishAtEpochMillis = publishAtMillis,
            remainingMinutes = message.remainingMinutes,
            createdAtEpochMillis = currentTime,
            fired = false
        )

        // Save and schedule
        store.savePending(entity)
        scheduler.scheduleReminder(entity)

        Log.i(tag, "Processed reminder: id=$id, fireAt=$publishAtMillis, remainingMin=${message.remainingMinutes}")
    }

    /**
     * Restore pending reminders from store and reschedule alarms.
     */
    private fun restorePendingReminders() {
        val currentTime = System.currentTimeMillis()

        // Prune stale reminders first
        store.pruneStale(currentTime)

        // Get valid pending reminders
        val pendingReminders = store.getValidPendingReminders(currentTime)

        if (pendingReminders.isNotEmpty()) {
            scheduler.rescheduleAll(pendingReminders)
            Log.i(tag, "Restored ${pendingReminders.size} pending reminders")
        } else {
            Log.d(tag, "No pending reminders to restore")
        }
    }

    /**
     * Get username for topic construction.
     */
    private fun getUsername(): String {
        return com.example.whisperandroid.util.DeviceUtils.getDeviceId(context)
    }

    /**
     * Get the notification topic.
     */
    fun getNotificationTopic(): String? = notificationTopic
}
