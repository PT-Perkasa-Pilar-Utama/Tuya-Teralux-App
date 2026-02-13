package com.example.whisper_android.presentation.meeting

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.DeleteOutline
import androidx.compose.material.icons.filled.Download
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.core.content.ContextCompat
import android.Manifest
import android.content.pm.PackageManager
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import dev.jeziellago.compose.markdowntext.MarkdownText
import com.example.whisper_android.presentation.components.*

@Composable
fun MeetingTranscriberScreen(
    onNavigateBack: () -> Unit
) {
    var isRecording by remember { mutableStateOf(false) }
    var isProcessing by remember { mutableStateOf(false) }
    var summaryLanguage by remember { mutableStateOf("id") }
    var displaySummary by remember { mutableStateOf("") }
    
    val context = LocalContext.current
    val scope = rememberCoroutineScope()
    
    val hasPermission = ContextCompat.checkSelfPermission(
        context,
        Manifest.permission.RECORD_AUDIO
    ) == PackageManager.PERMISSION_GRANTED

    FeatureBackground {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(horizontal = 4.dp, vertical = 2.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            // Header
            FeatureHeader(
                title = "Meeting Transcriber & Summary",
                onNavigateBack = onNavigateBack,
                titleColor = MaterialTheme.colorScheme.primary,
                iconColor = MaterialTheme.colorScheme.primary
            )

            // Main Transcription Card
            FeatureMainCard(
                modifier = Modifier.weight(1f)
            ) {
                Column(
                    modifier = Modifier
                        .fillMaxSize()
                        .verticalScroll(rememberScrollState())
                ) {
                    // Header Controls (Download + Language)
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        if (displaySummary.isNotEmpty()) {
                            Button(
                                onClick = { /* Download PDF logic */ },
                                modifier = Modifier.height(32.dp),
                                contentPadding = PaddingValues(horizontal = 12.dp, vertical = 0.dp),
                                colors = ButtonDefaults.buttonColors(
                                    containerColor = MaterialTheme.colorScheme.primary
                                ),
                                shape = RoundedCornerShape(16.dp)
                            ) {
                                Icon(
                                    imageVector = Icons.Default.Download,
                                    contentDescription = null,
                                    modifier = Modifier.size(14.dp),
                                    tint = Color.White
                                )
                                Spacer(modifier = Modifier.width(4.dp))
                                Text(
                                    text = "PDF",
                                    fontSize = 11.sp,
                                    fontWeight = FontWeight.Bold,
                                    color = Color.White
                                )
                            }
                        } else {
                            Spacer(modifier = Modifier.width(1.dp))
                        }

                        LanguagePillToggle(
                            selectedLanguage = summaryLanguage,
                            onLanguageSelected = { summaryLanguage = it }
                        )
                    }
                    
                    Spacer(modifier = Modifier.height(8.dp))

                    if (displaySummary.isEmpty() && !isProcessing) {
                        Box(
                            modifier = Modifier.weight(1f).fillMaxWidth().padding(vertical = 40.dp),
                            contentAlignment = Alignment.Center
                        ) {
                            Text(
                                text = "Record audio to generate a summary...",
                                style = MaterialTheme.typography.bodyLarge,
                                color = Color.Gray,
                                textAlign = TextAlign.Center
                            )
                        }
                    } else if (isProcessing) {
                        Box(
                            modifier = Modifier.weight(1f).fillMaxWidth().padding(vertical = 40.dp),
                            contentAlignment = Alignment.Center
                        ) {
                            Column(horizontalAlignment = Alignment.CenterHorizontally) {
                                CircularProgressIndicator(color = Color(0xFFFF9800))
                                Spacer(modifier = Modifier.height(8.dp))
                                Text(
                                    text = "Generating summary...",
                                    style = MaterialTheme.typography.bodyMedium.copy(fontSize = 13.sp),
                                    color = Color(0xFFFF9800)
                                )
                            }
                        }
                    } else {
                        MarkdownText(
                            markdown = displaySummary,
                            style = MaterialTheme.typography.bodyMedium.copy(
                                color = Color.Black,
                                fontSize = 13.sp,
                                lineHeight = 18.sp
                            ),
                            modifier = Modifier.fillMaxWidth()
                        )
                    }
                }
            }

            Spacer(modifier = Modifier.height(4.dp))

            // Bottom Centered Control Pill
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(bottom = 4.dp),
                contentAlignment = Alignment.Center
            ) {
                Surface(
                    modifier = Modifier
                        .height(48.dp)
                        .widthIn(min = 220.dp),
                    shape = RoundedCornerShape(24.dp),
                    color = Color.White.copy(alpha = 0.85f),
                    tonalElevation = 6.dp,
                    shadowElevation = 8.dp
                ) {
                    Row(
                        modifier = Modifier.padding(horizontal = 10.dp),
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(10.dp)
                    ) {
                        MicButton(
                            isRecording = isRecording,
                            hasPermission = hasPermission,
                            isProcessing = isProcessing,
                            onClick = { 
                                if (!isRecording) isRecording = true 
                            },
                            size = 38.dp
                        )

                        if (isRecording || isProcessing) {
                            IconButton(
                                onClick = { 
                                    if (isRecording) {
                                        isRecording = false
                                        isProcessing = true
                                        scope.launch {
                                            delay(3000)
                                            displaySummary = SummaryUtils.loadAndFormatSummary(context, summaryLanguage)
                                            isProcessing = false
                                        }
                                    }
                                },
                                modifier = Modifier.size(32.dp)
                            ) {
                                Surface(
                                    modifier = Modifier.size(16.dp),
                                    color = Color(0xFFEF5350),
                                    shape = RoundedCornerShape(2.dp)
                                ) {}
                            }
                        }

                        IconButton(
                            onClick = { 
                                displaySummary = ""
                                isRecording = false
                                isProcessing = false
                            },
                            modifier = Modifier.size(32.dp)
                        ) {
                            Icon(
                                imageVector = Icons.Default.DeleteOutline,
                                contentDescription = "Clear",
                                tint = MaterialTheme.colorScheme.primary,
                                modifier = Modifier.size(20.dp)
                            )
                        }

                        Text(
                            text = when {
                                isRecording -> "Recording..."
                                isProcessing -> "Thinking..."
                                else -> "Ready"
                            },
                            fontSize = 14.sp,
                            color = MaterialTheme.colorScheme.primary,
                            fontWeight = FontWeight.Medium,
                            modifier = Modifier.padding(end = 4.dp)
                        )
                    }
                }
            }
        }
    }
}
