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
import androidx.compose.ui.unit.sp
import com.example.whisper_android.util.MqttHelper
import java.io.File
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import com.example.whisper_android.presentation.components.*
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import android.util.Log

import androidx.lifecycle.viewmodel.compose.viewModel
import com.example.whisper_android.util.parseMarkdownToText

@Composable
fun AiAssistantScreen(
    onNavigateBack: () -> Unit,
    viewModel: AiAssistantViewModel = viewModel()
) {
    val transcriptionResults = viewModel.transcriptionResults
    val isRecording = viewModel.isRecording
    val isProcessing = viewModel.isProcessing
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
            // Permission granted
        }
    }

    // Wake Word Manager
    val wakeWordManager = remember {
        SensioWakeWordManager(context) {
            if (!isRecording && !isProcessing) {
                viewModel.startRecording()
            }
        }
    }

    // Wake Word Lifecycle
    DisposableEffect(hasPermission, isRecording, isProcessing) {
        if (hasPermission && !isRecording && !isProcessing) {
            wakeWordManager.startListening()
        } else {
            wakeWordManager.stopListening()
        }
        
        onDispose {
            wakeWordManager.stopListening()
            wakeWordManager.destroy()
        }
    }

    // Smart Mic: Auto-stop if no command detected
    val snackbarHostState = remember { SnackbarHostState() }
    LaunchedEffect(isRecording) {
        if (isRecording) {
            delay(6000)
            if (isRecording) {
                viewModel.stopRecording(File(context.cacheDir, "recording.wav"))
                snackbarHostState.showSnackbar(
                    message = "Mic auto-stopped (No command).",
                    duration = SnackbarDuration.Short
                )
            }
        }
    }

    Scaffold(
        snackbarHost = { SnackbarHost(snackbarHostState) },
        containerColor = Color.Transparent
    ) { padding ->
        FeatureBackground {
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(padding)
                    .padding(horizontal = 4.dp, vertical = 0.dp),
                horizontalAlignment = Alignment.CenterHorizontally
            ) {
            Box(modifier = Modifier.fillMaxWidth()) {
                FeatureHeader(
                    title = "Whisper Intelligence",
                    onNavigateBack = onNavigateBack,
                    titleColor = MaterialTheme.colorScheme.primary,
                    iconColor = MaterialTheme.colorScheme.primary
                )
                
                Box(
                    modifier = Modifier
                        .align(Alignment.CenterEnd)
                        .padding(end = 20.dp, top = 16.dp)
                        .size(20.dp)
                        .alpha(0.2f)
                ) {
                    WhisperLogo(showText = false)
                }
            }

            FeatureMainCard(
                modifier = Modifier.weight(1f)
            ) {
                if (transcriptionResults.isEmpty() && !isProcessing) {
                    EmptyAssistantState(
                        isProcessing = isProcessing,
                        onSuggestedAction = { prompt ->
                            viewModel.sendChat(prompt)
                        }
                    )
                } else {
                    ConversationList(
                        scrollState = scrollState,
                        messages = transcriptionResults,
                        isProcessing = isProcessing
                    )
                }
            }

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
                            viewModel.startRecording()
                        }
                    },
                    onStopClick = {
                        viewModel.stopRecording(File(context.cacheDir, "recording.wav"))
                    },
                    onSendClick = {
                        if (userInput.isNotBlank()) {
                            viewModel.sendChat(userInput)
                            userInput = ""
                        }
                    }
                )
            }
            }
        }
    }
}

@Composable
private fun EmptyAssistantState(
    isProcessing: Boolean,
    onSuggestedAction: (String) -> Unit
) {
    Column(
        modifier = Modifier.fillMaxSize(),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Top
    ) {
        Spacer(modifier = Modifier.weight(0.6f))
        AiMindVisual(isThinking = isProcessing)
        Spacer(modifier = Modifier.height(24.dp))
        Text(
            text = "Meeting Index Ready.",
            style = MaterialTheme.typography.titleLarge.copy(
                fontWeight = FontWeight.Bold,
                letterSpacing = (-0.5).sp,
                fontSize = 20.sp
            ),
            color = MaterialTheme.colorScheme.onSurface
        )
        Spacer(modifier = Modifier.height(8.dp))
        Text(
            text = "Ask anything from your captured data.",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.35f),
            textAlign = TextAlign.Center
        )
        Spacer(modifier = Modifier.height(32.dp))
        
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 24.dp),
            horizontalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            SuggestedActionCard(
                title = "Summarize Insight",
                subtitle = "Extract core discussion intent.",
                modifier = Modifier.weight(1f),
                onClick = { onSuggestedAction("Summarize the meeting.") }
            )
            SuggestedActionCard(
                title = "List Actions",
                subtitle = "Identify tasks and assignments.",
                modifier = Modifier.weight(1f),
                onClick = { onSuggestedAction("What are the action items?") }
            )
        }
        Spacer(modifier = Modifier.height(24.dp))
        Spacer(modifier = Modifier.weight(1f))
    }
}

@Composable
private fun ConversationList(
    scrollState: androidx.compose.foundation.lazy.LazyListState,
    messages: List<TranscriptionMessage>,
    isProcessing: Boolean
) {
    LazyColumn(
        state = scrollState,
        modifier = Modifier.fillMaxSize(),
        verticalArrangement = Arrangement.spacedBy(16.dp),
        contentPadding = PaddingValues(bottom = 16.dp)
    ) {
        items(messages) { message ->
            AssistantChatBubble(message)
        }
        if (isProcessing) {
            item {
                TypingIndicator()
            }
        }
    }
}
