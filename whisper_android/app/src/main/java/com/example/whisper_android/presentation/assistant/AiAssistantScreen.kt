package com.example.whisper_android.presentation.assistant

import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.core.content.ContextCompat
import android.Manifest
import android.content.pm.PackageManager
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

import com.example.whisper_android.presentation.components.FeatureScreenTemplate
import com.example.whisper_android.presentation.components.MessageRole
import com.example.whisper_android.presentation.components.TranscriptionMessage

@Composable
fun AiAssistantScreen(
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
        title = "AI Assistant",
        onNavigateBack = onNavigateBack,
        isRecording = isRecording,
        isProcessing = isProcessing,
        hasPermission = hasPermission,
        transcriptionResults = transcriptionResults,
        onMicClick = {
            if (!isRecording && !isProcessing) {
                isRecording = true
            }
        },
        onStopClick = {
            if (isRecording) {
                isRecording = false
                isProcessing = true
                scope.launch {
                    delay(1200)
                    val mockUserText = "Summarize the main points of our meeting."
                    val mockAssistantText = "Here is a summary: You discussed the Q3 timeline, focusing on Alpha and Beta release dates."
                    
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
                text = "Interactive AI Assistant for real-time tasks.",
                style = MaterialTheme.typography.bodyMedium
            )

            Text(
                text = "🔄 How it works",
                style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                color = MaterialTheme.colorScheme.primary
            )
            Text(
                text = "1. Press mic to start recording.\n" +
                        "2. System captures voice interaction.\n" +
                        "3. Real-time processing via Whisper engine.\n" +
                        "4. Results displayed in a conversational format.",
                style = MaterialTheme.typography.bodyMedium
            )
        }
    )
}
