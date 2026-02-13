package com.example.whisper_android.presentation.components

import androidx.compose.animation.core.*
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.combinedClickable
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Mic
import androidx.compose.material.icons.filled.Pause
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.scale
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp

@Composable
fun MicButton(
    isRecording: Boolean,
    hasPermission: Boolean,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    isProcessing: Boolean = false,
    isPaused: Boolean = false,
    size: Dp = 120.dp
) {
    // --- Pulse Animation for Active States ---
    val infiniteTransition = rememberInfiniteTransition(label = "Pulse")
    val pulseScale by infiniteTransition.animateFloat(
        initialValue = 1f,
        targetValue = 1.35f,
        animationSpec = infiniteRepeatable(
            animation = tween(1200, easing = FastOutSlowInEasing),
            repeatMode = RepeatMode.Restart
        ),
        label = "PulseScale"
    )
    val pulseAlpha by infiniteTransition.animateFloat(
        initialValue = 0.5f,
        targetValue = 0f,
        animationSpec = infiniteRepeatable(
            animation = tween(1200, easing = FastOutSlowInEasing),
            repeatMode = RepeatMode.Restart
        ),
        label = "PulseAlpha"
    )

    // --- Color Selection (Solid as requested) ---
    val buttonColor = remember(isRecording, isPaused, isProcessing, hasPermission) {
        when {
            !hasPermission -> Color.Gray
            isProcessing -> Color(0xFFFF9800) // Orange: Thinking
            isPaused -> Color(0xFFFF9800)     // Orange: Paused
            isRecording -> Color(0xFFEF5350)  // Red: Recording
            else -> Color(0xFF06B6D4)        // Teal/Cyan: Idle/Active
        }
    }

    Box(
        contentAlignment = Alignment.Center,
        modifier = modifier
    ) {
        // Outer Pulse Layer
        if ((isRecording && !isPaused) || isProcessing) {
            Box(
                modifier = Modifier
                    .size(size * 1.25f)
                    .scale(pulseScale)
                    .alpha(pulseAlpha)
                    .background(buttonColor.copy(alpha = 0.3f), CircleShape)
            )
            
            // Static Glow Layer for High Focus (Red Glow)
            if (isRecording && !isPaused) {
                Box(
                    modifier = Modifier
                        .size(size * 1.2f)
                        .background(Color(0xFFEF5350).copy(alpha = 0.2f), CircleShape)
                )
                Box(
                    modifier = Modifier
                        .size(size * 1.1f)
                        .background(Color(0xFFEF5350).copy(alpha = 0.15f), CircleShape)
                )
            }
        }

        // Main Surface Logic (No gradients, just solid)
        Surface(
            onClick = onClick,
            enabled = hasPermission && !isProcessing,
            shape = CircleShape,
            color = buttonColor,
            shadowElevation = if (isRecording || isProcessing) 12.dp else 4.dp,
            modifier = Modifier.size(size)
        ) {
            Box(contentAlignment = Alignment.Center) {
                Icon(
                    imageVector = when {
                        isProcessing -> Icons.Default.PlayArrow // Simple play/forward for thinking
                        isPaused -> Icons.Default.PlayArrow     // Play to resume
                        else -> Icons.Default.Mic
                    },
                    contentDescription = "Microphone",
                    tint = Color.White,
                    modifier = Modifier.size(size / 2.3f)
                )
            }
        }
    }
}
