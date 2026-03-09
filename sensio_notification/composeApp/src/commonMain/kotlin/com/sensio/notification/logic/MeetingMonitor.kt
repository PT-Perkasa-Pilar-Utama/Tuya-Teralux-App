package com.sensio.app.notif.logic

import com.sensio.app.notif.model.MeetingSession
import com.sensio.app.notif.showNotification
import kotlin.time.Duration
import kotlin.time.Duration.Companion.minutes
import kotlinx.datetime.Clock

class MeetingMonitor(
    private var activeSession: MeetingSession?,
    private val reminderOffset: Duration = 10.minutes
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
                message = "The current meeting session is ending soon"
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
