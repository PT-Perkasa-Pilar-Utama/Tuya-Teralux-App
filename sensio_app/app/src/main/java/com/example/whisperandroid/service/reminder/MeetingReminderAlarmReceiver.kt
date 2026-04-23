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

        NetworkModule.ensureInitialized(context)

        val store = NetworkModule.meetingReminderStore
        val notifier = NetworkModule.meetingReminderNotifier
        val overlayController = NetworkModule.meetingReminderOverlayController
        val arbiter = NetworkModule.overlayArbiter

        if (action == MeetingReminderScheduler.ACTION_IMMEDIATE_TRIGGER) {
            if (reminderId != null) {
                val pendingReminders = store.getPendingReminders()
                val entity = pendingReminders.find { it.id == reminderId }

                if (entity != null) {
                    fireReminder(context, entity, store, notifier, overlayController, arbiter)
                } else {
                    Log.w(tag, "Reminder not found for immediate trigger: $reminderId")
                }
            }
            return
        }

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
        Log.i(tag, "Firing reminder: ${entity.id}, eventType=${entity.eventType}")

        val uiModel = MeetingReminderUiModel.fromEntity(entity)
        notifier.showNotification(uiModel, entity.id)

        store.markFired(entity.id)

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
