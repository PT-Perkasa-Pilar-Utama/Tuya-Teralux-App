package com.example.whisperandroid.presentation.meeting

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
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
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
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import androidx.core.content.ContextCompat
import com.example.whisperandroid.domain.usecase.MeetingProcessState
import com.example.whisperandroid.presentation.components.EmailInputDialog
import com.example.whisperandroid.presentation.components.LanguagePillToggle
import com.example.whisperandroid.presentation.components.MqttStatusBadge
import com.example.whisperandroid.presentation.components.SensioFeatureLayout
import com.example.whisperandroid.presentation.components.UiState
import com.example.whisperandroid.presentation.meeting.components.MeetingControlPill
import com.example.whisperandroid.presentation.meeting.components.MeetingErrorContent
import com.example.whisperandroid.presentation.meeting.components.MeetingFilePickerSheet
import com.example.whisperandroid.presentation.meeting.components.MeetingHeaderControls
import com.example.whisperandroid.presentation.meeting.components.MeetingIdleContent
import com.example.whisperandroid.presentation.meeting.components.MeetingLoadingContent
import com.example.whisperandroid.presentation.meeting.components.MeetingRecordingContent
import com.example.whisperandroid.presentation.meeting.components.MeetingSuccessContent
import com.example.whisperandroid.util.DeviceUtils
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

