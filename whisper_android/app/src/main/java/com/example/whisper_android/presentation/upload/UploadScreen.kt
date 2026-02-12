package com.example.whisper_android.presentation.upload

import android.content.Intent
import android.net.Uri
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.AudioFile
import androidx.compose.material.icons.filled.Download
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.core.content.ContextCompat
import android.Manifest
import android.content.pm.PackageManager
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

import androidx.lifecycle.viewmodel.compose.viewModel
import com.example.whisper_android.data.di.NetworkModule
import com.example.whisper_android.presentation.components.FeatureScreenTemplate
import com.example.whisper_android.presentation.components.TranscriptionMessage
import com.example.whisper_android.presentation.components.ToastObserver
import com.example.whisper_android.presentation.components.AudioFilePicker
import com.example.whisper_android.MainActivity

import dev.jeziellago.compose.markdown.MarkdownText

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun UploadScreen(
    onNavigateBack: () -> Unit,
    viewModel: UploadViewModel = viewModel {
        UploadViewModel(
            NetworkModule.speechRepository,
            AudioRecorder(MainActivity.appContext),
            NetworkModule.tokenManager,
            MainActivity.appContext.cacheDir
        )
    }
) {
    val uiState by viewModel.uiState.collectAsState()
    val context = LocalContext.current
    
    val hasMicPermission = ContextCompat.checkSelfPermission(
        context,
        Manifest.permission.RECORD_AUDIO
    ) == PackageManager.PERMISSION_GRANTED

    val storagePermission = if (android.os.Build.VERSION.SDK_INT >= android.os.Build.VERSION_CODES.TIRAMISU) {
        Manifest.permission.READ_MEDIA_AUDIO
    } else {
        Manifest.permission.READ_EXTERNAL_STORAGE
    }

    var hasStoragePermission by remember {
        mutableStateOf(
            ContextCompat.checkSelfPermission(context, storagePermission) == PackageManager.PERMISSION_GRANTED
        )
    }

    val storagePermissionLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.RequestPermission()
    ) { isGranted: Boolean ->
        hasStoragePermission = isGranted
        if (!isGranted) {
            // Permission denied
        }
    }

    // Observe errors
    ToastObserver(
        message = uiState.error,
        onToastShown = { /* Error handled */ }
    )

    FeatureScreenTemplate(
        title = "Whisper Summary",
        onNavigateBack = onNavigateBack,
        isRecording = uiState.isRecording,
        isProcessing = false, // We use thinkingState instead
        isPaused = uiState.isPaused,
        thinkingState = uiState.isThinking,
        hasPermission = hasMicPermission,
        onMicClick = {
            if (hasMicPermission) {
                viewModel.handleMicClick()
            }
        },
        onLongMicClick = {
            if (hasMicPermission) {
                viewModel.handleMicStop()
            }
        },
        onDoubleMicClick = {
            if (hasMicPermission) {
                viewModel.handleMicStop()
            }
        },
        onClearChat = { viewModel.clearLog() },
        extraActions = {
            AudioFilePicker(
                onFileSelected = { uri -> viewModel.handleFileSelected(uri, context) },
                enabled = !uiState.isRecording && !uiState.isThinking,
                onPermissionDenied = {
                    storagePermissionLauncher.launch(storagePermission)
                },
                onFallbackNeeded = {
                    viewModel.scanDownloadsFolder()
                },
                hasPermission = hasStoragePermission
            )
        },
        customContent = {
            Column(modifier = Modifier.fillMaxSize()) {
                if (uiState.showInternalPicker) {
                    InternalFileSelectionDialog(
                        files = uiState.availableFiles,
                        onFileSelected = { viewModel.handleFileSelected(it) },
                        onDismiss = { viewModel.hideInternalPicker() }
                    )
                }

                // Language Selector
                Row(
                    modifier = Modifier.fillMaxWidth().padding(bottom = 12.dp),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = androidx.compose.ui.Alignment.CenterVertically
                ) {
                    Text(
                        text = "Summary Language:",
                        style = MaterialTheme.typography.labelMedium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                    Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                        listOf("id" to "🇮🇩 ID", "en" to "🇺🇸 EN").forEach { (code, label) ->
                            FilterChip(
                                selected = uiState.summaryLanguage == code,
                                onClick = { viewModel.setSummaryLanguage(code) },
                                label = { Text(label) }
                            )
                        }
                    }
                }

                HorizontalDivider(
                    modifier = Modifier.padding(bottom = 12.dp),
                    color = MaterialTheme.colorScheme.outlineVariant.copy(alpha = 0.5f)
                )

                // Summary Content
                if (uiState.displaySummary.isEmpty() && uiState.transcription.isEmpty() && !uiState.isThinking) {
                    Box(modifier = Modifier.weight(1f), contentAlignment = androidx.compose.ui.Alignment.Center) {
                        Text(
                            text = "Record audio to generate a summary...",
                            style = MaterialTheme.typography.bodyMedium,
                            color = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.5f),
                            textAlign = androidx.compose.ui.text.style.TextAlign.Center
                        )
                    }
                } else {
                    androidx.compose.foundation.lazy.LazyColumn(
                        modifier = Modifier.weight(1f).fillMaxWidth(),
                        verticalArrangement = Arrangement.spacedBy(16.dp)
                    ) {
                        if (uiState.displaySummary.isNotEmpty()) {
                            item {
                                Text(
                                    text = "✨ Summary",
                                    style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                                    color = MaterialTheme.colorScheme.primary
                                )
                                Spacer(modifier = Modifier.height(8.dp))
                                Card(
                                    colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.5f)),
                                    shape = androidx.compose.foundation.shape.RoundedCornerShape(12.dp)
                                ) {
                                    MarkdownText(
                                        markdown = uiState.displaySummary,
                                        modifier = Modifier.padding(16.dp),
                                        style = MaterialTheme.typography.bodyMedium.copy(
                                            lineHeight = 22.sp,
                                            color = MaterialTheme.colorScheme.onSurfaceVariant
                                        ),
                                        syntaxHighlightColor = MaterialTheme.colorScheme.primary,
                                        linkColor = MaterialTheme.colorScheme.primary
                                    )
                                }

                                if (uiState.pdfUrl != null) {
                                    Spacer(modifier = Modifier.height(12.dp))
                                    Button(
                                        onClick = {
                                            val fullUrl = "https://teralux.farismunir.my.id" + uiState.pdfUrl
                                            val intent = Intent(Intent.ACTION_VIEW, Uri.parse(fullUrl))
                                            context.startActivity(intent)
                                        },
                                        modifier = Modifier.fillMaxWidth(),
                                        shape = androidx.compose.foundation.shape.RoundedCornerShape(8.dp),
                                        colors = ButtonDefaults.buttonColors(
                                            containerColor = MaterialTheme.colorScheme.primaryContainer,
                                            contentColor = MaterialTheme.colorScheme.onPrimaryContainer
                                        )
                                    ) {
                                        Icon(Icons.Default.Download, contentDescription = null)
                                        Spacer(modifier = Modifier.width(8.dp))
                                        Text("Download PDF Report")
                                    }
                                }
                            }
                        }

                        if (uiState.refinedText.isNotEmpty()) {
                            item {
                                Text(
                                    text = "📝 Refined Transcription",
                                    style = MaterialTheme.typography.titleSmall.copy(fontWeight = FontWeight.Bold),
                                    color = MaterialTheme.colorScheme.secondary
                                )
                                Spacer(modifier = Modifier.height(4.dp))
                                Text(
                                    text = uiState.refinedText,
                                    style = MaterialTheme.typography.bodySmall,
                                    color = MaterialTheme.colorScheme.onSurfaceVariant
                                )
                            }
                        }
                    }
                }
            }
        },
        walkthroughContent = {
            Text(
                text = "🎯 Purpose",
                style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                color = MaterialTheme.colorScheme.primary
            )
            Text(
                text = "Long-Form Meeting & Audio Transcription (30 minutes to 4+ hours)",
                style = MaterialTheme.typography.bodyMedium
            )

            Text(
                text = "🔄 Flow Summary (Normal Upload)",
                style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                color = MaterialTheme.colorScheme.primary
            )
            Text(
                text = "1. The Android app (Kotlin) records audio using the microphone.\n" +
                        "2. The audio is saved as a complete audio file.\n" +
                        "3. The app sends the file to the backend (Golang) via REST API.\n" +
                        "4. The backend receives the audio file.\n" +
                        "5. The backend processes the file using whisper.cpp.\n" +
                        "6. The system returns the transcribed text.",
                style = MaterialTheme.typography.bodyMedium
            )

            Text(
                text = "✅ Advantages",
                style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                color = MaterialTheme.colorScheme.primary
            )
            Text(
                text = "• Simple architecture & easy to debug.\n" +
                        "• Stable for very long recordings.\n" +
                        "• No infrastructure message broker needed.\n" +
                        "• Better transcription consistency.",
                style = MaterialTheme.typography.bodyMedium
            )

            Text(
                text = "❌ Disadvantages",
                style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                color = MaterialTheme.colorScheme.error
            )
            Text(
                text = "• Not real-time (must wait until finish).\n" +
                        "• Higher perceived latency for long sessions.\n" +
                        "• Large file uploads required.\n" +
                        "• Requires stable internet connection.",
                style = MaterialTheme.typography.bodyMedium
            )
        }
    )
}

