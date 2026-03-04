package com.sensio.notification.logic

import com.sensio.notification.model.MeetingSession
import com.sensio.notification.showNotification
import kotlinx.datetime.Clock
import kotlin.time.Duration
import kotlin.time.Duration.Companion.minutes

class MeetingMonitor(
    private var activeSession: MeetingSession?,
    private val reminderOffset: Duration = 10.minutes,
) {
    fun updateSession(session: MeetingSession?) {
        activeSession = session
    }

    fun checkAndTrigger() {
        val session = activeSession ?: return
        if (session.reminderTriggered) return

        val now = Clock.System.now()
        val timeUntilEnd = session.endTime - now

        if (timeUntilEnd <= reminderOffset && timeUntilEnd > Duration.ZERO) {
            showNotification(
                title = "Sensio Notification",
                message = "You have 10 minutes remaining",
            )
            session.reminderTriggered = true
        }
    }

    fun extendMeeting(duration: Duration) {
        activeSession?.let {
            val newEndTimeString = (it.endTime + duration).toString()
            // In a real app, this would update a backend/state
            // it.endTime += duration
            // it.reminderTriggered = false
        }
    }
}
