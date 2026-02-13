package com.example.whisper_android.presentation.components

import androidx.compose.foundation.Canvas
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Path
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.StrokeJoin
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

@Composable
fun WhisperLogo(
    modifier: Modifier = Modifier,
    showText: Boolean = true
) {
    Row(
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(12.dp),
        modifier = modifier
    ) {
        WhisperWIcon(modifier = Modifier.size(48.dp))

        if (showText) {
            Text(
                text = "Whisper",
                fontSize = 32.sp,
                fontWeight = FontWeight.Bold,
                color = Color(0xFF06B6D4), // Primary Cyan color
                style = MaterialTheme.typography.headlineLarge
            )
        }
    }
}

@Composable
fun WhisperWIcon(modifier: Modifier = Modifier) {
    Canvas(modifier = modifier) {
        val width = size.width
        val height = size.height
        val strokeWidth = 8.dp.toPx()
        
        val path = Path().apply {
            moveTo(strokeWidth, height * 0.2f)
            lineTo(width * 0.3f, height * 0.8f)
            lineTo(width * 0.5f, height * 0.4f)
            lineTo(width * 0.7f, height * 0.8f)
            lineTo(width - strokeWidth, height * 0.2f)
        }
        
        drawPath(
            path = path,
            brush = Brush.linearGradient(
                colors = listOf(
                    Color(0xFF22D3EE), // Cyan 400
                    Color(0xFF06B6D4), // Cyan 500
                    Color(0xFF0891B2)  // Cyan 600 for more depth
                )
            ),
            style = Stroke(
                width = strokeWidth,
                cap = StrokeCap.Round,
                join = StrokeJoin.Round
            )
        )
    }
}
