package com.example.whisper_android.presentation.assistant

import androidx.compose.animation.core.*
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.scale
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

@Composable
fun AiMindVisual(
    isThinking: Boolean = false,
    modifier: Modifier = Modifier,
    size: androidx.compose.ui.unit.Dp = 64.dp
) {
    val infiniteTransition = rememberInfiniteTransition(label = "AiMindEngine")
    
    // Minimal Opacity Pulse (1.8s)
    val opacity by infiniteTransition.animateFloat(
        initialValue = if (isThinking) 0.85f else 0.96f,
        targetValue = 1f,
        animationSpec = infiniteRepeatable(
            animation = tween(if (isThinking) 900 else 1800, easing = LinearOutSlowInEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "EngineOpacity"
    )

    val scaleFactor = size / 64.dp

    Box(
        contentAlignment = Alignment.Center,
        modifier = modifier.size(size) 
    ) {
        Surface(
            modifier = Modifier
                .fillMaxSize()
                .alpha(opacity),
            shape = CircleShape,
            color = MaterialTheme.colorScheme.primary,
            tonalElevation = 2.dp,
            shadowElevation = if (size > 40.dp) 6.dp else 2.dp
        ) {
            Box(contentAlignment = Alignment.Center) {
                // Engine Signal (Minimalist Heartbeat)
                Row(
                    horizontalArrangement = Arrangement.spacedBy(3.dp * scaleFactor),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    repeat(3) { index ->
                        val barHeight by infiniteTransition.animateFloat(
                            initialValue = 12f * scaleFactor,
                            targetValue = if (isThinking) 24f * scaleFactor else 16f * scaleFactor,
                            animationSpec = infiniteRepeatable(
                                animation = tween(1200 + (index * 150), easing = FastOutSlowInEasing),
                                repeatMode = RepeatMode.Reverse
                            ),
                            label = "Signal$index"
                        )
                        Box(
                            modifier = Modifier
                                .width(3.dp * scaleFactor)
                                .height(barHeight.dp)
                                .background(Color.White.copy(alpha = 0.9f), RoundedCornerShape(1.dp * scaleFactor))
                        )
                    }
                }
            }
        }
    }
}

@Composable
fun SuggestedActionCard(
    title: String,
    subtitle: String,
    onClick: () -> Unit,
    enabled: Boolean = true,
    modifier: Modifier = Modifier
) {
    var isPressed by remember { mutableStateOf(false) }
    val scale by animateFloatAsState(
        targetValue = if (isPressed) 0.98f else 1f,
        animationSpec = spring(stiffness = Spring.StiffnessLow),
        label = "PressScale"
    )

    Surface(
        onClick = { if (enabled) onClick() },
        shape = RoundedCornerShape(16.dp),
        color = Color.White.copy(alpha = 0.95f),
        border = androidx.compose.foundation.BorderStroke(
            1.dp, 
            MaterialTheme.colorScheme.onSurface.copy(alpha = 0.05f)
        ),
        modifier = modifier
            .scale(scale)
            .alpha(if (enabled) 1f else 0.5f)
            .pointerInput(enabled) {
                if (enabled) {
                    detectTapGestures(
                        onPress = { 
                            isPressed = true
                            tryAwaitRelease()
                            isPressed = false
                        }
                    )
                }
            },
        shadowElevation = 2.dp,
        tonalElevation = 1.dp
    ) {
        Column(
            modifier = Modifier.padding(horizontal = 16.dp, vertical = 24.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            Text(
                text = title,
                style = MaterialTheme.typography.bodyLarge.copy(
                    fontSize = 16.sp, // Slightly bigger since icon is gone
                    fontWeight = FontWeight.SemiBold,
                    letterSpacing = (-0.3).sp
                ),
                color = MaterialTheme.colorScheme.onSurface,
                textAlign = TextAlign.Center
            )
            Spacer(modifier = Modifier.height(6.dp))
            Text(
                text = subtitle,
                style = MaterialTheme.typography.bodySmall.copy(
                    fontSize = 13.sp,
                    lineHeight = 18.sp
                ),
                color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.4f),
                textAlign = TextAlign.Center,
                maxLines = 2
            )
        }
    }
}

@Composable
fun TypingIndicator(
    modifier: Modifier = Modifier
) {
    val infiniteTransition = rememberInfiniteTransition(label = "Typing")
    
    Row(
        modifier = modifier.padding(horizontal = 20.dp, vertical = 14.dp),
        horizontalArrangement = Arrangement.spacedBy(6.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        repeat(3) { index ->
            val bounce by infiniteTransition.animateFloat(
                initialValue = 0f,
                targetValue = -8f,
                animationSpec = infiniteRepeatable(
                    animation = tween(400, delayMillis = index * 150, easing = FastOutSlowInEasing),
                    repeatMode = RepeatMode.Reverse
                ),
                label = "DotBounce$index"
            )
            
            val alpha by infiniteTransition.animateFloat(
                initialValue = 0.4f,
                targetValue = 1f,
                animationSpec = infiniteRepeatable(
                    animation = tween(400, delayMillis = index * 150),
                    repeatMode = RepeatMode.Reverse
                ),
                label = "DotAlpha$index"
            )

            Box(
                modifier = Modifier
                    .size(8.dp)
                    .graphicsLayer {
                        translationY = bounce.dp.toPx()
                    }
                    .alpha(alpha)
                    .background(
                        brush = Brush.verticalGradient(
                            colors = listOf(
                                MaterialTheme.colorScheme.primary,
                                MaterialTheme.colorScheme.primary.copy(alpha = 0.7f)
                            )
                        ),
                        shape = CircleShape
                    )
            )
        }
    }
}
