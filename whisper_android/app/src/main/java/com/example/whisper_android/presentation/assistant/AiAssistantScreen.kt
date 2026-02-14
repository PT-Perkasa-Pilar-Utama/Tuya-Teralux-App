package com.example.whisper_android.presentation.assistant

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import android.Manifest
import android.content.pm.PackageManager
import androidx.core.content.ContextCompat
import com.example.whisper_android.presentation.components.*
import com.example.whisper_android.data.di.NetworkModule
import com.example.whisper_android.domain.repository.Resource
import kotlinx.coroutines.flow.collect
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import androidx.compose.ui.unit.sp

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

    val hasPermission = ContextCompat.checkSelfPermission(
        context,
        Manifest.permission.RECORD_AUDIO
    ) == PackageManager.PERMISSION_GRANTED

    // Permission Launcher
    val permissionLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.RequestPermission()
    ) { isGranted ->
        if (isGranted) {
            // Permission granted, will update hasPermission next recompose
        }
    }

    FeatureBackground {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(horizontal = 4.dp, vertical = 0.dp), // Zero vertical padding on root for tighter control
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            // Header with Accent
            Box(modifier = Modifier.fillMaxWidth()) {
                FeatureHeader(
                    title = "Whisper Intelligence",
                    onNavigateBack = onNavigateBack,
                    titleColor = MaterialTheme.colorScheme.primary,
                    iconColor = MaterialTheme.colorScheme.primary
                )
                
                // Subtle Neural Accent Icon
                Box(
                    modifier = Modifier
                        .align(Alignment.CenterEnd)
                        .padding(end = 20.dp, top = 16.dp) // Tighter
                        .size(20.dp) // Even smaller/minimal
                        .alpha(0.2f) // "Visible but invisible"
                ) {
                    WhisperLogo(showText = false)
                }
            }

            // Main Conversation Card (Fills space)
            FeatureMainCard(
                modifier = Modifier.weight(1f)
            ) {
                if (transcriptionResults.isEmpty() && !isProcessing) {
                    Column(
                        modifier = Modifier
                            .fillMaxSize(),
                        horizontalAlignment = Alignment.CenterHorizontally,
                        verticalArrangement = Arrangement.Top
                    ) {
                        // 1. Optical Anchor (Upper 35-40%)
                        Spacer(modifier = Modifier.weight(0.6f))
                        
                        // 2. Identification (Orb)
                        AiMindVisual(isThinking = isProcessing)
                        
                        Spacer(modifier = Modifier.height(24.dp)) // Orb -> Title
                        
                        // 3. Functional Meta
                        Text(
                            text = "Meeting Index Ready.",
                            style = MaterialTheme.typography.titleLarge.copy(
                                fontWeight = FontWeight.Bold,
                                letterSpacing = (-0.5).sp,
                                fontSize = 20.sp
                            ),
                            color = MaterialTheme.colorScheme.onSurface
                        )
                        Spacer(modifier = Modifier.height(8.dp)) // Title -> Subtitle (Strict)
                        Text(
                            text = "Ask anything from your captured data.",
                            style = MaterialTheme.typography.bodyMedium,
                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.35f),
                            fontWeight = FontWeight.Normal,
                            textAlign = TextAlign.Center
                        )
                        
                        Spacer(modifier = Modifier.height(32.dp)) // Subtitle -> Intelligence Layer
                        
                        // 3. Suggested Action Cards (Horizontal Grid)
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(horizontal = 24.dp), // Tighter horizontal padding
                            horizontalArrangement = Arrangement.spacedBy(16.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            val promptActions = listOf(
                                "Summarize Insight" to "Extract core discussion intent.", // Tighter copy
                                "List Actions" to "Identify tasks and assignments."
                            )
                            
                            promptActions.forEach { (title, subtitle) ->
                                SuggestedActionCard(
                                    title = title,
                                    subtitle = subtitle,
                                    modifier = Modifier.weight(1f), // Equal distribution
                                    onClick = { 
                                        val prompt = title
                                        userInput = ""
                                        isProcessing = true
                                        transcriptionResults = transcriptionResults + 
                                            TranscriptionMessage(prompt, MessageRole.USER)
                                        
                                        scope.launch {
                                            val token = NetworkModule.tokenManager.getAccessToken() ?: ""
                                            // For now, using a dummy or previous text if we had it.
                                            // In a real app, this would use the meeting context.
                                            val contextText = transcriptionResults.filter { it.role == MessageRole.USER }
                                                .joinToString("\n") { it.text }

                                            if (title == "Summarize Insight") {
                                                NetworkModule.summarizeTextUseCase(contextText, null, "meeting_summary", token).collect { result ->
                                                    when (result) {
                                                        is Resource.Loading -> { /* already true */ }
                                                        is Resource.Success -> {
                                                            transcriptionResults = transcriptionResults + 
                                                                TranscriptionMessage(result.data?.summary ?: "No summary generated.", MessageRole.ASSISTANT)
                                                            isProcessing = false
                                                        }
                                                        is Resource.Error -> {
                                                            transcriptionResults = transcriptionResults + 
                                                                TranscriptionMessage("Error: ${result.message}", MessageRole.ASSISTANT)
                                                            isProcessing = false
                                                        }
                                                    }
                                                }
                                            } else {
                                                // List Actions might use a specific style
                                                NetworkModule.summarizeTextUseCase(contextText, null, "action_items", token).collect { result ->
                                                    when (result) {
                                                        is Resource.Loading -> { }
                                                        is Resource.Success -> {
                                                            transcriptionResults = transcriptionResults + 
                                                                TranscriptionMessage(result.data?.summary ?: "No actions identified.", MessageRole.ASSISTANT)
                                                            isProcessing = false
                                                        }
                                                        is Resource.Error -> {
                                                            transcriptionResults = transcriptionResults + 
                                                                TranscriptionMessage("Error: ${result.message}", MessageRole.ASSISTANT)
                                                            isProcessing = false
                                                        }
                                                    }
                                                }
                                            }
                                        }
                                    }
                                )
                            }
                        }
                        
                        // 4. Input Integration Gap
                        Spacer(modifier = Modifier.height(24.dp))
                        
                        // 5. Lower Weight to balance crop test
                        Spacer(modifier = Modifier.weight(1f))
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
                                TypingIndicator()
                            }
                        }
                    }
                }
            }

            // No Spacer needed for tighter integration

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
                        if (!hasPermission) {
                            permissionLauncher.launch(Manifest.permission.RECORD_AUDIO)
                        } else if (!isRecording && !isProcessing) {
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
                            transcriptionResults = transcriptionResults + 
                                TranscriptionMessage(question, MessageRole.USER)
                                
                            scope.launch {
                                val token = NetworkModule.tokenManager.getAccessToken() ?: ""
                                // Treating a general question as a summary request where the text is the question
                                // Ideally this would be a 'QA' endpoint, but we'll use Summary for demonstration
                                NetworkModule.summarizeTextUseCase(question, null, "qa", token).collect { result ->
                                    when (result) {
                                        is Resource.Success -> {
                                            transcriptionResults = transcriptionResults + 
                                                TranscriptionMessage(result.data?.summary ?: "I couldn't find an answer.", MessageRole.ASSISTANT)
                                            isProcessing = false
                                        }
                                        is Resource.Error -> {
                                            transcriptionResults = transcriptionResults + 
                                                TranscriptionMessage("Error: ${result.message}", MessageRole.ASSISTANT)
                                            isProcessing = false
                                        }
                                        else -> {}
                                    }
                                }
                            }
                        }
                    }
                )
            }
        }
    }
}
