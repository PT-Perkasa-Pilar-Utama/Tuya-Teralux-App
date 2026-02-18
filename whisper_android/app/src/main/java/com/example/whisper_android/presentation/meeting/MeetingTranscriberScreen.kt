package com.example.whisper_android.presentation.meeting

import androidx.compose.foundation.background
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
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import android.net.Uri
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import android.webkit.MimeTypeMap
import androidx.compose.material.icons.filled.FolderOpen
import android.content.Intent
import androidx.compose.animation.core.*
import androidx.compose.foundation.Canvas
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Shadow
import android.app.DownloadManager
import android.os.Environment
import android.widget.Toast

fun downloadPdf(context: android.content.Context, url: String, title: String) {
    try {
        // Ensure standard HTTP/HTTPS URL for DownloadManager
        val absoluteUrl = if (url.startsWith("/")) {
            val base = com.example.whisper_android.data.di.NetworkModule.BASE_URL.removeSuffix("/")
            "$base$url"
        } else {
            url
        }

        val request = DownloadManager.Request(Uri.parse(absoluteUrl))
            .setTitle(title)
            .setDescription("Downloading meeting summary PDF...")
            .setNotificationVisibility(DownloadManager.Request.VISIBILITY_VISIBLE_NOTIFY_COMPLETED)
            .setDestinationInExternalPublicDir(Environment.DIRECTORY_DOWNLOADS, "${title.replace(" ", "_")}.pdf")
            .setAllowedOverMetered(true)
            .setAllowedOverRoaming(true)

        val downloadManager = context.getSystemService(android.content.Context.DOWNLOAD_SERVICE) as DownloadManager
        downloadManager.enqueue(request)
        Toast.makeText(context, "Download started...", Toast.LENGTH_SHORT).show()

        // Auto-open PDF in browser/viewer
        val intent = android.content.Intent(android.content.Intent.ACTION_VIEW, Uri.parse(absoluteUrl))
        context.startActivity(intent)
    } catch (e: Exception) {
        Toast.makeText(context, "Download failed: ${e.message}", Toast.LENGTH_LONG).show()
    }
}

