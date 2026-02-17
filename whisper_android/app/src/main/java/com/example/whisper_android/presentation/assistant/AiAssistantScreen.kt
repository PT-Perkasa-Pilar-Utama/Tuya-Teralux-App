package com.example.whisper_android.presentation.assistant

import androidx.compose.animation.*
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.RoundedCornerShape
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
    val isMqttOnline = viewModel.mqttStatus == MqttHelper.MqttConnectionStatus.CONNECTED
    
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
    val currentOnWakeWordDetected by rememberUpdatedState {
        if (isMqttOnline && !isRecording && !isProcessing) {
            viewModel.startRecording(File(context.cacheDir, "recording.wav"))
        }
    }

    val wakeWordManager = remember {
        SensioWakeWordManager(context) {
            currentOnWakeWordDetected()
        }
    }

    // Wake Word Lifecycle
    DisposableEffect(hasPermission, isRecording, isProcessing, isMqttOnline) {
        if (hasPermission && isMqttOnline && !isRecording && !isProcessing) {
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
                viewModel.stopRecording()
                snackbarHostState.showSnackbar(
                    message = "Mic auto-stopped (No command).",
                    duration = SnackbarDuration.Short
                )
            }
        }
    }

    FeatureBackground {
        Scaffold(
            snackbarHost = { SnackbarHost(snackbarHostState) },
            containerColor = Color.Transparent,
            topBar = {
                FeatureHeader(
                    title = "Whisper Intelligence",
                    onNavigateBack = onNavigateBack,
                    titleColor = MaterialTheme.colorScheme.primary,
                    iconColor = MaterialTheme.colorScheme.primary,
                    actions = {
                        Row(
                            modifier = Modifier
                                .padding(end = 8.dp)
                                .background(
                                    Color.LightGray.copy(alpha = 0.2f),
                                    RoundedCornerShape(20.dp)
                                )
                                .padding(2.dp),
                            horizontalArrangement = Arrangement.spacedBy(2.dp)
                        ) {
                            listOf("id", "en").forEach { lang ->
                                val isSelected = viewModel.selectedLanguage == lang
                                Surface(
                                    onClick = { viewModel.selectLanguage(lang) },
                                    shape = RoundedCornerShape(16.dp),
                                    color = if (isSelected) MaterialTheme.colorScheme.primary else Color.Transparent,
                                    modifier = Modifier.size(width = 42.dp, height = 28.dp)
                                ) {
                                    Box(contentAlignment = Alignment.Center) {
                                        Text(
                                            text = lang.uppercase(),
                                            fontSize = 11.sp,
                                            fontWeight = FontWeight.Bold,
                                            color = if (isSelected) Color.White else MaterialTheme.colorScheme.onSurface.copy(
                                                alpha = 0.6f
                                            )
                                        )
                                    }
                                }
                            }
                        }
                        
                        MqttStatusBadge(viewModel.mqttStatus)
                    }
                )
            }
        ) { padding ->
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(padding)
                    .padding(horizontal = 4.dp, vertical = 0.dp),
                horizontalAlignment = Alignment.CenterHorizontally
            ) {
                FeatureMainCard(
                    modifier = Modifier.weight(1f)
                ) {
                    if (transcriptionResults.isEmpty() && !isProcessing) {
                        EmptyAssistantState(
                            isProcessing = isProcessing,
                            enabled = isMqttOnline,
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
                            viewModel.startRecording(File(context.cacheDir, "recording.wav"))
                        }
                    },
                    onStopClick = {
                        viewModel.stopRecording()
                    },
                    onSendClick = {
                        if (userInput.isNotBlank()) {
                            viewModel.sendChat(userInput)
                            userInput = ""
                        }
                    },
                    enabled = isMqttOnline
                )
            }
            }
        }
    }
}

@Composable
private fun EmptyAssistantState(
    isProcessing: Boolean,
    enabled: Boolean,
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
                title = "Introduce Yourself",
                subtitle = "Learn about my role and capabilities.",
                modifier = Modifier.weight(1f),
                onClick = { onSuggestedAction("Who are you?") },
                enabled = enabled
            )
            SuggestedActionCard(
                title = "Explore My Controls",
                subtitle = "Discover which devices I can manage.",
                modifier = Modifier.weight(1f),
                onClick = { onSuggestedAction("What devices can I control?") },
                enabled = enabled
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
        item {
            AnimatedVisibility(
                visible = isProcessing,
                enter = expandVertically() + fadeIn(),
                exit = shrinkVertically() + fadeOut()
            ) {
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 12.dp, vertical = 4.dp),
                    horizontalAlignment = Alignment.Start
                ) {
                    Surface(
                        shape = RoundedCornerShape(
                            topStart = 4.dp,
                            topEnd = 24.dp,
                            bottomStart = 24.dp,
                            bottomEnd = 24.dp
                        ),
                        color = Color.White.copy(alpha = 0.9f),
                        modifier = Modifier.widthIn(max = 300.dp),
                        border = androidx.compose.foundation.BorderStroke(
                            1.dp, 
                            MaterialTheme.colorScheme.primary.copy(alpha = 0.08f)
                        ),
                        shadowElevation = 8.dp,
                        tonalElevation = 4.dp
                    ) {
                        Row(
                            verticalAlignment = Alignment.CenterVertically,
                            modifier = Modifier.padding(horizontal = 4.dp)
                        ) {
                            AiMindVisual(
                                isThinking = true, 
                                size = 28.dp,
                                modifier = Modifier.padding(start = 12.dp)
                            )
                            TypingIndicator()
                        }
                    }
                }
            }
        }
    }
}
@Composable
private fun MqttStatusBadge(status: MqttHelper.MqttConnectionStatus) {
    val color = when (status) {
        MqttHelper.MqttConnectionStatus.CONNECTED -> Color(0xFF4CAF50)
        MqttHelper.MqttConnectionStatus.CONNECTING -> Color(0xFFFFC107)
        MqttHelper.MqttConnectionStatus.DISCONNECTED -> Color(0xFFF44336)
        MqttHelper.MqttConnectionStatus.FAILED -> Color(0xFFD32F2F)
    }
    
    val text = when (status) {
        MqttHelper.MqttConnectionStatus.CONNECTED -> "Online"
        MqttHelper.MqttConnectionStatus.CONNECTING -> "Connecting"
        MqttHelper.MqttConnectionStatus.DISCONNECTED -> "Offline"
        MqttHelper.MqttConnectionStatus.FAILED -> "Error"
    }

    Row(
        verticalAlignment = Alignment.CenterVertically,
        modifier = Modifier
            .padding(start = 4.dp)
            .background(color.copy(alpha = 0.1f), RoundedCornerShape(12.dp))
            .padding(horizontal = 8.dp, vertical = 4.dp)
    ) {
        Box(
            modifier = Modifier
                .size(8.dp)
                .background(color, androidx.compose.foundation.shape.CircleShape)
        )
        Spacer(modifier = Modifier.width(6.dp))
        Text(
            text = text,
            fontSize = 11.sp,
            fontWeight = FontWeight.Medium,
            color = color
        )
    }
}
