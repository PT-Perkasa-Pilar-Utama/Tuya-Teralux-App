package com.example.whisper_android.presentation.edge

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
fun EdgeComputingScreen(
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
                        text = "Offline Voice Transcription & On-Device Processing",
                        style = MaterialTheme.typography.bodyMedium,
                        modifier = Modifier.padding(bottom = 16.dp)
                    )

                    Text(
                        text = "🔄 Flow Summary (Edge Computing)",
                        style = MaterialTheme.typography.titleMedium,
                        color = MaterialTheme.colorScheme.primary
                    )
                    Text(
                        text = "1. The Android app (Kotlin) captures microphone input.\n" +
                                "2. Audio is processed directly on the device.\n" +
                                "3. whisper.cpp runs locally within the app.\n" +
                                "4. Speech is converted to text on-device.\n" +
                                "5. The transcribed text is returned instantly within the app.",
                        style = MaterialTheme.typography.bodyMedium,
                        modifier = Modifier.padding(bottom = 16.dp)
                    )

                    Text(
                        text = "✅ Advantages",
                        style = MaterialTheme.typography.titleMedium,
                        color = MaterialTheme.colorScheme.primary
                    )
                    Text(
                        text = "• Works offline (no internet required).\n" +
                                "• Lowest possible latency.\n" +
                                "• No server infrastructure cost.\n" +
                                "• Better privacy (audio never leaves the device).\n" +
                                "• No backend dependency.",
                        style = MaterialTheme.typography.bodyMedium,
                        modifier = Modifier.padding(bottom = 16.dp)
                    )

                    Text(
                        text = "❌ Disadvantages",
                        style = MaterialTheme.typography.titleMedium,
                        color = MaterialTheme.colorScheme.error
                    )
                    Text(
                        text = "• Heavy CPU and memory usage on the device.\n" +
                                "• Limited by mobile hardware performance.\n" +
                                "• Increases app size (model bundled locally).\n" +
                                "• Harder to update or upgrade the model.\n" +
                                "• Battery consumption can be high.",
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
                text = "Edge Computing - Hello World",
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
