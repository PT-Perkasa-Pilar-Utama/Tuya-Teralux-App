package com.example.whisperandroid.presentation.meeting

import android.content.Context
import android.content.Intent
import androidx.compose.runtime.getValue
import androidx.compose.runtime.setValue
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.whisperandroid.data.di.NetworkModule
import com.example.whisperandroid.data.manager.MeetingProcessManager
import com.example.whisperandroid.domain.repository.Resource
import com.example.whisperandroid.domain.usecase.MeetingProcessState
import com.example.whisperandroid.domain.usecase.ProcessMeetingUseCase
import com.example.whisperandroid.presentation.components.UiState
import com.example.whisperandroid.service.MeetingForegroundService
import java.io.File
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.collect
import kotlinx.coroutines.flow.collectLatest
import kotlinx.coroutines.launch

class MeetingViewModel(
    private val processMeetingUseCase: ProcessMeetingUseCase
) : ViewModel() {
    private val _uiState = MutableStateFlow<MeetingProcessState>(MeetingProcessState.Idle)
    val uiState: StateFlow<MeetingProcessState> = _uiState.asStateFlow()

    private val _emailState = MutableStateFlow<UiState<Boolean>>(UiState.Idle)
    val emailState: StateFlow<UiState<Boolean>> = _emailState.asStateFlow()

    init {
        // Synchronize UI state with MeetingProcessManager
        viewModelScope.launch {
            MeetingProcessManager.processState.collect { state ->
                _uiState.value = state
            }
        }
    }

    fun processRecording(
        context: Context,
        audioFile: File,
        token: String,
        targetLang: String = "Indonesian",
        macAddress: String? = null
    ) {
        val intent = Intent(context, MeetingForegroundService::class.java).apply {
            putExtra("AUDIO_PATH", audioFile.absolutePath)
            putExtra("TOKEN", token)
            putExtra("TARGET_LANG", targetLang)
            putExtra("MAC_ADDRESS", macAddress)
        }

        if (android.os.Build.VERSION.SDK_INT >= android.os.Build.VERSION_CODES.O) {
            context.startForegroundService(intent)
        } else {
            context.startService(intent)
        }
    }

    fun sendEmailSummary(
        email: String,
        subject: String,
        targetLang: String = "id"
    ) {
        val state = _uiState.value
        if (state !is MeetingProcessState.Success) return

        val token = NetworkModule.tokenManager.getAccessToken() ?: ""

        if (token.isEmpty()) {
            _emailState.value = UiState.Error("Authentication token not found. Please login again.")
            return
        }

        val recipients = email.split(",").map { it.trim() }.filter { it.isNotEmpty() }
        if (recipients.isEmpty()) {
            _emailState.value = UiState.Error("At least one valid email address is required.")
            return
        }

        viewModelScope.launch {
            NetworkModule
                .sendEmailUseCase(
                    to = recipients,
                    subject = subject,
                    template = "summary",
                    token = token,
                    attachmentPath = state.pdfUrl
                ).collectLatest { resource ->
                    when (resource) {
                        is Resource.Loading -> {
                            _emailState.value = UiState.Loading
                        }
                        is Resource.Success -> {
                            _emailState.value = UiState.Success(true)
                        }
                        is Resource.Error -> {
                            _emailState.value = UiState.Error(
                                resource.message ?: "Failed to send email"
                            )
                        }
                    }
                }
        }
    }

    fun sendEmailSummaryByMac(
        macAddress: String,
        subject: String,
        targetLang: String = "id"
    ) {
        val state = _uiState.value
        if (state !is MeetingProcessState.Success) return

        val token = NetworkModule.tokenManager.getAccessToken() ?: ""

        if (token.isEmpty()) {
            _emailState.value = UiState.Error("Authentication token not found. Please login again.")
            return
        }

        val overrideEmails = if (macAddress.contains("@")) {
            macAddress.split(",").map { it.trim() }.filter { it.isNotEmpty() }
        } else {
            null
        }

        viewModelScope.launch {
            NetworkModule
                .sendEmailByMacUseCase(
                    macAddress = if (overrideEmails != null) "" else macAddress,
                    subject = subject,
                    template = "summary",
                    token = token,
                    attachmentPath = state.pdfUrl,
                    overrideEmails = overrideEmails
                ).collectLatest { resource ->
                    when (resource) {
                        is Resource.Loading -> {
                            _emailState.value = UiState.Loading
                        }
                        is Resource.Success -> {
                            _emailState.value = UiState.Success(true)
                        }
                        is Resource.Error -> {
                            _emailState.value = UiState.Error(
                                resource.message ?: "Failed to send email by MAC"
                            )
                        }
                    }
                }
        }
    }

    fun resetEmailState() {
        _emailState.value = UiState.Idle
    }

    fun cancelProcessing(context: Context) {
        // Reset email state for clean UI
        _emailState.value = UiState.Idle

        // Only send cancel intent if there's an active processing job
        // This avoids Android O+ foreground service contract violations
        if (MeetingForegroundService.isProcessingActive()) {
            val cancelIntent = Intent(context, MeetingForegroundService::class.java).apply {
                action = MeetingForegroundService.ACTION_CANCEL
            }

            if (android.os.Build.VERSION.SDK_INT >= android.os.Build.VERSION_CODES.O) {
                context.startForegroundService(cancelIntent)
            } else {
                context.startService(cancelIntent)
            }
            // Don't set UI state here - let the service publish Cancelled when job actually stops
        } else {
            // No active processing job, clear local state and any persisted session state
            // Get the audio path from MeetingProcessManager to clear session state
            val audioPath = MeetingProcessManager.getCurrentAudioPath()
            if (audioPath != null) {
                processMeetingUseCase.clearSessionState(audioPath)
            }
            // MeetingProcessManager.cancel() will publish Cancelled state
            MeetingProcessManager.cancel()
        }
    }

    fun resetState() {
        _uiState.value = MeetingProcessState.Idle
        _emailState.value = UiState.Idle
        MeetingProcessManager.reset()
    }
}
