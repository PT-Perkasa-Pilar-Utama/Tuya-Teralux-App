package com.example.whisper_android.presentation.edge

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
fun EdgeComputingScreen(
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
        title = "Edge Computing",
        onNavigateBack = onNavigateBack,
        isRecording = isRecording,
        isProcessing = isProcessing,
        hasPermission = hasPermission,
        transcriptionResults = transcriptionResults,
        onMicClick = {
            if (!isRecording && !isProcessing) {
                isRecording = true
            } else if (isRecording) {
                isRecording = false
                isProcessing = true
                scope.launch {
                    delay(800) // Edge is faster (?) simulation
                    val mockUserText = "Status baterai perangkat."
                    val mockAssistantText = "Baterai perangkat IoT anda berada di level 85%. Aman bro!"
                    
                    transcriptionResults = transcriptionResults + 
                        TranscriptionMessage(mockUserText, MessageRole.USER)
                    
                    delay(400)
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
                text = "Offline Voice Transcription & On-Device Processing",
                style = MaterialTheme.typography.bodyMedium
            )

            Text(
                text = "🔄 Flow Summary (Edge Computing)",
                style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                color = MaterialTheme.colorScheme.primary
            )
            Text(
                text = "1. Microphone input is captured.\n" +
                        "2. whisper.cpp runs locally within the app.\n" +
                        "3. Audio is processed directly on the device.\n" +
                        "4. Results are returned instantly offline.",
                style = MaterialTheme.typography.bodyMedium
            )

            Text(
                text = "✅ Advantages",
                style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                color = MaterialTheme.colorScheme.primary
            )
            Text(
                text = "• Works offline (no internet required).\n" +
                        "• Lowest possible latency.\n" +
                        "• Superior privacy (audio stays on device).\n" +
                        "• No server infrastructure costs.",
                style = MaterialTheme.typography.bodyMedium
            )

            Text(
                text = "❌ Disadvantages",
                style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                color = MaterialTheme.colorScheme.error
            )
            Text(
                text = "• High CPU and battery usage.\n" +
                        "• Limited by mobile hardware performance.\n" +
                        "• Increases app size (bundled model).\n" +
                        "• Harder to update/upgrade the model.",
                style = MaterialTheme.typography.bodyMedium
            )
        }
    )
}
