package com.example.whisper_android.presentation.assistant

import android.Manifest
import android.content.pm.PackageManager
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.expandVertically
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.shrinkVertically
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.SnackbarDuration
import androidx.compose.material3.SnackbarHost
import androidx.compose.material3.SnackbarHostState
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.core.content.ContextCompat
import androidx.lifecycle.viewmodel.compose.viewModel
import com.example.whisper_android.presentation.components.LanguagePillToggle
import com.example.whisper_android.presentation.components.MqttStatusBadge
import com.example.whisper_android.presentation.components.SensioFeatureLayout
import com.example.whisper_android.presentation.components.TranscriptionMessage
import com.example.whisper_android.util.MqttHelper
import java.io.File
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

@OptIn(ExperimentalMaterial3Api::class)
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
                if (isProcessing) {
                    transcriptionResults.size
                } else {
                    maxOf(
                        0,
                        transcriptionResults.size - 1
                    )
                }
            )
        }
    }

    val hasPermission =
        ContextCompat.checkSelfPermission(
            context,
            Manifest.permission.RECORD_AUDIO
        ) == PackageManager.PERMISSION_GRANTED

    // Permission Launcher
    val permissionLauncher =
        rememberLauncherForActivityResult(
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

    val wakeWordManager =
        remember {
            WakeWordFactory.getManager(context) {
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

    SensioFeatureLayout(
        title = "Sensio Intelligence",
        onNavigateBack = onNavigateBack,
        snackbarHost = { SnackbarHost(snackbarHostState) },
        headerActions = {
            Row(
                verticalAlignment = Alignment.CenterVertically,
                modifier = Modifier.padding(end = 8.dp)
            ) {
                LanguagePillToggle(
                    selectedLanguage = viewModel.selectedLanguage,
                    onLanguageSelected = { viewModel.selectLanguage(it) }
                )

                MqttStatusBadge(
                    status = viewModel.mqttStatus,
                    onReconnectClick = { viewModel.reconnectMqtt() }
                )
            }
        },
        bottomContent = {
            Box(
                modifier =
                Modifier
                    .fillMaxWidth()
                    .align(Alignment.BottomCenter)
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
                        if (isRecording) {
                            viewModel.stopRecording()
                        } else if (isProcessing) {
                            viewModel.abortProcessing()
                        }
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
}

@Composable
fun EmptyAssistantState(
    isProcessing: Boolean,
    enabled: Boolean,
    onSuggestedAction: (String) -> Unit,
    modifier: Modifier = Modifier
) {
    BoxWithConstraints(modifier = modifier.fillMaxSize()) {
        val isWide = maxWidth > 600.dp

        Column(
            modifier = Modifier.fillMaxSize(),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            AiMindVisual(isThinking = isProcessing)
            Spacer(modifier = Modifier.height(24.dp))
            Text(
                text = "Meeting Index Ready.",
                style =
                MaterialTheme.typography.titleLarge.copy(
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

            // Suggested Actions
            if (isWide) {
                Row(
                    modifier = Modifier.fillMaxWidth().padding(horizontal = 24.dp),
                    horizontalArrangement = Arrangement.spacedBy(16.dp)
                ) {
                    SuggestedActionCard(
                        title = "Introduce Yourself",
                        subtitle = "Learn about my role.",
                        modifier = Modifier.weight(1f),
                        onClick = { onSuggestedAction("Who are you?") },
                        enabled = enabled
                    )
                    SuggestedActionCard(
                        title = "Explore Controls",
                        subtitle = "What devices can I control?",
                        modifier = Modifier.weight(1f),
                        onClick = { onSuggestedAction("What devices can I control?") },
                        enabled = enabled
                    )
                }
            } else {
                Column(
                    modifier = Modifier.fillMaxWidth().padding(horizontal = 24.dp),
                    verticalArrangement = Arrangement.spacedBy(12.dp)
                ) {
                    SuggestedActionCard(
                        title = "Introduce Yourself",
                        subtitle = "Learn about my role.",
                        modifier = Modifier.fillMaxWidth(),
                        onClick = { onSuggestedAction("Who are you?") },
                        enabled = enabled
                    )
                    SuggestedActionCard(
                        title = "Explore Controls",
                        subtitle = "What devices can I control?",
                        modifier = Modifier.fillMaxWidth(),
                        onClick = { onSuggestedAction("What devices can I control?") },
                        enabled = enabled
                    )
                }
            }
        }
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
                    modifier =
                    Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 12.dp, vertical = 4.dp),
                    horizontalAlignment = Alignment.Start
                ) {
                    Surface(
                        shape =
                        RoundedCornerShape(
                            topStart = 4.dp,
                            topEnd = 24.dp,
                            bottomStart = 24.dp,
                            bottomEnd = 24.dp
                        ),
                        color = Color.White.copy(alpha = 0.9f),
                        modifier = Modifier.widthIn(max = 300.dp),
                        border =
                        androidx.compose.foundation.BorderStroke(
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

// Removed duplicate MqttStatusBadge (moved to MqttStatusBadge.kt)
