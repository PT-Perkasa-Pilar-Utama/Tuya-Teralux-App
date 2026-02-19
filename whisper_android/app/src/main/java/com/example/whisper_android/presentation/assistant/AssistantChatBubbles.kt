package com.example.whisper_android.presentation.assistant

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.whisper_android.presentation.components.MessageRole
import com.example.whisper_android.presentation.components.TranscriptionMessage

@Composable
fun AssistantChatBubble(message: TranscriptionMessage) {
    val isUser = message.role == MessageRole.USER

    Column(
        modifier =
            Modifier
                .fillMaxWidth()
                .padding(horizontal = 12.dp),
        horizontalAlignment = if (isUser) Alignment.End else Alignment.Start,
    ) {
        Surface(
            shape =
                RoundedCornerShape(
                    topStart = if (isUser) 24.dp else 4.dp,
                    topEnd = if (isUser) 4.dp else 24.dp,
                    bottomStart = 24.dp,
                    bottomEnd = 24.dp,
                ),
            color =
                if (isUser) {
                    MaterialTheme.colorScheme.primary
                } else {
                    Color.White.copy(alpha = 0.9f)
                },
            modifier =
                Modifier
                    .widthIn(max = 300.dp),
            border =
                if (!isUser) {
                    androidx.compose.foundation.BorderStroke(
                        1.dp,
                        MaterialTheme.colorScheme.primary.copy(alpha = 0.08f),
                    )
                } else {
                    null
                },
            shadowElevation = if (isUser) 6.dp else 8.dp,
            tonalElevation = if (isUser) 0.dp else 4.dp,
        ) {
            Text(
                text = message.text,
                modifier = Modifier.padding(horizontal = 20.dp, vertical = 14.dp),
                style =
                    MaterialTheme.typography.bodyLarge.copy(
                        lineHeight = 24.sp,
                        fontSize = 15.sp,
                        fontWeight = if (isUser) FontWeight.Medium else FontWeight.Normal,
                    ),
                color =
                    if (isUser) {
                        Color.White
                    } else {
                        MaterialTheme.colorScheme.onSurface
                    },
            )
        }
    }
}
