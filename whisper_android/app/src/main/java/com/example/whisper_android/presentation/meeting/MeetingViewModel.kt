package com.example.whisper_android.presentation.meeting

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.whisper_android.data.di.NetworkModule
import com.example.whisper_android.domain.repository.Resource
import com.example.whisper_android.domain.usecase.MeetingProcessState
import com.example.whisper_android.domain.usecase.ProcessMeetingUseCase
import com.example.whisper_android.presentation.components.UiState
import java.io.File
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.collectLatest
import kotlinx.coroutines.launch

class MeetingViewModel(
    private val processMeetingUseCase: ProcessMeetingUseCase
) : ViewModel() {
    private val _uiState = MutableStateFlow<MeetingProcessState>(MeetingProcessState.Idle)
    val uiState: StateFlow<MeetingProcessState> = _uiState.asStateFlow()

    private val _emailState = MutableStateFlow<UiState<Boolean>>(UiState.Idle)
    val emailState: StateFlow<UiState<Boolean>> = _emailState.asStateFlow()

    fun processRecording(
        audioFile: File,
        token: String,
        targetLang: String = "Indonesian"
    ) {
        viewModelScope.launch {
            processMeetingUseCase(audioFile, token, targetLang).collect { state ->
                _uiState.value = state
            }
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

        viewModelScope.launch {
            NetworkModule
                .sendEmailUseCase(
                    to = email,
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
                            _emailState.value = UiState.Error(resource.message ?: "Failed to send email")
                        }
                    }
                }
        }
    }

    fun resetEmailState() {
        _emailState.value = UiState.Idle
    }

    fun resetState() {
        _uiState.value = MeetingProcessState.Idle
        _emailState.value = UiState.Idle
    }
}
