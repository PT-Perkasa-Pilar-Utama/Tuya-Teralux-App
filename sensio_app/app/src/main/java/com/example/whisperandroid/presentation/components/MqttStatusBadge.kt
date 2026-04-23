package com.example.whisperandroid.presentation.components

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.whisperandroid.utils.MqttHelper

@Composable
fun MqttStatusBadge(
    status: MqttHelper.MqttConnectionStatus,
    onReconnectClick: () -> Unit = {},
    modifier: Modifier = Modifier
) {
    val isError = status == MqttHelper.MqttConnectionStatus.DISCONNECTED ||
        status == MqttHelper.MqttConnectionStatus.FAILED ||
        status == MqttHelper.MqttConnectionStatus.NO_INTERNET

    val color =
        when (status) {
            MqttHelper.MqttConnectionStatus.CONNECTED -> Color(0xFF4CAF50)
            MqttHelper.MqttConnectionStatus.CONNECTING -> Color(0xFFFFC107)
            MqttHelper.MqttConnectionStatus.DISCONNECTED -> Color(0xFFF44336)
            MqttHelper.MqttConnectionStatus.FAILED -> Color(0xFFD32F2F)
            MqttHelper.MqttConnectionStatus.NO_INTERNET -> Color(0xFF9C27B0)
        }

    val text =
        when (status) {
            MqttHelper.MqttConnectionStatus.CONNECTED -> "Online"
            MqttHelper.MqttConnectionStatus.CONNECTING -> "Connecting"
            MqttHelper.MqttConnectionStatus.DISCONNECTED -> "Offline"
            MqttHelper.MqttConnectionStatus.FAILED -> "Error"
            MqttHelper.MqttConnectionStatus.NO_INTERNET -> "No Internet"
        }

    Row(
        verticalAlignment = Alignment.CenterVertically,
        modifier = modifier
            .padding(start = 4.dp)
            .background(color.copy(alpha = 0.1f), RoundedCornerShape(12.dp))
            .then(
                if (isError) {
                    Modifier.clickable { onReconnectClick() }
                } else {
                    Modifier
                }
            )
            .padding(horizontal = 8.dp, vertical = 4.dp)
    ) {
        Box(
            modifier = Modifier
                .size(8.dp)
                .background(color, CircleShape)
        )
        Spacer(modifier = Modifier.width(6.dp))
        Text(
            text = text,
            fontSize = 11.sp,
            fontWeight = FontWeight.Bold,
            color = color
        )
        if (isError) {
            Spacer(modifier = Modifier.width(4.dp))
            Icon(
                imageVector = Icons.Default.Refresh,
                contentDescription = "Reconnect",
                tint = color,
                modifier = Modifier.size(12.dp)
            )
        }
    }
}
