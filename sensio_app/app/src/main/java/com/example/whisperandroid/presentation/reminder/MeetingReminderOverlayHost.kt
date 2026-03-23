package com.example.whisperandroid.presentation.reminder

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Notifications
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.window.Dialog
import com.example.whisperandroid.domain.model.reminder.MeetingReminderUiModel
import com.example.whisperandroid.presentation.components.SensioButton
import com.example.whisperandroid.ui.theme.Cyan600
import com.example.whisperandroid.ui.theme.Slate400
import com.example.whisperandroid.ui.theme.Slate50
import com.example.whisperandroid.ui.theme.Slate900

/**
 * Composable overlay UI for meeting reminders.
 *
 * Displays a centered glass-like card with reminder content.
 */
@Composable
fun MeetingReminderOverlayHost(
    uiModel: MeetingReminderUiModel,
    onClose: () -> Unit
) {
    Dialog(
        onDismissRequest = onClose
    ) {
        Card(
            modifier = Modifier
                .fillMaxWidth()
                .width(420.dp)
                .padding(16.dp),
            shape = RoundedCornerShape(28.dp),
            colors = CardDefaults.cardColors(
                containerColor = Slate900.copy(alpha = 0.92f)
            ),
            elevation = CardDefaults.cardElevation(defaultElevation = 8.dp)
        ) {
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(24.dp),
                horizontalAlignment = Alignment.CenterHorizontally
            ) {
                // Icon chip
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.Center
                ) {
                    Icon(
                        imageVector = Icons.Default.Notifications,
                        contentDescription = null,
                        tint = Cyan600,
                        modifier = Modifier.size(32.dp)
                    )
                }

                Spacer(modifier = Modifier.height(16.dp))

                // Small label
                Text(
                    text = "Sensio Notification",
                    fontSize = 12.sp,
                    fontWeight = FontWeight.Medium,
                    color = Slate400,
                    letterSpacing = 0.5.sp
                )

                Spacer(modifier = Modifier.height(8.dp))

                // Title
                Text(
                    text = uiModel.title,
                    fontSize = 22.sp,
                    fontWeight = FontWeight.Bold,
                    color = Slate50,
                    letterSpacing = (-0.5).sp
                )

                Spacer(modifier = Modifier.height(12.dp))

                // Message
                Text(
                    text = uiModel.message,
                    fontSize = 16.sp,
                    fontWeight = FontWeight.Normal,
                    color = Slate400,
                    lineHeight = 24.sp,
                    letterSpacing = 0.2.sp
                )

                Spacer(modifier = Modifier.height(24.dp))

                // Close button
                SensioButton(
                    text = "Close",
                    onClick = onClose,
                    modifier = Modifier.fillMaxWidth()
                )
            }
        }
    }
}
