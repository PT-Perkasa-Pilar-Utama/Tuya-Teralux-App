package com.example.whisperandroid.service.reminder

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.util.Log
import com.example.whisperandroid.data.di.NetworkModule
import com.example.whisperandroid.domain.model.reminder.MeetingReminderEntity
import com.example.whisperandroid.domain.model.reminder.MeetingReminderUiModel

/**
 * Broadcast receiver for meeting reminder alarm triggers.
 *
 * Fires notification and requests overlay when alarm triggers.
 */
class MeetingReminderAlarmReceiver : BroadcastReceiver() {
    private val tag = "MeetingReminderAlarm"

    override fun onReceive(context: Context, intent: Intent) {
        val action = intent.action
        val reminderId = intent.getStringExtra(EXTRA_REMINDER_ID)
            ?: intent.getStringExtra(MeetingReminderScheduler.EXTRA_REMINDER_ID)

        Log.d(tag, "Alarm received: action=$action, reminderId=$reminderId")

        // Ensure NetworkModule is initialized
        NetworkModule.ensureInitialized(context)

        val store = NetworkModule.meetingReminderStore
        val notifier = NetworkModule.meetingReminderNotifier
        val overlayController = NetworkModule.meetingReminderOverlayController
        val arbiter = NetworkModule.overlayArbiter

        // Handle immediate trigger (late reminder within grace window)
        if (action == MeetingReminderScheduler.ACTION_IMMEDIATE_TRIGGER) {
            val publishAt = intent.getLongExtra(MeetingReminderScheduler.EXTRA_PUBLISH_AT, 0L)
            val remainingMinutes = intent.getIntExtra(MeetingReminderScheduler.EXTRA_REMAINING_MINUTES, 0)

            if (publishAt > 0 && remainingMinutes > 0) {
                val entity = MeetingReminderEntity(
                    id = reminderId ?: MeetingReminderEntity.generateId(publishAt, remainingMinutes),
                    publishAtEpochMillis = publishAt,
                    remainingMinutes = remainingMinutes,
                    createdAtEpochMillis = System.currentTimeMillis(),
                    fired = false
                )
                fireReminder(context, entity, store, notifier, overlayController, arbiter)
            }
            return
        }

        // Normal scheduled alarm trigger
        if (reminderId != null) {
            val pendingReminders = store.getPendingReminders()
            val entity = pendingReminders.find { it.id == reminderId }

            if (entity != null) {
                fireReminder(context, entity, store, notifier, overlayController, arbiter)
            } else {
                Log.w(tag, "Reminder not found: $reminderId")
            }
        }
    }

    private fun fireReminder(
        context: Context,
        entity: MeetingReminderEntity,
        store: com.example.whisperandroid.data.local.reminder.MeetingReminderStore,
        notifier: MeetingReminderNotifier,
        overlayController: MeetingReminderOverlayController,
        arbiter: OverlayArbiter
    ) {
        Log.i(tag, "Firing reminder: ${entity.id}, remainingMinutes=${entity.remainingMinutes}")

        // Always post notification first (guaranteed delivery)
        val uiModel = MeetingReminderUiModel.fromEntity(entity)
        notifier.showNotification(uiModel, entity.id)

        // Mark as fired in store
        store.markFired(entity.id)

        // Attempt overlay only if allowed by arbiter
        if (arbiter.canShowReminderOverlay()) {
            Log.d(tag, "Overlay allowed, showing reminder overlay")
            overlayController.show(uiModel)
        } else {
            Log.i(tag, "Overlay suppressed (assistant active or permission missing), notification only")
        }
    }

    companion object {
        const val EXTRA_REMINDER_ID = MeetingReminderScheduler.EXTRA_REMINDER_ID
    }
}
