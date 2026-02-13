package com.example.whisper_android.presentation.assistant

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import com.example.whisper_android.presentation.components.*
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

@Composable
fun AiAssistantScreen(
    onNavigateBack: () -> Unit
) {
    var isRecording by remember { mutableStateOf(false) }
    var isProcessing by remember { mutableStateOf(false) }
    var transcriptionResults by remember { mutableStateOf(listOf<TranscriptionMessage>()) }
    var userInput by remember { mutableStateOf("") }
    
    val context = LocalContext.current
    val scope = rememberCoroutineScope()
    val scrollState = rememberLazyListState()

    // Auto-scroll to bottom
    LaunchedEffect(transcriptionResults.size, isProcessing) {
        if (transcriptionResults.isNotEmpty() || isProcessing) {
            scrollState.animateScrollToItem(
                if (isProcessing) transcriptionResults.size else maxOf(0, transcriptionResults.size - 1)
            )
        }
    }

    val hasPermission = androidx.core.content.ContextCompat.checkSelfPermission(
        context,
        android.Manifest.permission.RECORD_AUDIO
    ) == android.content.pm.PackageManager.PERMISSION_GRANTED

    FeatureBackground {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(horizontal = 4.dp, vertical = 2.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            // Header
            FeatureHeader(
                title = "AI Assistant",
                onNavigateBack = onNavigateBack,
                titleColor = MaterialTheme.colorScheme.primary,
                iconColor = MaterialTheme.colorScheme.primary
            )

            // Main Conversation Card (Fills space)
            FeatureMainCard(
                modifier = Modifier.weight(1f)
            ) {
                if (transcriptionResults.isEmpty() && !isProcessing) {
                    Box(modifier = Modifier.fillMaxSize().padding(32.dp), contentAlignment = Alignment.Center) {
                        Text(
                            text = "Ask me anything about your meeting...",
                            style = MaterialTheme.typography.bodyMedium,
                            color = Color.Gray,
                            textAlign = TextAlign.Center
                        )
                    }
                } else {
                    LazyColumn(
                        state = scrollState,
                        modifier = Modifier.fillMaxSize(),
                        verticalArrangement = Arrangement.spacedBy(16.dp),
                        contentPadding = PaddingValues(bottom = 16.dp)
                    ) {
                        items(transcriptionResults) { message ->
                            AssistantChatBubble(message)
                        }
                        if (isProcessing) {
                            item {
                                Text(
                                    text = "Thinking...",
                                    style = MaterialTheme.typography.labelSmall,
                                    color = MaterialTheme.colorScheme.primary,
                                    modifier = Modifier.padding(start = 12.dp)
                                )
                            }
                        }
                    }
                }
            }

            Spacer(modifier = Modifier.height(4.dp))

            // Bottom Area (Natural height)
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(bottom = 4.dp),
                contentAlignment = Alignment.Center
            ) {
                AssistantInputPill(
                    inputValue = userInput,
                    onValueChange = { userInput = it },
                    isRecording = isRecording,
                    isProcessing = isProcessing,
                    hasPermission = hasPermission,
                    onMicClick = {
                        if (!isRecording && !isProcessing) {
                            isRecording = true
                        }
                    },
                    onStopClick = {
                        if (isRecording) {
                            isRecording = false
                            isProcessing = true
                            // Simulate processing and answer
                            scope.launch {
                                delay(2000)
                                transcriptionResults = transcriptionResults + 
                                    TranscriptionMessage("What's the summary of the Q3 report?", MessageRole.USER)
                                delay(800)
                                transcriptionResults = transcriptionResults + 
                                    TranscriptionMessage("The Q3 report highlights significant market share growth and assigned several follow-up action items. A preliminary budget agreement was also reached.", MessageRole.ASSISTANT)
                                isProcessing = false
                            }
                        }
                    },
                    onSendClick = {
                        if (userInput.isNotBlank()) {
                            val question = userInput
                            userInput = ""
                            isProcessing = true
                            scope.launch {
                                transcriptionResults = transcriptionResults + 
                                    TranscriptionMessage(question, MessageRole.USER)
                                delay(1500)
                                transcriptionResults = transcriptionResults + 
                                    TranscriptionMessage("I'm looking into your data for: \"$question\". Here's the most relevant information...", MessageRole.ASSISTANT)
                                isProcessing = false
                            }
                        }
                    }
                )
            }
        }
    }
}