@Composable
fun InternalFileSelectionDialog(
    files: List<java.io.File>,
    onFileSelected: (java.io.File) -> Unit,
    onDismiss: () -> Unit
) {
    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("Milih File Manual (Quick select)") },
        text = {
            if (files.isEmpty()) {
                Text("Nggak nemu file audio di folder Download emulator lo bro. Pastiin udah di-push.")
            } else {
                LazyColumn(
                    modifier = Modifier.fillMaxWidth().heightIn(max = 400.dp),
                    verticalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    items(files) { file ->
                        Card(
                            onClick = { onFileSelected(file) },
                            modifier = Modifier.fillMaxWidth(),
                            colors = CardDefaults.cardColors(
                                containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.5f)
                            )
                        ) {
                            Row(
                                modifier = Modifier.padding(12.dp),
                                verticalAlignment = Alignment.CenterVertically
                            ) {
                                Icon(
                                    imageVector = Icons.Default.AudioFile,
                                    contentDescription = null,
                                    tint = MaterialTheme.colorScheme.primary,
                                    modifier = Modifier.size(24.dp)
                                )
                                Spacer(modifier = Modifier.width(12.dp))
                                Column {
                                    Text(
                                        text = file.name,
                                        style = MaterialTheme.typography.bodyMedium,
                                        fontWeight = FontWeight.Bold,
                                        maxLines = 1
                                    )
                                    Text(
                                        text = "${file.length() / 1024} KB",
                                        style = MaterialTheme.typography.labelSmall,
                                        color = MaterialTheme.colorScheme.onSurfaceVariant
                                    )
                                }
                            }
                        }
                    }
                }
            }
        },
        confirmButton = {
            TextButton(onClick = onDismiss) {
                Text("Cancel")
            }
        }
    )
}