@Composable
fun MeetingTranscriberScreen(
    onNavigateBack: () -> Unit,
    viewModel: MeetingViewModel = remember { 
        MeetingViewModel(com.example.whisper_android.data.di.NetworkModule.processMeetingUseCase) 
    }
) {
    val uiState by viewModel.uiState.collectAsState()
    var summaryLanguage by remember { mutableStateOf("id") } // Default to "id" (Indonesian)
    
    val context = LocalContext.current
    val primaryColor = MaterialTheme.colorScheme.primary
    val scope = rememberCoroutineScope()
    
    // Get Token (Simplified: Fetch from TokenManager directly for this scope)
    val token = remember { 
        com.example.whisper_android.data.di.NetworkModule.tokenManager.getAccessToken() ?: "" 
    }

    // Animations for Pulse and Glow
    val infiniteTransition = rememberInfiniteTransition(label = "pulse")
    val pulseScale by infiniteTransition.animateFloat(
        initialValue = 1f,
        targetValue = 1.15f,
        animationSpec = infiniteRepeatable(
            animation = tween(1200, easing = FastOutSlowInEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "pulseScale"
    )
    val glowAlpha by infiniteTransition.animateFloat(
        initialValue = 0.3f,
        targetValue = 0.8f,
        animationSpec = infiniteRepeatable(
            animation = tween(1500, easing = LinearEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "glowAlpha"
    )

    // Audio Recorder Setup
    val audioRecorder = remember { AudioRecorder(context) }
    var isRecording by remember { mutableStateOf(false) }
    var audioFile by remember { mutableStateOf<java.io.File?>(null) }

    // File Picker Launcher
    val launcher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.GetContent()
    ) { uri: Uri? ->
        uri?.let { selectedUri ->
            scope.launch(Dispatchers.IO) {
                val contentResolver = context.contentResolver
                val type = contentResolver.getType(selectedUri)
                val extension = MimeTypeMap.getSingleton().getExtensionFromMimeType(type) ?: "m4a"
                val outputFile = java.io.File(context.cacheDir, "upload_audio.$extension")
                
                try {
                    contentResolver.openInputStream(selectedUri)?.use { input ->
                        outputFile.outputStream().use { output ->
                            input.copyTo(output)
                        }
                    }
                    
                    withContext(Dispatchers.Main) {
                        audioFile = outputFile
                        if (token.isNotEmpty()) {
                            viewModel.processRecording(outputFile, token, summaryLanguage)
                        }
                    }
                } catch (e: Exception) {
                    e.printStackTrace()
                }
            }
        }
    }
    // Audio Permission Launcher
    val permissionLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.RequestPermission()
    ) { isGranted ->
        if (isGranted) {
            // Updated hasPermission will activate UI
        }
    }

    // Storage Permission Launcher for PDF Download (Android < 13)
    val storagePermissionLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.RequestPermission()
    ) { isGranted ->
        if (isGranted) {
            Toast.makeText(context, "Permission granted. Tap PDF again to download.", Toast.LENGTH_SHORT).show()
        } else {
            Toast.makeText(context, "Storage permission required to save PDF.", Toast.LENGTH_LONG).show()
        }
    }

    // Permission Check
    val hasPermission = ContextCompat.checkSelfPermission(
        context,
        Manifest.permission.RECORD_AUDIO
    ) == PackageManager.PERMISSION_GRANTED

    FeatureBackground {
        Scaffold(
            containerColor = Color.Transparent,
            contentWindowInsets = WindowInsets.systemBars,
            topBar = {
                // Header
                FeatureHeader(
                    title = "Meeting Insights",
                    onNavigateBack = onNavigateBack
                )
            }
        ) { padding ->
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(padding)
                    .padding(horizontal = 4.dp, vertical = 2.dp),
                horizontalAlignment = Alignment.CenterHorizontally
            ) {
                // Main Transcription Card
                FeatureMainCard(
                    modifier = Modifier.weight(1f)
                ) {
                    Column(
                        modifier = Modifier.fillMaxSize()
                    ) {
                        // Header Controls (Download + Language) - Always visible at top
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(bottom = 8.dp),
                            horizontalArrangement = Arrangement.SpaceBetween,
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            // PDF Download Button (Show only on Success)
                            if (uiState is com.example.whisper_android.domain.usecase.MeetingProcessState.Success) {
                                val state = uiState as com.example.whisper_android.domain.usecase.MeetingProcessState.Success
                                if (state.pdfUrl != null) {
                                    Button(
                                        onClick = { 
                                            // Check for permission if SDK < 33 (Android 13)
                                            if (android.os.Build.VERSION.SDK_INT < android.os.Build.VERSION_CODES.TIRAMISU) {
                                                if (ContextCompat.checkSelfPermission(context, Manifest.permission.WRITE_EXTERNAL_STORAGE) == PackageManager.PERMISSION_GRANTED) {
                                                    downloadPdf(context, state.pdfUrl, "Meeting_Summary_${System.currentTimeMillis()}")
                                                } else {
                                                    storagePermissionLauncher.launch(Manifest.permission.WRITE_EXTERNAL_STORAGE)
                                                }
                                            } else {
                                                // Android 13+ doesn't need WRITE_EXTERNAL_STORAGE for public downloads
                                                downloadPdf(context, state.pdfUrl, "Meeting_Summary_${System.currentTimeMillis()}")
                                            }
                                        },
                                        modifier = Modifier.height(24.dp),
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
                            } else {
                                Spacer(modifier = Modifier.width(1.dp))
                            }

                            LanguagePillToggle(
                                selectedLanguage = summaryLanguage,
                                onLanguageSelected = { summaryLanguage = it }
                            )
                        }
                        
                        // Content Area (Takes remaining space)
                        Box(
                            modifier = Modifier
                                .weight(1f)
                                .fillMaxWidth(),
                            contentAlignment = Alignment.Center // Default center for Loading/Idle
                        ) {
                            when (uiState) {
                                is com.example.whisper_android.domain.usecase.MeetingProcessState.Idle -> {
                                    Column(
                                        horizontalAlignment = Alignment.CenterHorizontally,
                                        verticalArrangement = Arrangement.Center
                                    ) {
                                        // Subtle Waveform Cue
                                        Row(
                                            modifier = Modifier.height(40.dp),
                                            horizontalArrangement = Arrangement.spacedBy(4.dp),
                                            verticalAlignment = Alignment.CenterVertically
                                        ) {
                                            repeat(5) { index ->
                                                Box(
                                                    modifier = Modifier
                                                        .width(4.dp)
                                                        .height(if (index % 2 == 0) 24.dp else 16.dp)
                                                        .background(
                                                            MaterialTheme.colorScheme.primary.copy(alpha = 0.2f),
                                                            RoundedCornerShape(2.dp)
                                                        )
                                                )
                                            }
                                        }
                                        Spacer(modifier = Modifier.height(24.dp))
                                        Text(
                                            text = "Ready to capture your next breakthrough.",
                                            style = MaterialTheme.typography.titleMedium,
                                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.4f),
                                            textAlign = TextAlign.Center,
                                            fontWeight = FontWeight.Medium
                                        )
                                        Spacer(modifier = Modifier.height(8.dp))
                                        Text(
                                            text = "Tap the mic to start recording your meeting.",
                                            style = MaterialTheme.typography.bodySmall,
                                            color = MaterialTheme.colorScheme.onSurface.copy(alpha = 0.3f),
                                            textAlign = TextAlign.Center
                                        )
                                    }
                                }
                                is com.example.whisper_android.domain.usecase.MeetingProcessState.Recording -> {
                                    Text(
                                        text = "Recording...",
                                        style = MaterialTheme.typography.headlineSmall,
                                        color = MaterialTheme.colorScheme.error,
                                        fontWeight = FontWeight.Bold,
                                        textAlign = TextAlign.Center
                                    )
                                }
                                is com.example.whisper_android.domain.usecase.MeetingProcessState.Success -> {
                                    val state = uiState as com.example.whisper_android.domain.usecase.MeetingProcessState.Success
                                    Column(
                                        modifier = Modifier
                                            .fillMaxSize()
                                            .verticalScroll(rememberScrollState())
                                    ) {
                                        Text(
                                            text = "Meeting Summary",
                                            style = MaterialTheme.typography.titleLarge,
                                            color = MaterialTheme.colorScheme.primary,
                                            fontWeight = FontWeight.Bold,
                                            modifier = Modifier.padding(bottom = 2.dp)
                                        )
                                        
                                        MarkdownText(
                                            markdown = state.summary
                                                .replace(Regex("^-+\\s*$", RegexOption.MULTILINE), "")
                                                .replace(Regex("^.*–.*$", RegexOption.MULTILINE), "")
                                                .replace("\n\n\n", "\n\n")
                                                .replace(Regex("\n{3,}"), "\n\n")
                                                .trim(),
                                            style = MaterialTheme.typography.bodyLarge.copy(
                                                color = Color.DarkGray,
                                                fontSize = 13.sp,
                                                lineHeight = 16.sp
                                            ),
                                            modifier = Modifier.fillMaxWidth()
                                        )
                                        Spacer(modifier = Modifier.height(16.dp))
                                    }
                                }
                                is com.example.whisper_android.domain.usecase.MeetingProcessState.Error -> {
                                    val state = uiState as com.example.whisper_android.domain.usecase.MeetingProcessState.Error
                                    Text(
                                        text = "Error: ${state.message}",
                                        style = MaterialTheme.typography.bodyLarge,
                                        color = MaterialTheme.colorScheme.error,
                                        textAlign = TextAlign.Center
                                    )
                                }
                                else -> {
                                    // Loading States - Perfectly Centered
                                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                                        Box(contentAlignment = Alignment.Center) {
                                            // Glow effect (Primary themed)
                                            Canvas(modifier = Modifier.size(100.dp)) {
                                                drawCircle(
                                                    color = primaryColor,
                                                    alpha = glowAlpha * 0.15f,
                                                    radius = size.minDimension / 2
                                                )
                                            }
                                            CircularProgressIndicator(
                                                color = primaryColor,
                                                strokeWidth = 3.dp,
                                                modifier = Modifier.size(56.dp)
                                            )
                                        }
                                        Spacer(modifier = Modifier.height(8.dp))
                                        Text(
                                            text = when(uiState) {
                                                com.example.whisper_android.domain.usecase.MeetingProcessState.Uploading -> "Securely Uploading..."
                                                com.example.whisper_android.domain.usecase.MeetingProcessState.Transcribing -> "AI Transcribing..."
                                                com.example.whisper_android.domain.usecase.MeetingProcessState.Translating -> "Translating Context..."
                                                com.example.whisper_android.domain.usecase.MeetingProcessState.Summarizing -> "Generating Insights..."
                                                else -> "Thinking..."
                                            },
                                            style = MaterialTheme.typography.titleMedium,
                                            color = primaryColor,
                                            fontWeight = FontWeight.Black,
                                            letterSpacing = 0.5.sp
                                        )
                                    }
                                }
                            }
                        }
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
                            isProcessing = uiState !is com.example.whisper_android.domain.usecase.MeetingProcessState.Idle && 
                                           uiState !is com.example.whisper_android.domain.usecase.MeetingProcessState.Success &&
                                           uiState !is com.example.whisper_android.domain.usecase.MeetingProcessState.Error &&
                                           uiState !is com.example.whisper_android.domain.usecase.MeetingProcessState.Recording,
                            onClick = { 
                                if (!isRecording && (uiState is com.example.whisper_android.domain.usecase.MeetingProcessState.Idle || 
                                    uiState is com.example.whisper_android.domain.usecase.MeetingProcessState.Success || 
                                    uiState is com.example.whisper_android.domain.usecase.MeetingProcessState.Error)) {
                                    
                                    if (!hasPermission) {
                                        permissionLauncher.launch(Manifest.permission.RECORD_AUDIO)
                                    } else {
                                        // Start Recording
                                        val file = java.io.File(context.cacheDir, "meeting_audio.wav")
                                        audioRecorder.start(file)
                                        audioFile = file
                                        isRecording = true
                                        viewModel.resetState() // Clear previous results
                                    }
                                }
                            },
                            size = 36.dp,
                            modifier = Modifier.graphicsLayer {
                                if (isRecording) {
                                    scaleX = pulseScale
                                    scaleY = pulseScale
                                }
                            }
                        )

                        // Upload Button
                        if (!isRecording) {
                             IconButton(
                                onClick = { 
                                    if (uiState is com.example.whisper_android.domain.usecase.MeetingProcessState.Idle || 
                                        uiState is com.example.whisper_android.domain.usecase.MeetingProcessState.Success || 
                                        uiState is com.example.whisper_android.domain.usecase.MeetingProcessState.Error) {
                                        launcher.launch("audio/*")
                                    }
                                },
                                modifier = Modifier.size(32.dp)
                            ) {
                                Icon(
                                    imageVector = Icons.Default.FolderOpen,
                                    contentDescription = "Upload Audio",
                                    tint = MaterialTheme.colorScheme.primary,
                                    modifier = Modifier.size(24.dp)
                                )
                            }
                        }

                        // Stop Button (Processing Trigger)
                        if (isRecording) {
                            IconButton(
                                onClick = { 
                                    audioRecorder.stop()
                                    audioFile?.let { file ->
                                        audioRecorder.finalizeWav(file)
                                    }
                                    isRecording = false
                                    audioFile?.let { file ->
                                        if (token.isNotEmpty()) {
                                            viewModel.processRecording(file, token, summaryLanguage)
                                        } else {
                                            // Handle missing token?
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
                                if (isRecording) {
                                    audioRecorder.stop()
                                    audioFile?.let { file ->
                                        audioRecorder.finalizeWav(file)
                                    }
                                }
                                isRecording = false
                                viewModel.resetState()
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
                                isRecording -> "Live"
                                uiState !is com.example.whisper_android.domain.usecase.MeetingProcessState.Idle && 
                                uiState !is com.example.whisper_android.domain.usecase.MeetingProcessState.Success &&
                                uiState !is com.example.whisper_android.domain.usecase.MeetingProcessState.Error -> "AI Active"
                                else -> "Ready"
                            },
                            fontSize = 12.sp,
                            color = if (isRecording) MaterialTheme.colorScheme.error else MaterialTheme.colorScheme.primary,
                            fontWeight = FontWeight.Black,
                            letterSpacing = 1.sp,
                            modifier = Modifier.padding(end = 4.dp)
                        )
                    }
                }
            }
        }
    }
}
