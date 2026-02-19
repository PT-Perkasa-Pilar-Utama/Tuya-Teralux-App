package com.example.whisper_android.presentation.components

import androidx.compose.animation.animateColorAsState
import androidx.compose.animation.core.tween
import androidx.compose.foundation.Indication
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

@Composable
fun LanguagePillToggle(
    selectedLanguage: String,
    onLanguageSelected: (String) -> Unit,
) {
    val languages = listOf("ID" to "id", "EN" to "en")
    val primaryColor = MaterialTheme.colorScheme.primary

    Row(
        modifier =
            Modifier
                .clip(RoundedCornerShape(16.dp))
                .background(Color.Black.copy(alpha = 0.05f))
                .padding(4.dp),
        horizontalArrangement = Arrangement.spacedBy(4.dp),
    ) {
        languages.forEach { (label, code) ->
            val isSelected = selectedLanguage == code
            val backgroundColor by animateColorAsState(
                targetValue = if (isSelected) primaryColor else Color.Transparent,
                animationSpec = tween(300),
            )
            val textColor by animateColorAsState(
                targetValue = if (isSelected) Color.White else primaryColor,
                animationSpec = tween(300),
            )

            Box(
                modifier =
                    Modifier
                        .clip(RoundedCornerShape(12.dp))
                        .background(backgroundColor)
                        .then(
                            if (!isSelected) {
                                Modifier.border(1.dp, primaryColor.copy(alpha = 0.5f), RoundedCornerShape(12.dp))
                            } else {
                                Modifier
                            },
                        ).clickable(
                            interactionSource = remember { MutableInteractionSource() },
                            indication = null,
                        ) { onLanguageSelected(code) }
                        .padding(horizontal = 14.dp, vertical = 6.dp),
                contentAlignment = Alignment.Center,
            ) {
                Text(
                    text = label,
                    style =
                        MaterialTheme.typography.labelLarge.copy(
                            fontWeight = if (isSelected) FontWeight.Bold else FontWeight.Medium,
                            letterSpacing = 0.5.sp,
                        ),
                    color = textColor,
                )
            }
        }
    }
}
