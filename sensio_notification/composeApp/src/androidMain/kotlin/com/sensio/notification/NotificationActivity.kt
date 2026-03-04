package com.sensio.notification

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import com.sensio.notification.ui.NotificationModal

class NotificationActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        val title = intent.getStringExtra("title") ?: "Meeting Reminder"
        val message = intent.getStringExtra("message") ?: "You have 10 minutes remaining"

        setContent {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                NotificationModal(
                    title = title,
                    message = message,
                    onDismiss = {
                        finish()
                    },
                )
            }
        }
    }
}
