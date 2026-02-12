package com.example.whisper_android.presentation.upload

import android.content.Intent
import android.net.Uri
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.AudioFile
import androidx.compose.material.icons.filled.Description
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
import androidx.compose.foundation.background
import androidx.compose.foundation.shape.CircleShape
import com.example.whisper_android.data.di.NetworkModule
import com.example.whisper_android.presentation.components.FeatureScreenTemplate
import com.example.whisper_android.presentation.components.ToastObserver
import com.example.whisper_android.MainActivity
import com.example.whisper_android.presentation.components.*

import dev.jeziellago.compose.markdowntext.MarkdownText

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

    // Helper for specialized download
    fun downloadPdf(url: String, fileName: String) {
        try {
            val request = android.app.DownloadManager.Request(Uri.parse(url))
                .setTitle(fileName)
                .setDescription("Downloading Meeting Report")
                .setNotificationVisibility(android.app.DownloadManager.Request.VISIBILITY_VISIBLE_NOTIFY_COMPLETED)
                .setDestinationInExternalPublicDir(android.os.Environment.DIRECTORY_DOWNLOADS, fileName)
                .setAllowedOverMetered(true)
                .setAllowedOverRoaming(true)

            val dm = context.getSystemService(android.content.Context.DOWNLOAD_SERVICE) as android.app.DownloadManager
            dm.enqueue(request)
            android.widget.Toast.makeText(context, "Downloading report...", android.widget.Toast.LENGTH_SHORT).show()
        } catch (e: Exception) {
            // Fallback to browser if DownloadManager fails
            val intent = Intent(Intent.ACTION_VIEW, Uri.parse(url))
            context.startActivity(intent)
        }
    }
    
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
        onStopClick = {
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

                // Modern Language Selector (Subtle Pill Replaced by Component)
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(bottom = 16.dp),
                    horizontalArrangement = Arrangement.End,
                    verticalAlignment = androidx.compose.ui.Alignment.CenterVertically
                ) {
                    LanguagePillToggle(
                        selectedLanguage = uiState.summaryLanguage,
                        onLanguageSelected = { viewModel.setSummaryLanguage(it) }
                    )
                }

                // Summary Content (Replaced by Component)
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
                                SummaryCard(
                                    summary = uiState.displaySummary,
                                    pdfUrl = uiState.pdfUrl,
                                    onDownloadPdf = { pdfUrl ->
                                        val fullUrl = NetworkModule.BASE_URL.removeSuffix("/") + pdfUrl
                                        val fileName = pdfUrl.split("/").last()
                                        downloadPdf(fullUrl, fileName)
                                    }
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
