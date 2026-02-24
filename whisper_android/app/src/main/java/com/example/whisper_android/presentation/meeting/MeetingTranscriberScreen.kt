package com.example.whisper_android.presentation.meeting

import android.Manifest
import android.content.pm.PackageManager
import android.net.Uri
import android.webkit.MimeTypeMap
import android.widget.Toast
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.animation.core.FastOutSlowInEasing
import androidx.compose.animation.core.LinearEasing
import androidx.compose.animation.core.RepeatMode
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.tween
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.systemBars
import androidx.compose.material3.Scaffold
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import androidx.core.content.ContextCompat
import com.example.whisper_android.domain.usecase.MeetingProcessState
import com.example.whisper_android.presentation.components.EmailInputDialog
import com.example.whisper_android.presentation.components.FeatureBackground
import com.example.whisper_android.presentation.components.FeatureHeader
import com.example.whisper_android.presentation.components.FeatureMainCard
import com.example.whisper_android.presentation.components.UiState
import com.example.whisper_android.presentation.meeting.components.MeetingControlPill
import com.example.whisper_android.presentation.meeting.components.MeetingErrorContent
import com.example.whisper_android.presentation.meeting.components.MeetingHeaderControls
import com.example.whisper_android.presentation.meeting.components.MeetingIdleContent
import com.example.whisper_android.presentation.meeting.components.MeetingLoadingContent
import com.example.whisper_android.presentation.meeting.components.MeetingRecordingContent
import com.example.whisper_android.presentation.meeting.components.MeetingSuccessContent
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

