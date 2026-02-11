package com.example.whisper_android.presentation.components

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.automirrored.outlined.HelpOutline
import androidx.compose.material.icons.filled.DeleteSweep
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

enum class MessageRole { USER, ASSISTANT }
data class TranscriptionMessage(val text: String, val role: MessageRole)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun FeatureScreenTemplate(
    title: String,
    onNavigateBack: () -> Unit,
    isRecording: Boolean,
    isProcessing: Boolean,
    hasPermission: Boolean,
    transcriptionResults: List<TranscriptionMessage> = emptyList(),
    onMicClick: () -> Unit,
    onClearChat: () -> Unit,
    walkthroughContent: @Composable ColumnScope.() -> Unit,
    modifier: Modifier = Modifier,
    isPaused: Boolean = false,
    thinkingState: Boolean = false,
    onLongMicClick: (() -> Unit)? = null,
    onDoubleMicClick: (() -> Unit)? = null,
    customContent: @Composable (ColumnScope.() -> Unit)? = null,
    extraActions: @Composable RowScope.() -> Unit = {}
) {
    var showWalkthrough by remember { mutableStateOf(false) } // Default to false for manual trigger
    val scrollState = rememberLazyListState()

    // Auto-scroll to bottom
    LaunchedEffect(transcriptionResults.size, isProcessing) {
        if (transcriptionResults.isNotEmpty() || isProcessing) {
            scrollState.animateScrollToItem(
                if (isProcessing) transcriptionResults.size else maxOf(0, transcriptionResults.size - 1)
            )
        }
    }

    // Walkthrough Modal (Reusable Component)
    ScrollableWalkthroughModal(
        title = title,
        showDialog = showWalkthrough,
        onDismiss = { showWalkthrough = false },
        content = walkthroughContent
    )

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(title, fontWeight = FontWeight.Bold) },
                navigationIcon = {
                    IconButton(onClick = onNavigateBack) {
                        Icon(imageVector = Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
                    }
                },
                actions = {
                    if (transcriptionResults.isNotEmpty()) {
                        IconButton(onClick = onClearChat) {
                            Icon(
                                imageVector = Icons.Default.DeleteSweep,
                                contentDescription = "Clear Chat",
                                tint = MaterialTheme.colorScheme.onSurfaceVariant
                            )
                        }
                    }
                    extraActions()
                    IconButton(onClick = { showWalkthrough = true }) {
                        Icon(
                            imageVector = Icons.AutoMirrored.Outlined.HelpOutline,
                            contentDescription = "Help",
                            tint = MaterialTheme.colorScheme.primary
                        )
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(containerColor = Color.Transparent)
            )
        },
        modifier = modifier
    ) { paddingValues ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(
                    Brush.verticalGradient(
                        colors = listOf(
                            MaterialTheme.colorScheme.surface,
                            MaterialTheme.colorScheme.primaryContainer.copy(alpha = 0.05f)
                        )
                    )
                )
                .padding(paddingValues)
        ) {
            Row(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(horizontal = 8.dp),
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                // --- LEFT PANEL: Controls ---
                Column(
                    modifier = Modifier
                        .weight(0.3f)
                        .fillMaxHeight(),
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.Center
                ) {
                    MicButton(
                        isRecording = isRecording,
                        isProcessing = isProcessing || thinkingState,
                        isPaused = isPaused,
                        hasPermission = hasPermission,
                        size = 80.dp,
                        onClick = onMicClick,
                        onLongClick = onLongMicClick,
                        onDoubleClick = onDoubleMicClick
                    )

                    Spacer(modifier = Modifier.height(16.dp))

                    Text(
                        text = when {
                            !hasPermission -> "No Access"
                            isProcessing || thinkingState -> "Thinking..."
                            isPaused -> "Paused"
                            isRecording -> "REC"
                            else -> "Ready"
                        },
                        style = MaterialTheme.typography.titleSmall.copy(fontWeight = FontWeight.Bold),
                        color = when {
                            !hasPermission -> Color.Gray
                            isProcessing || thinkingState -> Color(0xFFFF9800)
                            isPaused -> Color(0xFF2196F3)
                            isRecording -> MaterialTheme.colorScheme.error
                            else -> Color(0xFF4CAF50)
                        }
                    )
                }

                // --- RIGHT PANEL: Chat History ---
                Column(
                    modifier = Modifier
                        .weight(0.7f)
                        .fillMaxHeight()
                        .padding(vertical = 16.dp)
                ) {
                    Text(
                        text = "Interaction Log",
                        style = MaterialTheme.typography.labelMedium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant,
                        modifier = Modifier.padding(bottom = 8.dp)
                    )

                    Surface(
                        modifier = Modifier.fillMaxSize(),
                        color = MaterialTheme.colorScheme.surface.copy(alpha = 0.4f),
                        shape = RoundedCornerShape(16.dp),
                        border = BorderStroke(1.dp, MaterialTheme.colorScheme.outlineVariant.copy(alpha = 0.5f))
                    ) {
                        if (customContent != null) {
                            Column(modifier = Modifier.fillMaxSize().padding(12.dp)) {
                                customContent()
                            }
                        } else if (transcriptionResults.isEmpty() && !isProcessing) {
                            Box(contentAlignment = Alignment.Center) {
                                Text(
                                    text = "Ready to start...",
                                    style = MaterialTheme.typography.bodySmall,
                                    color = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.5f)
                                )
                            }
                        } else {
                            Box(modifier = Modifier.fillMaxSize()) {
                                LazyColumn(
                                    state = scrollState,
                                    modifier = Modifier.fillMaxSize(),
                                    contentPadding = PaddingValues(12.dp),
                                    verticalArrangement = Arrangement.spacedBy(12.dp)
                                ) {
                                    items(transcriptionResults) { message ->
                                        TemplateChatBubble(message)
                                    }
                                    if (isProcessing) {
                                        item {
                                            Text(
                                                text = "Agent is thinking...",
                                                style = MaterialTheme.typography.bodySmall,
                                                modifier = Modifier.padding(start = 8.dp),
                                                color = MaterialTheme.colorScheme.primary
                                            )
                                        }
                                    }
                                }
                                VerticalScrollbar(
                                    modifier = Modifier
                                        .align(Alignment.CenterEnd)
                                        .padding(vertical = 12.dp, horizontal = 4.dp),
                                    lazyListState = scrollState
                                )
                            }
                        }
                    }
                }
            }
        }
    }
}

