package com.example.whisperandroid.presentation.meeting.components

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.example.whisperandroid.domain.usecase.MeetingProcessState
import com.example.whisperandroid.presentation.components.AnimatedEmailButton
import com.example.whisperandroid.presentation.components.UiState

@Composable
fun MeetingHeaderControls(
    uiState: MeetingProcessState,
    emailState: UiState<Boolean>,
    onEmailClick: () -> Unit
) {
    val isEmailSending = emailState is UiState.Loading

    Row(
        modifier =
        Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 8.dp),
        horizontalArrangement = Arrangement.Start,
        verticalAlignment = Alignment.CenterVertically
    ) {
        if (uiState is MeetingProcessState.Success) {
            AnimatedEmailButton(
                isEmailSending = isEmailSending,
                onClick = onEmailClick
            )
        }
    }
}

// Removed duplicate MqttStatusBadge