@Composable
fun MeetingTranscriberScreen(
    onNavigateBack: () -> Unit,
    viewModel: MeetingViewModel =
        remember {
            MeetingViewModel(
                com.example.whisper_android.data.di.NetworkModule.processMeetingUseCase
            )
        }
) {
    val uiState by viewModel.uiState.collectAsState()
    val emailState by viewModel.emailState.collectAsState()
    var summaryLanguage by remember { mutableStateOf("id") }
    var showEmailDialog by remember { mutableStateOf(false) }

    val context = LocalContext.current
    val scope = rememberCoroutineScope()

    val token =
        remember {
            com.example.whisper_android.data.di.NetworkModule.tokenManager
                .getAccessToken() ?: ""
        }

    val infiniteTransition = rememberInfiniteTransition(label = "pulse")
    val pulseScale by infiniteTransition.animateFloat(
        initialValue = 1f,
        targetValue = 1.15f,
        animationSpec = infiniteRepeatable(
            tween(1200, easing = FastOutSlowInEasing),
            RepeatMode.Reverse
        ),
        label = "pulseScale"
    )
    val glowAlpha by infiniteTransition.animateFloat(
        initialValue = 0.3f,
        targetValue = 0.8f,
        animationSpec = infiniteRepeatable(tween(1500, easing = LinearEasing), RepeatMode.Reverse),
        label = "glowAlpha"
    )

    val audioRecorder = remember { AudioRecorder(context) }
    var isRecording by remember { mutableStateOf(false) }
    var audioFile by remember { mutableStateOf<java.io.File?>(null) }

    LaunchedEffect(emailState) {
        when (emailState) {
            is UiState.Success -> {
                Toast.makeText(context, "Email sent successfully", Toast.LENGTH_SHORT).show()
                viewModel.resetEmailState()
            }

            is UiState.Error -> {
                Toast.makeText(context, (emailState as UiState.Error).message, Toast.LENGTH_LONG).show()
                viewModel.resetEmailState()
            }

            else -> {}
        }
    }

    val launcher =
        rememberLauncherForActivityResult(ActivityResultContracts.GetContent()) { uri: Uri? ->
            uri?.let { selectedUri ->
                scope.launch(Dispatchers.IO) {
                    val contentResolver = context.contentResolver
                    val type = contentResolver.getType(selectedUri)
                    val extension = MimeTypeMap.getSingleton().getExtensionFromMimeType(type) ?: "m4a"
                    val outputFile = java.io.File(context.cacheDir, "upload_audio.$extension")
                    try {
                        contentResolver.openInputStream(selectedUri)?.use { input ->
                            outputFile.outputStream().use { output -> input.copyTo(output) }
                        }
                        withContext(Dispatchers.Main) {
                            audioFile = outputFile
                            if (token.isNotEmpty()) {
                                viewModel.processRecording(
                                    outputFile,
                                    token,
                                    summaryLanguage
                                )
                            }
                        }
                    } catch (e: Exception) {
                        e.printStackTrace()
                    }
                }
            }
        }

    val permissionLauncher = rememberLauncherForActivityResult(
        ActivityResultContracts.RequestPermission()
    ) { }

    val storagePermissionLauncher =
        rememberLauncherForActivityResult(ActivityResultContracts.RequestPermission()) { isGranted ->
            if (isGranted) {
                Toast.makeText(
                    context,
                    "Permission granted. Tap PDF again to download.",
                    Toast.LENGTH_SHORT
                ).show()
            } else {
                Toast.makeText(
                    context,
                    "Storage permission required to save PDF.",
                    Toast.LENGTH_LONG
                ).show()
            }
        }

    val hasPermission = ContextCompat.checkSelfPermission(context, Manifest.permission.RECORD_AUDIO) == PackageManager.PERMISSION_GRANTED

    FeatureBackground {
        Scaffold(
            containerColor = Color.Transparent,
            contentWindowInsets = WindowInsets.systemBars,
            topBar = { FeatureHeader(title = "Meeting Insights", onNavigateBack = onNavigateBack) }
        ) { paddingValues ->
            Box(modifier = Modifier.fillMaxSize().padding(paddingValues)) {
                Column(
                    modifier = Modifier.fillMaxSize().padding(horizontal = 4.dp, vertical = 2.dp).padding(
                        bottom = 60.dp
                    ),
                    horizontalAlignment = Alignment.CenterHorizontally
                ) {
                    FeatureMainCard(modifier = Modifier.weight(1f)) {
                        Column(modifier = Modifier.fillMaxSize()) {
                            MeetingHeaderControls(
                                uiState = uiState,
                                emailState = emailState,
                                summaryLanguage = summaryLanguage,
                                onLanguageSelected = { summaryLanguage = it },
                                onDownloadClick = { url ->
                                    if (android.os.Build.VERSION.SDK_INT < android.os.Build.VERSION_CODES.TIRAMISU) {
                                        if (ContextCompat.checkSelfPermission(
                                                context,
                                                Manifest.permission.WRITE_EXTERNAL_STORAGE
                                            ) ==
                                            PackageManager.PERMISSION_GRANTED
                                        ) {
                                            downloadPdf(
                                                context,
                                                url,
                                                "Meeting_Summary_${System.currentTimeMillis()}"
                                            )
                                        } else {
                                            storagePermissionLauncher.launch(
                                                Manifest.permission.WRITE_EXTERNAL_STORAGE
                                            )
                                        }
                                    } else {
                                        downloadPdf(
                                            context,
                                            url,
                                            "Meeting_Summary_${System.currentTimeMillis()}"
                                        )
                                    }
                                },
                                onEmailClick = { showEmailDialog = true }
                            )

                            Box(
                                modifier = Modifier.weight(1f).fillMaxWidth(),
                                contentAlignment = Alignment.Center
                            ) {
                                when (uiState) {
                                    is MeetingProcessState.Idle -> MeetingIdleContent()
                                    is MeetingProcessState.Recording -> MeetingRecordingContent()
                                    is MeetingProcessState.Success -> MeetingSuccessContent(
                                        uiState as MeetingProcessState.Success
                                    )
                                    is MeetingProcessState.Error -> MeetingErrorContent(
                                        uiState as MeetingProcessState.Error
                                    )
                                    else -> MeetingLoadingContent(uiState, glowAlpha)
                                }
                            }
                        }
                    }
                }

                MeetingControlPill(
                    isRecording = isRecording,
                    hasPermission = hasPermission,
                    uiState = uiState,
                    pulseScale = pulseScale,
                    onMicClick = {
                        if (!isRecording &&
                            (
                                uiState is MeetingProcessState.Idle || uiState is MeetingProcessState.Success ||
                                    uiState is MeetingProcessState.Error
                                )
                        ) {
                            if (!hasPermission) {
                                permissionLauncher.launch(Manifest.permission.RECORD_AUDIO)
                            } else {
                                val file = java.io.File(context.cacheDir, "meeting_audio.wav")
                                audioRecorder.start(file)
                                audioFile = file
                                isRecording = true
                                viewModel.resetState()
                            }
                        }
                    },
                    onUploadClick = {
                        if (uiState is MeetingProcessState.Idle || uiState is MeetingProcessState.Success ||
                            uiState is MeetingProcessState.Error
                        ) {
                            launcher.launch("audio/*")
                        }
                    },
                    onStopClick = {
                        audioRecorder.stop()
                        audioFile?.let { file -> audioRecorder.finalizeWav(file) }
                        isRecording = false
                        audioFile?.let { file ->
                            if (token.isNotEmpty()) {
                                viewModel.processRecording(
                                    file,
                                    token,
                                    summaryLanguage
                                )
                            }
                        }
                    },
                    onClearClick = {
                        if (isRecording) {
                            audioRecorder.stop()
                            audioFile?.let { file -> audioRecorder.finalizeWav(file) }
                        }
                        isRecording = false
                        viewModel.resetState()
                    },
                    modifier = Modifier.align(Alignment.BottomCenter)
                )
            }
        }

        if (showEmailDialog) {
            EmailInputDialog(
                onDismiss = { showEmailDialog = false },
                onSend = { email, subject ->
                    viewModel.sendEmailSummary(email, subject, summaryLanguage)
                    showEmailDialog = false
                }
            )
        }
    }
}
