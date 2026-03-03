package com.example.whisper_android.presentation.meeting

import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
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
import kotlinx.coroutines.flow.collect
import kotlinx.coroutines.flow.collectLatest
import kotlinx.coroutines.launch
import kotlinx.coroutines.channels.Channel
import com.google.gson.JsonParser

class MeetingViewModel(
    private val processMeetingUseCase: ProcessMeetingUseCase
) : ViewModel() {
    private val _uiState = MutableStateFlow<MeetingProcessState>(MeetingProcessState.Idle)
    val uiState: StateFlow<MeetingProcessState> = _uiState.asStateFlow()

    private val _emailState = MutableStateFlow<UiState<Boolean>>(UiState.Idle)
    val emailState: StateFlow<UiState<Boolean>> = _emailState.asStateFlow()

    private val _mqttStatus = MutableStateFlow(com.example.whisper_android.util.MqttHelper.MqttConnectionStatus.DISCONNECTED)
    val mqttStatus: StateFlow<com.example.whisper_android.util.MqttHelper.MqttConnectionStatus> = _mqttStatus

    private val mqttHelper = com.example.whisper_android.data.di.NetworkModule.mqttHelper

    init {
        mqttHelper.onConnectionStatusChanged = { status ->
            viewModelScope.launch {
                _mqttStatus.value = status
            }
        }
    }

    fun reconnectMqtt(deviceId: String) {
        viewModelScope.launch {
            val pwdResult = NetworkModule.repository.fetchMqttPassword(deviceId)
            if (pwdResult.isSuccess) {
                mqttHelper.connect(pwdResult.getOrNull()!!)
            }
        }
    }

    fun processRecording(
        audioFile: File,
        token: String,
        targetLang: String = "Indonesian",
        macAddress: String? = null
    ) {
        viewModelScope.launch {
            val signalChannel = Channel<String>(1)
            
            val messageJob = launch {
                mqttHelper.messages.collect { (topic, msg) ->
                    val taskTopic = mqttHelper.getTaskTopic()
                    if (taskTopic != null && topic == taskTopic) {
                        try {
                            val json = JsonParser.parseString(msg).asJsonObject
                            val event = if (json.has("event") && !json.get("event").isJsonNull) json.get("event").getAsString() else null
                            val taskLabel = if (json.has("task") && !json.get("task").isJsonNull) json.get("task").getAsString() else null
                            if (event == "stop" && taskLabel != null) {
                                signalChannel.trySend(taskLabel)
                            }
                        } catch (e: Exception) {
                            android.util.Log.e("MeetingViewModel", "Error parsing task JSON", e)
                        }
                    }
                }
            }

            try {
                processMeetingUseCase(
                    audioFile = audioFile,
                    token = token,
                    targetLang = targetLang,
                    macAddress = macAddress,
                    waitSignal = { taskName ->
                        mqttStatus = com.example.whisper_android.util.MqttHelper.MqttConnectionStatus.CONNECTED // Assume connected
                        
                        // Publish Start signal
                        mqttHelper.publishTaskMessage("start", taskName)
                        
                        // Wait for stop signal
                        while (true) {
                            val receivedTaskName = signalChannel.receive()
                            if (receivedTaskName == taskName) break
                        }
                    }
                ).collect { state ->
                    _uiState.value = state
                }
            } finally {
                messageJob.cancel()
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

    fun resetState() {
        _uiState.value = MeetingProcessState.Idle
        _emailState.value = UiState.Idle
    }
}
