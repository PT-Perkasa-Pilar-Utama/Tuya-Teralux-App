package com.example.whisperandroid.presentation.components

import androidx.compose.animation.core.animateDpAsState
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.whisperandroid.ui.theme.SensioElevation
import com.example.whisperandroid.ui.theme.SensioRadius
import com.example.whisperandroid.ui.theme.SensioTypography

@Composable
fun SensioButton(
    text: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    enabled: Boolean = true,
    isLoading: Boolean = false
) {
    val elevation by animateDpAsState(
        targetValue = when {
            !enabled -> SensioElevation.None
            isLoading -> SensioElevation.Sm
            else -> SensioElevation.Md
        },
        label = "buttonElevation"
    )

    Button(
        onClick = onClick,
        enabled = enabled && !isLoading,
        shape = RoundedCornerShape(SensioRadius.Lg),
        colors =
        ButtonDefaults.buttonColors(
            containerColor = MaterialTheme.colorScheme.primary,
            contentColor = MaterialTheme.colorScheme.onPrimary,
            disabledContainerColor = MaterialTheme.colorScheme.outlineVariant.copy(alpha = 0.3f),
            disabledContentColor = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.4f)
        ),
        elevation =
        ButtonDefaults.buttonElevation(
            defaultElevation = elevation,
            pressedElevation = SensioElevation.Sm,
            disabledElevation = SensioElevation.None
        ),
        modifier =
        modifier
            .fillMaxWidth()
            .height(52.dp)
    ) {
        if (isLoading) {
            androidx.compose.material3.CircularProgressIndicator(
                modifier = Modifier.height(20.dp),
                color = MaterialTheme.colorScheme.onPrimary,
                strokeWidth = 2.dp
            )
        } else {
            Text(
                text = text,
                fontSize = SensioTypography.ButtonText,
                fontWeight = FontWeight.SemiBold,
                letterSpacing = 0.3.sp
            )
        }
    }
}
