package com.example.whisper_android.presentation.meeting

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.whisper_android.data.di.NetworkModule
import com.example.whisper_android.domain.usecase.MeetingProcessState
import com.example.whisper_android.domain.usecase.ProcessMeetingUseCase
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import java.io.File

class MeetingViewModel(
    private val processMeetingUseCase: ProcessMeetingUseCase
) : ViewModel() {

    private val _uiState = MutableStateFlow<MeetingProcessState>(MeetingProcessState.Idle)
    val uiState: StateFlow<MeetingProcessState> = _uiState.asStateFlow()

    // Retrieve token (e.g., from TokenManager via NetworkModule or passed in)
    // For simplicity, we'll fetch from NetworkModule directly if needed, or assume it's passed
    // But ViewModel shouldn't depend on Context.
    // Ideally, UseCase handles token retrieval.
    // Let's pass token to processMeeting
    
    fun processRecording(audioFile: File, token: String, targetLang: String = "English") {
        viewModelScope.launch {
            processMeetingUseCase(audioFile, token, targetLang).collect { state ->
                _uiState.value = state
            }
        }
    }

    fun resetState() {
        _uiState.value = MeetingProcessState.Idle
    }
}
