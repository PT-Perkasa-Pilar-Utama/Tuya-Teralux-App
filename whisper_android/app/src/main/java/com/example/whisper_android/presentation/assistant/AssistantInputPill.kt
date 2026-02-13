package com.example.whisper_android.presentation.assistant

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.Send
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
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
    modifier: Modifier = Modifier
) {
    Surface(
        modifier = modifier
            .fillMaxWidth()
            .height(56.dp),
        shape = RoundedCornerShape(28.dp),
        color = Color.White,
        shadowElevation = 4.dp
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
                onClick = onMicClick,
                size = 40.dp
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
                        text = if (isRecording) "Recording..." else if (isProcessing) "Thinking..." else "Ask me anything...",
                        fontSize = 16.sp,
                        color = if (isRecording || isProcessing) MaterialTheme.colorScheme.primary else Color.Gray
                    )
                },
                colors = TextFieldDefaults.colors(
                    focusedContainerColor = Color.Transparent,
                    unfocusedContainerColor = Color.Transparent,
                    disabledContainerColor = Color.Transparent,
                    focusedIndicatorColor = Color.Transparent,
                    unfocusedIndicatorColor = Color.Transparent,
                    disabledIndicatorColor = Color.Transparent,
                    cursorColor = MaterialTheme.colorScheme.primary
                ),
                singleLine = true,
                enabled = !isRecording && !isProcessing,
                textStyle = MaterialTheme.typography.bodyLarge.copy(fontSize = 16.sp)
            )

            // Send Icon Button
            IconButton(
                onClick = onSendClick,
                enabled = !isRecording && !isProcessing && inputValue.isNotBlank()
            ) {
                Icon(
                    imageVector = Icons.AutoMirrored.Filled.Send,
                    contentDescription = "Send",
                    tint = if (!isRecording && !isProcessing && inputValue.isNotBlank()) 
                        MaterialTheme.colorScheme.primary 
                    else 
                        Color.LightGray,
                    modifier = Modifier.size(24.dp)
                )
            }
        }
    }
}
