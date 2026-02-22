package com.example.whisper_android.presentation.components

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ColumnScope
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.RowScope
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.automirrored.outlined.HelpOutline
import androidx.compose.material.icons.filled.DeleteSweep
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

enum class MessageRole { USER, ASSISTANT }

data class TranscriptionMessage(
    val text: String,
    val role: MessageRole
)

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
    onStopClick: (() -> Unit)? = null, // Added dedicated stop
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
                        Icon(
                            imageVector = Icons.AutoMirrored.Filled.ArrowBack,
                            contentDescription = "Back"
                        )
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
            modifier =
            Modifier
                .fillMaxSize()
                .background(
                    Brush.verticalGradient(
                        colors =
                        listOf(
                            MaterialTheme.colorScheme.surface,
                            MaterialTheme.colorScheme.primaryContainer.copy(alpha = 0.15f),
                            MaterialTheme.colorScheme.secondaryContainer.copy(alpha = 0.05f)
                        )
                    )
                ).padding(paddingValues)
        ) {
            Column(
                modifier =
                Modifier
                    .fillMaxSize()
                    .padding(horizontal = 16.dp)
            ) {
                // --- Top Area: Strategic Insights / Insights Area ---
                Column(
                    modifier =
                    Modifier
                        .weight(1f)
                        .fillMaxWidth()
                        .padding(top = 16.dp, bottom = 8.dp)
                ) {
                    if (customContent != null) {
                        customContent()
                    } else {
                        // Default transcription list if no custom content
                        Surface(
                            modifier = Modifier.fillMaxSize(),
                            color = Color.Transparent, // Let the background show
                            shape = RoundedCornerShape(16.dp)
                        ) {
                            if (transcriptionResults.isEmpty() && !isProcessing) {
                                Box(contentAlignment = Alignment.Center) {
                                    Text(
                                        text = "Strategic Insights will appear here...",
                                        style = MaterialTheme.typography.bodyMedium,
                                        color = MaterialTheme.colorScheme.onSurfaceVariant.copy(
                                            alpha = 0.4f
                                        )
                                    )
                                }
                            } else {
                                LazyColumn(
                                    state = scrollState,
                                    modifier = Modifier.fillMaxSize(),
                                    contentPadding = PaddingValues(bottom = 16.dp),
                                    verticalArrangement = Arrangement.spacedBy(16.dp)
                                ) {
                                    items(transcriptionResults) { message ->
                                        TemplateChatBubble(message)
                                    }
                                    if (isProcessing) {
                                        item {
                                            Text(
                                                text = "Analyzing...",
                                                style = MaterialTheme.typography.labelSmall,
                                                color = MaterialTheme.colorScheme.primary,
                                                modifier = Modifier.padding(start = 8.dp)
                                            )
                                        }
                                    }
                                }
                            }
                        }
                    }
                }

                // --- Bottom Area: Sleek Control Panel ---
                Card(
                    modifier =
                    Modifier
                        .fillMaxWidth()
                        .padding(bottom = 24.dp),
                    shape = RoundedCornerShape(24.dp),
                    colors =
                    CardDefaults.cardColors(
                        containerColor = MaterialTheme.colorScheme.surface.copy(alpha = 0.7f)
                    ),
                    elevation = CardDefaults.cardElevation(defaultElevation = 0.dp),
                    border = BorderStroke(
                        1.dp,
                        MaterialTheme.colorScheme.outlineVariant.copy(alpha = 0.3f)
                    )
                ) {
                    Row(
                        modifier =
                        Modifier
                            .fillMaxWidth()
                            .padding(16.dp),
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.SpaceEvenly
                    ) {
                        // Left: Action Icons
                        IconButton(
                            onClick = onClearChat,
                            enabled = transcriptionResults.isNotEmpty() || customContent != null
                        ) {
                            Icon(
                                Icons.Default.DeleteSweep,
                                contentDescription = "Clear",
                                tint = MaterialTheme.colorScheme.onSurfaceVariant
                            )
                        }

                        // Center: Primary Actions (Mic & Stop)
                        Box(contentAlignment = Alignment.Center) {
                            Row(
                                verticalAlignment = Alignment.CenterVertically,
                                horizontalArrangement = Arrangement.spacedBy(16.dp)
                            ) {
                                MicButton(
                                    isRecording = isRecording,
                                    isProcessing = isProcessing || thinkingState,
                                    isPaused = isPaused,
                                    hasPermission = hasPermission,
                                    size = if (isRecording) 64.dp else 80.dp, // Shrink a bit when stop is shown
                                    onClick = onMicClick
                                )

                                if (isRecording && onStopClick != null) {
                                    StopButton(onClick = onStopClick, size = 64.dp)
                                }
                            }
                        }

                        // Status Indicator/Help
                        Column(horizontalAlignment = Alignment.CenterHorizontally) {
                            Text(
                                text =
                                when {
                                    !hasPermission -> "No Mic"
                                    isProcessing || thinkingState -> "Thinking"
                                    isPaused -> "Paused"
                                    isRecording -> "Recording"
                                    else -> "Ready"
                                },
                                style = MaterialTheme.typography.labelSmall.copy(
                                    fontWeight = FontWeight.Bold
                                ),
                                color =
                                when {
                                    !hasPermission -> Color.Gray
                                    isProcessing || thinkingState -> Color(0xFFFF9800)
                                    isPaused -> Color(0xFF2196F3)
                                    isRecording -> MaterialTheme.colorScheme.error
                                    else -> Color(0xFF4CAF50)
                                }
                            )
                        }
                    }
                }
            }
        }
    }
}

@Composable
fun StopButton(
    onClick: () -> Unit,
    size: Dp = 64.dp
) {
    Surface(
        onClick = onClick,
        shape = CircleShape,
        color = MaterialTheme.colorScheme.surface,
        border = BorderStroke(2.dp, MaterialTheme.colorScheme.error.copy(alpha = 0.5f)),
        modifier = Modifier.size(size),
        tonalElevation = 2.dp
    ) {
        Box(contentAlignment = Alignment.Center) {
            Box(
                modifier =
                Modifier
                    .size(size / 2.5f)
                    .background(MaterialTheme.colorScheme.error, RoundedCornerShape(4.dp))
            )
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
            shape =
            RoundedCornerShape(
                topStart = 12.dp,
                topEnd = 12.dp,
                bottomStart = if (isUser) 12.dp else 2.dp,
                bottomEnd = if (isUser) 2.dp else 12.dp
            ),
            colors =
            CardDefaults.cardColors(
                containerColor =
                if (isUser) {
                    MaterialTheme.colorScheme.primaryContainer
                } else {
                    MaterialTheme.colorScheme.secondaryContainer.copy(alpha = 0.8f)
                }
            ),
            modifier = Modifier.widthIn(max = 220.dp)
        ) {
            Text(
                text = message.text,
                modifier = Modifier.padding(10.dp),
                style = MaterialTheme.typography.bodySmall.copy(lineHeight = 16.sp),
                color =
                if (isUser) {
                    MaterialTheme.colorScheme.onPrimaryContainer
                } else {
                    MaterialTheme.colorScheme.onSecondaryContainer
                }
            )
        }
    }
}