@Composable
fun MeetingTranscriberScreen(
    onNavigateBack: () -> Unit,
    viewModel: MeetingViewModel = androidx.lifecycle.viewmodel.compose.viewModel(
        factory = MeetingViewModelFactory(
            com.example.whisperandroid.data.di.NetworkModule.processMeetingUseCase
        )
    )
) {
    val uiState by viewModel.uiState.collectAsState()
    val emailState by viewModel.emailState.collectAsState()
    val mqttStatus by viewModel.mqttStatus.collectAsState()
    var summaryLanguage by androidx.compose.runtime.saveable.rememberSaveable {
        mutableStateOf("id")
    }
    var showEmailDialog by androidx.compose.runtime.saveable.rememberSaveable {
        mutableStateOf(
            false
        )
    }
    var showFilePickerSheet by androidx.compose.runtime.saveable.rememberSaveable {
        mutableStateOf(
            false
        )
    }
    var hasAutoSent by androidx.compose.runtime.saveable.rememberSaveable { mutableStateOf(false) }

    val context = LocalContext.current
    val scope = rememberCoroutineScope()

    // Auto-connect to MQTT when screen is mounted
    LaunchedEffect(Unit) {
        viewModel.reconnectMqtt(DeviceUtils.getDeviceId(context))
    }

    // Reset auto-send flag when process starts (Idle/Recording)
    LaunchedEffect(uiState) {
        if (uiState is MeetingProcessState.Idle || uiState is MeetingProcessState.Recording) {
            hasAutoSent = false
        }
    }

    // Auto-send summary when success
    LaunchedEffect(uiState) {
        if (uiState is MeetingProcessState.Success && !hasAutoSent) {
            val deviceId = DeviceUtils.getDeviceId(context)
            viewModel.sendEmailSummaryByMac(
                macAddress = deviceId,
                subject = "Auto-generated",
                targetLang = summaryLanguage
            )
            hasAutoSent = true
        }
    }

    val token =
        remember {
            com.example.whisperandroid.data.di.NetworkModule.tokenManager
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
        animationSpec = infiniteRepeatable(
            tween(1500, easing = LinearEasing),
            RepeatMode.Reverse
        ),
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
                val errorMsg = (emailState as UiState.Error).message
                Toast.makeText(context, errorMsg, Toast.LENGTH_LONG).show()
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
                    val extension = MimeTypeMap.getSingleton()
                        .getExtensionFromMimeType(type) ?: "m4a"
                    val outputFile =
                        com.example.whisperandroid.data.local.MeetingAudioFileStore
                            .createImportedAudioFile(context, extension)
                    try {
                        contentResolver.openInputStream(selectedUri)?.use { input ->
                            outputFile.outputStream().use { output -> input.copyTo(output) }
                        }
                        withContext(Dispatchers.Main) {
                            audioFile = outputFile
                            if (token.isNotEmpty()) {
                                viewModel.processRecording(
                                    context,
                                    outputFile,
                                    token,
                                    summaryLanguage,
                                    DeviceUtils.getDeviceId(context)
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
        rememberLauncherForActivityResult(
            ActivityResultContracts.RequestPermission()
        ) { isGranted ->
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

    val hasPermission = ContextCompat.checkSelfPermission(
        context,
        Manifest.permission.RECORD_AUDIO
    ) == PackageManager.PERMISSION_GRANTED

    SensioFeatureLayout(
        title = "Meeting Insights",
        onNavigateBack = onNavigateBack,
        titleTestTag = "meeting_screen_title",
        headerActions = {
            Row(
                verticalAlignment = Alignment.CenterVertically,
                modifier = Modifier.padding(end = 8.dp)
            ) {
                LanguagePillToggle(
                    selectedLanguage = summaryLanguage,
                    onLanguageSelected = { summaryLanguage = it }
                )

                MqttStatusBadge(
                    status = mqttStatus,
                    onReconnectClick = {
                        viewModel.reconnectMqtt(DeviceUtils.getDeviceId(context))
                    }
                )
            }
        },
        bottomContent = {
            MeetingControlPill(
                isRecording = isRecording,
                hasPermission = hasPermission,
                uiState = uiState,
                pulseScale = pulseScale,
                isEnabled = mqttStatus ==
                    com.example.whisperandroid.util.MqttHelper.MqttConnectionStatus.CONNECTED,
                onMicClick = {
                    val canRecord = uiState is MeetingProcessState.Idle ||
                        uiState is MeetingProcessState.Success ||
                        uiState is MeetingProcessState.Error
                    if (!isRecording && canRecord) {
                        if (!hasPermission) {
                            permissionLauncher.launch(Manifest.permission.RECORD_AUDIO)
                        } else {
                            val file =
                                com.example.whisperandroid.data.local.MeetingAudioFileStore
                                    .createMicAudioFile(context)
                            audioRecorder.start(file)
                            audioFile = file
                            isRecording = true
                            viewModel.resetState()
                        }
                    }
                },
                onUploadClick = {
                    val canUpload = uiState is MeetingProcessState.Idle ||
                        uiState is MeetingProcessState.Success ||
                        uiState is MeetingProcessState.Error
                    if (canUpload) {
                        showFilePickerSheet = true
                    }
                },
                onStopClick = {
                    audioRecorder.stop()
                    audioFile?.let { file -> audioRecorder.finalizeWav(file) }
                    isRecording = false
                    audioFile?.let { file ->
                        if (token.isNotEmpty()) {
                            viewModel.processRecording(
                                context,
                                file,
                                token,
                                summaryLanguage,
                                DeviceUtils.getDeviceId(context)
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
    ) {
        Column(modifier = Modifier.fillMaxSize()) {
            MeetingHeaderControls(
                uiState = uiState,
                emailState = emailState,
                onDownloadClick = { url ->
                    val isLegacyStorage =
                        android.os.Build.VERSION.SDK_INT <
                            android.os.Build.VERSION_CODES.TIRAMISU
                    if (isLegacyStorage) {
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

    if (showEmailDialog) {
        EmailInputDialog(
            onDismiss = { showEmailDialog = false },
            onSend = { isMacMode, target, subject ->
                if (isMacMode) {
                    viewModel.sendEmailSummaryByMac(target, subject, summaryLanguage)
                } else {
                    viewModel.sendEmailSummary(target, subject, summaryLanguage)
                }
            }
        )
    }

    if (showFilePickerSheet) {
        val files = remember {
            com.example.whisperandroid.data.local.MeetingAudioFileStore
                .listMeetingAudioFiles(context)
        }
        MeetingFilePickerSheet(
            files = files,
            onFileSelected = { file ->
                showFilePickerSheet = false
                if (file.exists() && file.length() > 0) {
                    audioFile = file
                    if (token.isNotEmpty()) {
                        viewModel.processRecording(
                            context.applicationContext,
                            file,
                            token,
                            summaryLanguage,
                            DeviceUtils.getDeviceId(context.applicationContext)
                        )
                    }
                } else {
                    Toast.makeText(context, "Invalid or corrupted file", Toast.LENGTH_SHORT).show()
                }
            },
            onBrowseOtherClick = {
                showFilePickerSheet = false
                launcher.launch("audio/*")
            },
            onDismiss = { showFilePickerSheet = false }
        )
    }
}
