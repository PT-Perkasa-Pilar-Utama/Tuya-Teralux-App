package com.example.whisperandroid.service.reminder

import android.content.Context
import android.util.Log
import com.example.whisperandroid.data.di.NetworkModule
import com.example.whisperandroid.data.local.reminder.MeetingReminderStore
import com.example.whisperandroid.domain.model.reminder.MeetingReminderEntity
import com.example.whisperandroid.domain.model.reminder.MeetingReminderMessage
import com.example.whisperandroid.util.MeetingReminderParser
import kotlinx.coroutines.CoroutineScope
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

        // Build notification topic (no environment segment to match backend)
        val username = getUsername()
        notificationTopic = "users/$username/notification"

        Log.i(tag, "Starting reminder coordinator, topic: $notificationTopic")

        // Restore pending reminders on startup
        restorePendingReminders()

        // Listen to MQTT messages
        scope.launch {
            mqttHelper.messages.collectLatest { (topic, payload) ->
                if (topic == notificationTopic) {
                    Log.d(tag, "Notification message received: $payload")
                    onNotificationMessage(payload)
                }
            }
        }

        // Subscribe to the topic
        mqttHelper.subscribe(notificationTopic!!)
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
