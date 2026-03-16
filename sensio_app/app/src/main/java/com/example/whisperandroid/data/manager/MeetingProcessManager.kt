package com.example.whisperandroid.data.manager

import com.example.whisperandroid.domain.usecase.MeetingProcessState
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow

object MeetingProcessManager {
    private val _processState = MutableStateFlow<MeetingProcessState>(MeetingProcessState.Idle)
    val processState = _processState.asStateFlow()

    // Track the current audio file path for session cleanup
    private var currentAudioPath: String? = null

    fun updateState(state: MeetingProcessState) {
        _processState.value = state
    }

    fun reset() {
        currentAudioPath = null
        _processState.value = MeetingProcessState.Idle
    }

    fun cancel() {
        currentAudioPath = null
        _processState.value = MeetingProcessState.Cancelled
    }

    fun setCurrentAudioPath(path: String) {
        currentAudioPath = path
    }

    fun getCurrentAudioPath(): String? = currentAudioPath
}
