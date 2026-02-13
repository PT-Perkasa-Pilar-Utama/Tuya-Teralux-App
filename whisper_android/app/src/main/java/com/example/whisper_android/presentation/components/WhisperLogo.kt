package com.example.whisper_android.presentation.components

import androidx.compose.foundation.Image
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.whisper_android.R

@Composable
fun WhisperLogo(
    modifier: Modifier = Modifier,
    showText: Boolean = true
) {
    Row(
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(8.dp),
        modifier = modifier
    ) {
        Image(
            painter = painterResource(id = R.drawable.sensio_icon),
            contentDescription = "Whisper Logo",
            modifier = Modifier.size(40.dp),
            colorFilter = androidx.compose.ui.graphics.ColorFilter.tint(MaterialTheme.colorScheme.primary)
        )

        Text(
            text = "Whisper",
            fontSize = 32.sp,
            fontWeight = FontWeight.Black,
            color = MaterialTheme.colorScheme.primary,
            style = MaterialTheme.typography.headlineLarge.copy(
                shadow = androidx.compose.ui.graphics.Shadow(
                    color = MaterialTheme.colorScheme.primary.copy(alpha = 0.2f),
                    offset = androidx.compose.ui.geometry.Offset(1f, 1f),
                    blurRadius = 4f
                )
            )
        )
    }
}
