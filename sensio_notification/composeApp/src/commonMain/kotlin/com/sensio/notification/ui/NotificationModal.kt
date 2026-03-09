package com.sensio.app.notif.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.Button
import androidx.compose.material.ButtonDefaults
import androidx.compose.material.Icon
import androidx.compose.material.Surface
import androidx.compose.material.Text
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Notifications
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.sensio.app.notif.ui.theme.Cyan400
import com.sensio.app.notif.ui.theme.Cyan500
import com.sensio.app.notif.ui.theme.Cyan600
import com.sensio.app.notif.ui.theme.Slate400
import com.sensio.app.notif.ui.theme.Slate50
import com.sensio.app.notif.ui.theme.Slate900

@Composable
@Suppress("FunctionName")
fun NotificationModal(
    title: String,
    message: String,
    onDismiss: () -> Unit
) {
    Surface(
        modifier =
        Modifier
            .widthIn(max = 340.dp)
            .padding(24.dp)
            .shadow(24.dp, RoundedCornerShape(32.dp)),
        shape = RoundedCornerShape(32.dp),
        // Matching sensio_app dark surface
        color = Slate900
    ) {
        Column(
            modifier = Modifier.padding(24.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            // Sensio Logo Section
            Row(
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(10.dp)
            ) {
                Box(
                    modifier =
                    Modifier
                        .size(42.dp)
                        .background(
                            brush =
                            Brush.linearGradient(
                                colors =
                                listOf(
                                    Cyan400,
                                    Cyan600
                                )
                            ),
                            shape = CircleShape
                        ),
                    contentAlignment = Alignment.Center
                ) {
                    Icon(
                        Icons.Default.Notifications,
                        contentDescription = null,
                        tint = Color.White,
                        modifier = Modifier.size(24.dp)
                    )
                }

                Text(
                    text = "Sensio",
                    fontSize = 28.sp,
                    fontWeight = FontWeight.Black,
                    color = Cyan500,
                    letterSpacing = (-0.5).sp
                )
            }

            Spacer(Modifier.height(24.dp))

            Text(
                text = title,
                fontWeight = FontWeight.ExtraBold,
                fontSize = 22.sp,
                color = Slate50,
                textAlign = TextAlign.Center
            )

            Text(
                text = message,
                fontSize = 16.sp,
                color = Slate400,
                textAlign = TextAlign.Center,
                modifier = Modifier.padding(top = 8.dp),
                lineHeight = 22.sp
            )

            Spacer(Modifier.height(32.dp))

            Button(
                onClick = onDismiss,
                modifier = Modifier.fillMaxWidth().height(56.dp),
                colors =
                ButtonDefaults.buttonColors(
                    backgroundColor = Cyan600,
                    contentColor = Color.White
                ),
                shape = RoundedCornerShape(16.dp),
                elevation =
                ButtonDefaults.elevation(
                    defaultElevation = 0.dp,
                    pressedElevation = 2.dp
                )
            ) {
                Text(
                    "Close",
                    fontWeight = FontWeight.Bold,
                    fontSize = 16.sp
                )
            }
        }
    }
}
