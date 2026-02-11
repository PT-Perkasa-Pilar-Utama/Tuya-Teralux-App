package com.example.whisper_android.presentation.upload

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
fun UploadScreen(
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
                        text = "Long-Form Meeting & Audio Transcription (30 minutes to 4+ hours)",
                        style = MaterialTheme.typography.bodyMedium,
                        modifier = Modifier.padding(bottom = 16.dp)
                    )

                    Text(
                        text = "🔄 Flow Summary (Normal Upload)",
                        style = MaterialTheme.typography.titleMedium,
                        color = MaterialTheme.colorScheme.primary
                    )
                    Text(
                        text = "1. The Android app (Kotlin) records audio using the microphone.\n" +
                                "2. The audio is saved as a complete audio file.\n" +
                                "3. The app sends the file to the backend (Golang) via REST API.\n" +
                                "4. The backend receives the audio file (mp3, wav, m4a, aac, flac).\n" +
                                "5. The backend processes the file using whisper.cpp.\n" +
                                "6. The system returns the transcribed text.",
                        style = MaterialTheme.typography.bodyMedium,
                        modifier = Modifier.padding(bottom = 16.dp)
                    )

                    Text(
                        text = "✅ Advantages",
                        style = MaterialTheme.typography.titleMedium,
                        color = MaterialTheme.colorScheme.primary
                    )
                    Text(
                        text = "• Simple architecture.\n" +
                                "• Easy to implement and debug.\n" +
                                "• No message broker required.\n" +
                                "• Stable for long recordings.\n" +
                                "• Full context improves transcription consistency.\n" +
                                "• Lower operational complexity.",
                        style = MaterialTheme.typography.bodyMedium,
                        modifier = Modifier.padding(bottom = 16.dp)
                    )

                    Text(
                        text = "❌ Disadvantages",
                        style = MaterialTheme.typography.titleMedium,
                        color = MaterialTheme.colorScheme.error
                    )
                    Text(
                        text = "• Not real-time (must wait until recording finishes).\n" +
                                "• Higher perceived latency for long meetings.\n" +
                                "• Large file upload required before processing.\n" +
                                "• May require background processing for very long sessions.\n" +
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
                text = "Upload Files - Hello World",
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
