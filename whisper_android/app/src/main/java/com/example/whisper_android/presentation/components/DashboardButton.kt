package com.example.whisper_android.presentation.components

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ChevronRight
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp

@Composable
fun DashboardButton(
    text: String,
    subtitle: String,
    icon: ImageVector,
    color: Color,
    onClick: () -> Unit
) {
    ElevatedButton(
        onClick = onClick,
        modifier = Modifier
            .fillMaxWidth()
            .height(100.dp),
        shape = RoundedCornerShape(24.dp),
        colors = ButtonDefaults.elevatedButtonColors(
            containerColor = MaterialTheme.colorScheme.surface,
            contentColor = color
        ),
        elevation = ButtonDefaults.elevatedButtonElevation(
            defaultElevation = 2.dp,
            pressedElevation = 6.dp
        ),
        contentPadding = PaddingValues(horizontal = 24.dp)
    ) {
        Row(
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.Start,
            modifier = Modifier.fillMaxWidth()
        ) {
            Box(
                modifier = Modifier
                    .size(52.dp)
                    .background(
                        color.copy(alpha = 0.1f),
                        RoundedCornerShape(16.dp)
                    ),
                contentAlignment = Alignment.Center
            ) {
                Icon(
                    imageVector = icon,
                    contentDescription = null,
                    modifier = Modifier.size(28.dp),
                    tint = color
                )
            }
            Spacer(modifier = Modifier.size(20.dp))
            Column {
                Text(
                    text = text,
                    style = MaterialTheme.typography.titleLarge.copy(
                        fontWeight = FontWeight.Bold
                    ),
                    color = MaterialTheme.colorScheme.onSurface
                )
                Text(
                    text = subtitle,
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
            Spacer(modifier = Modifier.weight(1f))
            Icon(
                imageVector = Icons.Default.ChevronRight,
                contentDescription = null,
                modifier = Modifier.size(20.dp),
                tint = MaterialTheme.colorScheme.outline
            )
        }
    }
}
