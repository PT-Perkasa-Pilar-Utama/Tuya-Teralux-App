package com.example.whisper_android.presentation.streaming

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.getValue
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.verticalScroll
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.layout.padding

@Composable
fun StreamingScreen(
    onNavigateBack: () -> Unit
) {
    var showWalkthrough by remember { mutableStateOf(true) }

    if (showWalkthrough) {
        AlertDialog(
            onDismissRequest = { showWalkthrough = false },
            title = {
                Text(
                    text = "Walkthrough",
                    style = MaterialTheme.typography.headlineSmall
                )
            },
            text = {
                Column(
                    modifier = Modifier.verticalScroll(rememberScrollState())
                ) {
                    Text(
                        text = "🎯 Purpose",
                        style = MaterialTheme.typography.titleMedium,
                        color = MaterialTheme.colorScheme.primary
                    )
                    Text(
                        text = "Real-Time Voice Assistant & Live Voice Control",
                        style = MaterialTheme.typography.bodyMedium,
                        modifier = Modifier.padding(bottom = 16.dp)
                    )

                    Text(
                        text = "🔄 Flow Summary (Streaming)",
                        style = MaterialTheme.typography.titleMedium,
                        color = MaterialTheme.colorScheme.primary
                    )
                    Text(
                        text = "1. The Android app (Kotlin) captures microphone input in real time.\n" +
                                "2. Audio is chunked and encoded into small segments.\n" +
                                "3. The app publishes audio chunks to a message broker topic.\n" +
                                "4. The backend (Golang) subscribes to the topic.\n" +
                                "5. The backend decodes and reassembles the audio stream.\n" +
                                "6. The backend processes the stream using whisper.cpp.\n" +
                                "7. The system returns near real-time transcribed text.",
                        style = MaterialTheme.typography.bodyMedium,
                        modifier = Modifier.padding(bottom = 16.dp)
                    )

                    Text(
                        text = "✅ Advantages",
                        style = MaterialTheme.typography.titleMedium,
                        color = MaterialTheme.colorScheme.primary
                    )
                    Text(
                        text = "• Near real-time transcription.\n" +
                                "• Lower perceived latency.\n" +
                                "• Suitable for interactive voice commands.\n" +
                                "• Scalable via Pub/Sub architecture.\n" +
                                "• Supports distributed backend consumers.",
                        style = MaterialTheme.typography.bodyMedium,
                        modifier = Modifier.padding(bottom = 16.dp)
                    )

                    Text(
                        text = "❌ Disadvantages",
                        style = MaterialTheme.typography.titleMedium,
                        color = MaterialTheme.colorScheme.error
                    )
                    Text(
                        text = "• More complex architecture (broker, chunking, reassembly).\n" +
                                "• Higher implementation complexity.\n" +
                                "• Requires message ordering and reliability handling.\n" +
                                "• Higher infrastructure cost.\n" +
                                "• Requires stable internet connection.",
                        style = MaterialTheme.typography.bodyMedium
                    )
                }
            },
            confirmButton = {
                Button(onClick = { showWalkthrough = false }) {
                    Text("Close")
                }
            }
        )
    }

    Scaffold { paddingValues ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(paddingValues),
            contentAlignment = Alignment.Center
        ) {
            Text(
                text = "Realtime Streaming - Hello World",
                style = MaterialTheme.typography.headlineMedium
            )

            Button(
                onClick = onNavigateBack,
                modifier = Modifier
                    .align(Alignment.BottomCenter)
                    .padding(16.dp)
            ) {
                Text("Back to Dashboard")
            }
        }
    }
}
