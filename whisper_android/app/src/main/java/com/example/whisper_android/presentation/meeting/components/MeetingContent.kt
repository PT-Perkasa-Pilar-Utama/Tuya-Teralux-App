package com.example.whisper_android.presentation.meeting.components

import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.whisper_android.domain.usecase.MeetingProcessState
import dev.jeziellago.compose.markdowntext.MarkdownText

@Composable
fun MeetingIdleContent() {
    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center,
    ) {
        Row(
            modifier = Modifier.height(40.dp),
            horizontalArrangement = Arrangement.spacedBy(4.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            repeat(5) { index ->
                Box(
                    modifier =
                        Modifier
                            .width(4.dp)
                            .height(if (index % 2 == 0) 24.dp else 16.dp)
                            .background(
                                MaterialTheme.colorScheme.primary.copy(alpha = 0.2f),
                                RoundedCornerShape(2.dp),
                            ),
                )
            }
        }
        Spacer(modifier = Modifier.height(24.dp))
        Text(
            text = "Ready to capture your next breakthrough.",
            style = MaterialTheme.typography.titleMedium,
            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.4f),
            textAlign = TextAlign.Center,
            fontWeight = FontWeight.Medium,
        )
        Spacer(modifier = Modifier.height(8.dp))
        Text(
            text = "Tap the mic to start recording your meeting.",
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.3f),
            textAlign = TextAlign.Center,
        )
    }
}

@Composable
fun MeetingRecordingContent() {
    Text(
        text = "Recording...",
        style = MaterialTheme.typography.headlineSmall,
        color = MaterialTheme.colorScheme.error,
        fontWeight = FontWeight.Bold,
        textAlign = TextAlign.Center,
    )
}

@Composable
fun MeetingSuccessContent(state: MeetingProcessState.Success) {
    Column(
        modifier =
            Modifier
                .fillMaxSize()
                .verticalScroll(rememberScrollState()),
    ) {
        Text(
            text = "Meeting Summary",
            style = MaterialTheme.typography.titleLarge,
            color = MaterialTheme.colorScheme.primary,
            fontWeight = FontWeight.Bold,
            modifier = Modifier.padding(bottom = 2.dp),
        )

        MarkdownText(
            markdown =
                state.summary
                    .replace(Regex("^-+\\s*$", RegexOption.MULTILINE), "")
                    .replace(Regex("^.*–.*$", RegexOption.MULTILINE), "")
                    .replace("\n\n\n", "\n\n")
                    .replace(Regex("\n{3,}"), "\n\n")
                    .trim(),
            style =
                MaterialTheme.typography.bodyLarge.copy(
                    color = Color.DarkGray,
                    fontSize = 13.sp,
                    lineHeight = 16.sp,
                ),
            modifier = Modifier.fillMaxWidth(),
        )
        Spacer(modifier = Modifier.height(16.dp))
    }
}

@Composable
fun MeetingErrorContent(state: MeetingProcessState.Error) {
    Text(
        text = "Error: ${state.message}",
        style = MaterialTheme.typography.bodyLarge,
        color = MaterialTheme.colorScheme.error,
        textAlign = TextAlign.Center,
    )
}

@Composable
fun MeetingLoadingContent(
    uiState: MeetingProcessState,
    glowAlpha: Float,
) {
    val primaryColor = MaterialTheme.colorScheme.primary
    Column(horizontalAlignment = Alignment.CenterHorizontally) {
        Box(contentAlignment = Alignment.Center) {
            Canvas(modifier = Modifier.size(100.dp)) {
                drawCircle(
                    color = primaryColor,
                    alpha = glowAlpha * 0.15f,
                    radius = size.minDimension / 2,
                )
            }
            CircularProgressIndicator(
                color = primaryColor,
                strokeWidth = 3.dp,
                modifier = Modifier.size(56.dp),
            )
        }
        Spacer(modifier = Modifier.height(8.dp))
        Text(
            text =
                when (uiState) {
                    is MeetingProcessState.Uploading -> "Securely Uploading..."
                    is MeetingProcessState.Transcribing -> "AI Transcribing..."
                    is MeetingProcessState.Translating -> "Translating Context..."
                    is MeetingProcessState.Summarizing -> "Generating Insights..."
                    else -> "Thinking..."
                },
            style = MaterialTheme.typography.titleMedium,
            color = primaryColor,
            fontWeight = FontWeight.Black,
            letterSpacing = 0.5.sp,
        )
    }
}
