package com.example.whisper_android.presentation.streaming

import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.core.content.ContextCompat
import android.Manifest
import android.content.pm.PackageManager
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

import com.example.whisper_android.presentation.components.FeatureScreenTemplate
import com.example.whisper_android.presentation.components.MessageRole
import com.example.whisper_android.presentation.components.TranscriptionMessage

@Composable
fun StreamingScreen(
    onNavigateBack: () -> Unit
) {
    var isRecording by remember { mutableStateOf(false) }
    var isProcessing by remember { mutableStateOf(false) }
    var transcriptionResults by remember { mutableStateOf(listOf<TranscriptionMessage>()) }
    
    val context = LocalContext.current
    val scope = rememberCoroutineScope()
    
    val hasPermission = ContextCompat.checkSelfPermission(
        context,
        Manifest.permission.RECORD_AUDIO
    ) == PackageManager.PERMISSION_GRANTED

    FeatureScreenTemplate(
        title = "Realtime Streaming",
        onNavigateBack = onNavigateBack,
        isRecording = isRecording,
        isProcessing = isProcessing,
        hasPermission = hasPermission,
        transcriptionResults = transcriptionResults,
        onMicClick = {
            if (!isRecording && !isProcessing) {
                isRecording = true
                // In streaming, we might simulate periodic results, 
                // but for now, let's keep it similar to Upload for consistency
            } else if (isRecording) {
                isRecording = false
                isProcessing = true
                scope.launch {
                    delay(1200)
                    val mockUserText = "Turn off the living room lights."
                    val mockAssistantText = "Sure! The living room lights have been turned off."
                    
                    transcriptionResults = transcriptionResults + 
                        TranscriptionMessage(mockUserText, MessageRole.USER)
                    
                    delay(600)
                    transcriptionResults = transcriptionResults + 
                        TranscriptionMessage(mockAssistantText, MessageRole.ASSISTANT)
                    
                    isProcessing = false
                }
            }
        },
        onClearChat = { transcriptionResults = emptyList() },
        walkthroughContent = {
            Text(
                text = "🎯 Purpose",
                style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                color = MaterialTheme.colorScheme.primary
            )
            Text(
                text = "Real-Time Voice Assistant & Live Voice Control",
                style = MaterialTheme.typography.bodyMedium
            )

            Text(
                text = "🔄 Flow Summary (Streaming)",
                style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                color = MaterialTheme.colorScheme.primary
            )
            Text(
                text = "1. App captures microphone input in real time.\n" +
                        "2. Audio is chunked and sent to message broker.\n" +
                        "3. Backend subscribes and reassembles segments.\n" +
                        "4. Backend processes stream via whisper.cpp.\n" +
                        "5. System returns near real-time text.",
                style = MaterialTheme.typography.bodyMedium
            )

            Text(
                text = "✅ Advantages",
                style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                color = MaterialTheme.colorScheme.primary
            )
            Text(
                text = "• Near real-time transcription.\n" +
                        "• Suitable for interactive voice commands.\n" +
                        "• Scalable via Pub/Sub architecture.\n" +
                        "• Lower perceived latency for users.",
                style = MaterialTheme.typography.bodyMedium
            )

            Text(
                text = "❌ Disadvantages",
                style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                color = MaterialTheme.colorScheme.error
            )
            Text(
                text = "• Complex architecture (broker, chunking).\n" +
                        "• Higher infrastructure & implementation cost.\n" +
                        "• Requires message ordering/reliability.\n" +
                        "• Requires stable internet connection.",
                style = MaterialTheme.typography.bodyMedium
            )
        }
    )
}
