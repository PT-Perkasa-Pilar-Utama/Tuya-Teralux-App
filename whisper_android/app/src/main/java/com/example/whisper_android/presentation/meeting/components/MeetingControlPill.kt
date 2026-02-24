package com.example.whisper_android.presentation.meeting.components

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.DeleteOutline
import androidx.compose.material.icons.filled.FolderOpen
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.whisper_android.domain.usecase.MeetingProcessState
import com.example.whisper_android.presentation.components.MicButton

@Composable
fun MeetingControlPill(
    isRecording: Boolean,
    hasPermission: Boolean,
    uiState: MeetingProcessState,
    pulseScale: Float,
    onMicClick: () -> Unit,
    onUploadClick: () -> Unit,
    onStopClick: () -> Unit,
    onClearClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    Box(
        modifier =
        modifier
            .fillMaxWidth()
            .padding(bottom = 8.dp),
        contentAlignment = Alignment.Center
    ) {
        Surface(
            modifier =
            Modifier
                .height(48.dp)
                .widthIn(min = 220.dp),
            shape = RoundedCornerShape(24.dp),
            color = Color.White.copy(alpha = 0.85f),
            tonalElevation = 6.dp,
            shadowElevation = 8.dp
        ) {
            Row(
                modifier = Modifier.padding(horizontal = 10.dp),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(10.dp)
            ) {
                MicButton(
                    isRecording = isRecording,
                    hasPermission = hasPermission,
                    isProcessing =
                    uiState !is MeetingProcessState.Idle &&
                        uiState !is MeetingProcessState.Success &&
                        uiState !is MeetingProcessState.Error &&
                        uiState !is MeetingProcessState.Recording,
                    onClick = onMicClick,
                    size = 36.dp,
                    modifier =
                    Modifier.graphicsLayer {
                        if (isRecording) {
                            scaleX = pulseScale
                            scaleY = pulseScale
                        }
                    }
                )

                if (!isRecording) {
                    IconButton(
                        onClick = onUploadClick,
                        modifier = Modifier.size(32.dp)
                    ) {
                        Icon(
                            imageVector = Icons.Default.FolderOpen,
                            contentDescription = "Upload Audio",
                            tint = MaterialTheme.colorScheme.primary,
                            modifier = Modifier.size(24.dp)
                        )
                    }
                }

                if (isRecording) {
                    IconButton(
                        onClick = onStopClick,
                        modifier = Modifier.size(32.dp)
                    ) {
                        Surface(
                            modifier = Modifier.size(16.dp),
                            color = Color(0xFFEF5350),
                            shape = RoundedCornerShape(2.dp)
                        ) {}
                    }
                }

                IconButton(
                    onClick = onClearClick,
                    modifier = Modifier.size(32.dp)
                ) {
                    Icon(
                        imageVector = Icons.Default.DeleteOutline,
                        contentDescription = "Clear",
                        tint = MaterialTheme.colorScheme.primary,
                        modifier = Modifier.size(20.dp)
                    )
                }

                Text(
                    text =
                    when {
                        isRecording -> "Live"

                        uiState !is MeetingProcessState.Idle &&
                            uiState !is MeetingProcessState.Success &&
                            uiState !is MeetingProcessState.Error -> "AI Active"

                        else -> "Ready"
                    },
                    fontSize = 12.sp,
                    color = if (isRecording) MaterialTheme.colorScheme.error else MaterialTheme.colorScheme.primary,
                    fontWeight = FontWeight.Black,
                    letterSpacing = 1.sp,
                    modifier = Modifier.padding(end = 4.dp)
                )
            }
        }
    }
}