@Composable
fun TemplateChatBubble(message: TranscriptionMessage) {
    val isUser = message.role == MessageRole.USER
    Column(
        modifier = Modifier.fillMaxWidth(),
        horizontalAlignment = if (isUser) Alignment.End else Alignment.Start
    ) {
        Card(
            shape = RoundedCornerShape(
                topStart = 12.dp,
                topEnd = 12.dp,
                bottomStart = if (isUser) 12.dp else 2.dp,
                bottomEnd = if (isUser) 2.dp else 12.dp
            ),
            colors = CardDefaults.cardColors(
                containerColor = if (isUser) 
                    MaterialTheme.colorScheme.primaryContainer 
                else 
                    MaterialTheme.colorScheme.secondaryContainer.copy(alpha = 0.8f)
            ),
            modifier = Modifier.widthIn(max = 220.dp)
        ) {
            Text(
                text = message.text,
                modifier = Modifier.padding(10.dp),
                style = MaterialTheme.typography.bodySmall.copy(lineHeight = 16.sp),
                color = if (isUser) 
                    MaterialTheme.colorScheme.onPrimaryContainer 
                else 
                    MaterialTheme.colorScheme.onSecondaryContainer
            )
        }
        Text(
            text = if (isUser) "You" else "Senso",
            style = MaterialTheme.typography.labelSmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.6f),
            modifier = Modifier.padding(horizontal = 4.dp, vertical = 2.dp)
        )
    }
}
