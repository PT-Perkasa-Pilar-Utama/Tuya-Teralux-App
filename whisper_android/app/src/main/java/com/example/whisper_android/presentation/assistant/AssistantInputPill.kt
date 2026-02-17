package com.example.whisper_android.presentation.assistant

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.Send
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.whisper_android.presentation.components.MicButton

@Composable
fun AssistantInputPill(
    inputValue: String,
    onValueChange: (String) -> Unit,
    isRecording: Boolean,
    isProcessing: Boolean,
    hasPermission: Boolean,
    onMicClick: () -> Unit,
    onStopClick: () -> Unit,
    onSendClick: () -> Unit,
    enabled: Boolean = true,
    modifier: Modifier = Modifier
) {
    Surface(
        modifier = modifier
            .fillMaxWidth()
            .height(64.dp) // Professional 64dp height
            .padding(horizontal = 4.dp),
        shape = RoundedCornerShape(32.dp), // Height / 2
        color = Color.White.copy(alpha = 0.95f), // High-Fi opacity
        tonalElevation = 2.dp,
        shadowElevation = 6.dp
    ) {
        Row(
            modifier = Modifier
                .fillMaxSize()
                .padding(horizontal = 8.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            // Reusable Mic Button
            MicButton(
                isRecording = isRecording,
                hasPermission = hasPermission,
                isProcessing = isProcessing,
                onClick = if (enabled) onMicClick else ({}),
                size = 40.dp,
                modifier = Modifier.alpha(if (enabled) 1f else 0.4f)
            )

            // Red Stop Button (Shown when active)
            if (isRecording || isProcessing) {
                Spacer(modifier = Modifier.width(8.dp))
                IconButton(
                    onClick = onStopClick,
                    modifier = Modifier.size(32.dp)
                ) {
                    Surface(
                        modifier = Modifier.size(16.dp),
                        color = Color(0xFFEF5350), // Red
                        shape = RoundedCornerShape(2.dp)
                    ) {}
                }
            }

            // Input TextField
            TextField(
                value = inputValue,
                onValueChange = onValueChange,
                modifier = Modifier
                    .weight(1f)
                    .padding(horizontal = 4.dp),
                placeholder = {
                    Text(
                        text = if (isRecording) "Recording..." else if (isProcessing) "Thinking..." else "Ask Intelligence...",
                        fontSize = 15.sp,
                        fontWeight = FontWeight.Medium,
                        color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.4f)
                    )
                },
                colors = TextFieldDefaults.colors(
                    focusedContainerColor = Color.Transparent,
                    unfocusedContainerColor = Color.Transparent,
                    disabledContainerColor = Color.Transparent,
                    focusedIndicatorColor = Color.Transparent,
                    unfocusedIndicatorColor = Color.Transparent,
                    disabledIndicatorColor = Color.Transparent,
                    cursorColor = MaterialTheme.colorScheme.primary,
                    focusedTextColor = MaterialTheme.colorScheme.onSurface,
                    unfocusedTextColor = MaterialTheme.colorScheme.onSurface
                ),
                singleLine = true,
                enabled = enabled && !isRecording && !isProcessing,
                textStyle = MaterialTheme.typography.bodyLarge.copy(fontSize = 15.sp)
            )

            // Send Icon Button
            IconButton(
                onClick = onSendClick,
                enabled = enabled && !isRecording && !isProcessing && inputValue.isNotBlank()
            ) {
                Icon(
                    imageVector = Icons.AutoMirrored.Filled.Send,
                    contentDescription = "Send",
                    tint = if (!isRecording && !isProcessing && inputValue.isNotBlank()) 
                        MaterialTheme.colorScheme.primary 
                    else 
                        MaterialTheme.colorScheme.primary.copy(alpha = 0.2f),
                    modifier = Modifier.size(26.dp)
                )
            }
        }
    }
}
