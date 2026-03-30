package com.example.whisperandroid.presentation.reminder

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
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
import com.example.whisperandroid.domain.model.reminder.MeetingReminderUiModel
import com.example.whisperandroid.presentation.components.SensioButton
import com.example.whisperandroid.ui.theme.Cyan600
import com.example.whisperandroid.ui.theme.Slate50
import com.example.whisperandroid.ui.theme.Slate500
import com.example.whisperandroid.ui.theme.Slate700
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
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Slate500.copy(alpha = 0.3f)),
        contentAlignment = Alignment.Center
    ) {
        Card(
            modifier = Modifier
                .fillMaxWidth()
                .width(420.dp)
                .padding(16.dp),
            shape = RoundedCornerShape(28.dp),
            colors = CardDefaults.cardColors(
                containerColor = Slate50
            ),
            elevation = CardDefaults.cardElevation(defaultElevation = 12.dp)
        ) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
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
                        modifier = Modifier.size(40.dp)
                    )
                }

                Spacer(modifier = Modifier.height(20.dp))

                // Small label
                Text(
                    text = "SENSIO REMINDER",
                    fontSize = 12.sp,
                    fontWeight = FontWeight.SemiBold,
                    color = Cyan600,
                    letterSpacing = 1.sp
                )

                Spacer(modifier = Modifier.height(12.dp))

                // Title
                Text(
                    text = uiModel.title,
                    fontSize = 24.sp,
                    fontWeight = FontWeight.Bold,
                    color = Slate900,
                    letterSpacing = (-0.5).sp
                )

                Spacer(modifier = Modifier.height(16.dp))

                // Message
                Text(
                    text = uiModel.message,
                    fontSize = 18.sp,
                    fontWeight = FontWeight.Normal,
                    color = Slate700,
                    lineHeight = 28.sp,
                    letterSpacing = 0.2.sp
                )

                Spacer(modifier = Modifier.height(32.dp))

                // Close button
                SensioButton(
                    text = "Selesai",
                    onClick = onClose,
                    modifier = Modifier.fillMaxWidth()
                )
            }
        }
    }
}
