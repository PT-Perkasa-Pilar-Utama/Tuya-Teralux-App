package com.example.whisper_android.data.manager

import com.example.whisper_android.domain.usecase.MeetingProcessState
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow

object MeetingProcessManager {
    private val _processState = MutableStateFlow<MeetingProcessState>(MeetingProcessState.Idle)
    val processState = _processState.asStateFlow()

    fun updateState(state: MeetingProcessState) {
        _processState.value = state
    }

    fun reset() {
        _processState.value = MeetingProcessState.Idle
    }
}
