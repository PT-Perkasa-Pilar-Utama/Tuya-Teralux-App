package com.sensio.notification

import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.unit.DpSize
import androidx.compose.ui.unit.dp
import androidx.compose.ui.window.Window
import androidx.compose.ui.window.WindowPosition
import androidx.compose.ui.window.application
import androidx.compose.ui.window.rememberWindowState
import com.sensio.notification.logic.MeetingMonitor
import com.sensio.notification.model.MeetingSession
import com.sensio.notification.ui.NotificationModal
import kotlinx.coroutines.delay
import kotlinx.datetime.Clock
import kotlin.time.Duration.Companion.minutes
import kotlin.time.Duration.Companion.seconds

fun main() =
    application {
        var showNotification by remember { mutableStateOf(false) }
        var notificationTitle by remember { mutableStateOf("") }
        var notificationMessage by remember { mutableStateOf("") }

        val mockSession =
            remember {
                MeetingSession(
                    id = "1",
                    title = "Project Sync",
                    endTime = Clock.System.now() + 6.minutes,
                )
            }

        val monitor = remember { MeetingMonitor(mockSession) }

        // Backend loop
        LaunchedEffect(Unit) {
            // In a real desktop app, this would run in a system tray worker
            // Here we just use the main application scope for demo
            while (true) {
                val now = Clock.System.now()
                val timeUntilEnd = mockSession.endTime - now

                if (timeUntilEnd <= 10.minutes && !mockSession.reminderTriggered) {
                    notificationTitle = "Meeting Reminder"
                    notificationMessage = "You have 10 minutes remaining"
                    showNotification = true
                    mockSession.reminderTriggered = true
                }
                delay(5.seconds)
            }
        }

        if (showNotification) {
            Window(
                onCloseRequest = { showNotification = false },
                state =
                    rememberWindowState(
                        position = WindowPosition(Alignment.Center),
                        size = DpSize(450.dp, 300.dp),
                    ),
                title = "Reminder",
                transparent = true,
                undecorated = true,
                alwaysOnTop = true,
            ) {
                NotificationModal(
                    title = notificationTitle,
                    message = notificationMessage,
                    onExtend = {
                        // Logic to extend meeting
                        mockSession.reminderTriggered = false
                        // We'd typically update endTime here
                        showNotification = false
                    },
                    onDismiss = {
                        showNotification = false
                    },
                )
            }
        }
    }
