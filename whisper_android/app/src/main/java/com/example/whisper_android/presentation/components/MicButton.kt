package com.example.whisper_android.presentation.components

import androidx.compose.animation.core.*
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Mic
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.scale
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp

@Composable
fun MicButton(
    isRecording: Boolean,
    hasPermission: Boolean,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    isProcessing: Boolean = false, // Added isProcessing
    size: Dp = 120.dp
) {
    // --- Animation Logic ---
    val infiniteTransition = rememberInfiniteTransition(label = "Pulse")
    val pulseScale by infiniteTransition.animateFloat(
        initialValue = 1f,
        targetValue = 1.4f,
        animationSpec = infiniteRepeatable(
            animation = tween(1000, easing = FastOutSlowInEasing),
            repeatMode = RepeatMode.Restart
        ),
        label = "PulseScale"
    )
    val pulseAlpha by infiniteTransition.animateFloat(
        initialValue = 0.6f,
        targetValue = 0f,
        animationSpec = infiniteRepeatable(
            animation = tween(1000, easing = FastOutSlowInEasing),
            repeatMode = RepeatMode.Restart
        ),
        label = "PulseAlpha"
    )

    Box(
        contentAlignment = Alignment.Center,
        modifier = modifier
    ) {
        // Pulse layer
        if (isRecording || isProcessing) { // Pulse also when processing
            val pulseColor = if (isProcessing) Color(0xFFFF9800) else MaterialTheme.colorScheme.error
            Box(
                modifier = Modifier
                    .size(size * 1.2f)
                    .scale(pulseScale)
                    .alpha(pulseAlpha)
                    .background(pulseColor.copy(alpha = 0.3f), CircleShape)
            )
        }
        
        // Main Button Surface
        Surface(
            onClick = {
                if (hasPermission) {
                    onClick()
                }
            },
            enabled = hasPermission && !isProcessing, // Disable click while processing
            shape = CircleShape,
            color = when {
                !hasPermission -> Color.LightGray
                isProcessing -> Color(0xFFFF9800) // Orange for "Thinking"
                isRecording -> MaterialTheme.colorScheme.error
                else -> Color(0xFF4CAF50) // Material Green 500
            },
            shadowElevation = if (isRecording || isProcessing) 12.dp else 4.dp,
            modifier = Modifier.size(size)
        ) {
            Box(contentAlignment = Alignment.Center) {
                Icon(
                    imageVector = Icons.Default.Mic,
                    contentDescription = "Microphone",
                    tint = Color.White,
                    modifier = Modifier.size(size / 2)
                )
            }
        }
    }
}
