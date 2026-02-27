package com.sensio.notification.model

import kotlinx.datetime.Instant

data class MeetingSession(
    val id: String,
    val title: String,
    val endTime: Instant,
    var reminderTriggered: Boolean = false,
)
