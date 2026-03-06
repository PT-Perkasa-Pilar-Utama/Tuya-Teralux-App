package com.example.whisperandroid.presentation.assistant

import android.app.Application
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.example.whisperandroid.presentation.components.MessageRole
import com.example.whisperandroid.presentation.components.TranscriptionMessage
import com.example.whisperandroid.presentation.meeting.AudioRecorder
import com.example.whisperandroid.util.MqttHelper
import com.example.whisperandroid.util.parseMarkdownToText
import java.io.File
import kotlinx.coroutines.flow.collect
import kotlinx.coroutines.launch

class AiAssistantViewModel(
    application: Application
) : AndroidViewModel(application) {
    var transcriptionResults by mutableStateOf(listOf<TranscriptionMessage>())
        private set

    enum class AssistantState {
        Idle, Recording, Publishing, WaitingResponse, FallbackRunning, Completed, Failed
    }

    var assistantState by mutableStateOf(AssistantState.Idle)
        private set

    val isRecording: Boolean
        get() = assistantState == AssistantState.Recording

    val isProcessing: Boolean
        get() = assistantState == AssistantState.Publishing ||
            assistantState == AssistantState.WaitingResponse ||
            assistantState == AssistantState.FallbackRunning

    var activeRequestId by mutableStateOf<String?>(null)
        private set

    enum class RequestType { Audio, Chat }

    var activeRequestType by mutableStateOf<RequestType?>(null)
        private set

    var activeRequestText by mutableStateOf<String?>(null)
        private set

    var selectedLanguage by mutableStateOf("id")
        private set

    var mqttStatus by mutableStateOf(MqttHelper.MqttConnectionStatus.DISCONNECTED)
        private set

    private var lastUserMessageNormalized: String? = null
    private var lastUserMessageAtMs: Long = 0L

    private val mqttHelper = com.example.whisperandroid.data.di.NetworkModule.mqttHelper
    private val audioRecorder = AudioRecorder(application)
    private var currentRecordingFile: File? = null

    init {
        viewModelScope.launch {
            mqttHelper.messages.collect { (rawTopic, rawMessage) ->
                val topic = rawTopic as String
                val message = rawMessage as String
                android.util.Log.d(
                    "AiAssistantViewModel",
                    "MQTT Message: topic=$topic, message=$message"
                )
                try {
                    when {
                        topic.endsWith("chat/answer") -> {
                            android.util.Log.d(
                                "AiAssistantViewModel",
                                "Received chat/answer: $message"
                            )
                            val json = try {
                                com.google.gson.JsonParser.parseString(message).asJsonObject
                            } catch (e: Exception) {
                                null
                            }

                            val responseText = if (json != null) {
                                val data = if (json.has("data") && !json.get("data").isJsonNull) {
                                    json.getAsJsonObject("data")
                                } else {
                                    null
                                }
                                if (data != null &&
                                    data.has("response") &&
                                    !data.get("response").isJsonNull
                                ) {
                                    data.get("response").asString
                                } else if (json.has("message") && !json.get("message").isJsonNull) {
                                    json.get("message").asString
                                } else {
                                    null
                                }
                            } else {
                                // Fallback for log-style strings as shown in user example
                                if (message.contains("Response: \"")) {
                                    message.substringAfter("Response: \"").substringBeforeLast("\"")
                                } else {
                                    null
                                }
                            }

                            // Check if the response was blocked by the guard
                            val isBlocked = if (json != null) {
                                val data = if (json.has("data") && !json.get("data").isJsonNull) {
                                    json.getAsJsonObject("data")
                                } else {
                                    null
                                }
                                val hasBlocked = data != null &&
                                    data.has("is_blocked") &&
                                    !data.get("is_blocked").isJsonNull
                                hasBlocked && data!!.get("is_blocked").asBoolean
                            } else {
                                false
                            }

                            val wasRecording = assistantState == AssistantState.Recording

                            // ALWAYS stop active states when any answer arrives
                            assistantState = AssistantState.Completed
                            activeRequestId = null
                            android.util.Log.d(
                                "AiAssistantViewModel",
                                "assistantState set to Completed (answer received)"
                            )

                            if (isBlocked) {
                                // Remove BOTH the preemptively-added USER bubble
                                // and any sync USER bubble from previous prompt
                                val userMessages = transcriptionResults.filter {
                                    it.role == MessageRole.USER
                                }
                                if (userMessages.isNotEmpty()) {
                                    transcriptionResults = transcriptionResults
                                        .toMutableList()
                                        .apply {
                                            val lastUserIndex = indexOfLast {
                                                it.role == MessageRole.USER
                                            }
                                            if (lastUserIndex >= 0) removeAt(lastUserIndex)
                                        }
                                }

                                if (wasRecording) {
                                    audioRecorder.stop()
                                }

                                android.util.Log.d(
                                    "AiAssistantViewModel",
                                    "Guard blocked prompt: $message, showing identity fallback"
                                )
                            }

                            val isValidationError = json != null &&
                                json.has("message") &&
                                !json.get("message").isJsonNull &&
                                json.get("message").asString == "Validation Error"

                            val cleanMessage = when {
                                isValidationError ->
                                    "Maaf, suara tidak terdengar dengan jelas. Silakan coba lagi."
                                responseText != null && responseText.isNotBlank() ->
                                    parseMarkdownToText(responseText).trim().removeSurrounding("\"")
                                isBlocked ->
                                    "Halo! Saya Sensio, asisten rumah pintar Anda. " +
                                        "Ada yang bisa saya bantu?"
                                else -> null
                            }

                            if (cleanMessage != null) {
                                transcriptionResults = transcriptionResults + TranscriptionMessage(
                                    text = cleanMessage,
                                    role = MessageRole.ASSISTANT
                                )
                            }
                        }

                        topic.endsWith("chat") -> {
                            android.util.Log.d(
                                "AiAssistantViewModel",
                                "Received chat (sync): $message"
                            )
                            // Extract prompt if it's JSON, otherwise use raw message
                            val prompt =
                                try {
                                    val parsed = com.google.gson.JsonParser.parseString(message)
                                    val jsonObj = parsed.asJsonObject
                                    if (jsonObj.has("prompt") &&
                                        !jsonObj.get("prompt").isJsonNull
                                    ) {
                                        jsonObj.get("prompt").asString
                                    } else {
                                        message
                                    }
                                } catch (e: Exception) {
                                    message
                                }

                            val cleanPrompt = prompt.trim().removeSurrounding("\"")

                            if (cleanPrompt.isNotBlank()) {
                                appendUserMessageIfNeeded(cleanPrompt)
                            }
                        }

                        topic.endsWith("whisper/answer") -> {
                            // Handle whisper answer (e.g. task ID)
                            android.util.Log.d("AiAssistantViewModel", "Whisper Task: $message")
                        }
                    }
                } catch (e: Exception) {
                    android.util.Log.e("AiAssistantViewModel", "Error parsing MQTT message", e)
                    // Robust reset on error
                    assistantState = AssistantState.Failed
                    activeRequestId = null
                }
            }
        }

        viewModelScope.launch {
            mqttHelper.connectionStatus.collect { status ->
                android.util.Log.d("AiAssistantViewModel", "MQTT Status Changed: $status")
                mqttStatus = status
            }
        }

        // Auto-connect when the ViewModel is initialized
        reconnectMqtt()
    }

    fun reconnectMqtt() {
        viewModelScope.launch {
            if (mqttHelper.connectionStatus.value == MqttHelper.MqttConnectionStatus.CONNECTED ||
                mqttHelper.connectionStatus.value == MqttHelper.MqttConnectionStatus.CONNECTING
            ) {
                return@launch
            }
            android.util.Log.d("AiAssistantViewModel", "Manual MQTT Reconnection...")
            val username = com.example.whisperandroid.util.DeviceUtils.getDeviceId(
                getApplication()
            )
            val repo = com.example.whisperandroid.data.di.NetworkModule.repository
            val pwdResult = repo.fetchMqttPassword(
                username
            )
            if (pwdResult.isSuccess) {
                mqttHelper.connect(pwdResult.getOrNull()!!)
            } else {
                android.util.Log.e(
                    "AiAssistantViewModel",
                    "Failed to fetch MQTT password: ${pwdResult.exceptionOrNull()?.message}"
                )
                mqttStatus = MqttHelper.MqttConnectionStatus.NO_INTERNET
            }
        }
    }

    fun selectLanguage(language: String) {
        selectedLanguage = language
    }

    private fun isBusy(): Boolean {
        return assistantState == AssistantState.Recording ||
            assistantState == AssistantState.Publishing ||
            assistantState == AssistantState.WaitingResponse ||
            assistantState == AssistantState.FallbackRunning
    }

    fun sendChat(text: String) {
        if (isBusy()) {
            android.util.Log.w(
                "AiAssistantViewModel",
                "sendChat: Request already in progress, ignoring"
            )
            return
        }

        if (text.isNotBlank()) {
            android.util.Log.d("AiAssistantViewModel", "sendChat: $text")

            appendUserMessageIfNeeded(text)

            activeRequestId = java.util.UUID.randomUUID().toString()
            activeRequestType = RequestType.Chat
            activeRequestText = text
            assistantState = AssistantState.Publishing

            viewModelScope.launch {
                val result = mqttHelper.publishChat(text, selectedLanguage)
                if (result.isSuccess) {
                    assistantState = AssistantState.WaitingResponse
                    startResponseTimeout(activeRequestId)
                } else {
                    assistantState = AssistantState.Failed
                    activeRequestId = null
                    android.util.Log.e("AiAssistantViewModel", "publishChat failed")
                }
            }
        }
    }

    fun startRecording(file: File) {
        if (isBusy()) {
            android.util.Log.w(
                "AiAssistantViewModel",
                "startRecording: Request already in progress, ignoring"
            )
            return
        }

        activeRequestId = java.util.UUID.randomUUID().toString()
        activeRequestType = RequestType.Audio
        activeRequestText = null
        assistantState = AssistantState.Recording
        currentRecordingFile = file
        val started = audioRecorder.start(file)
        if (!started) {
            assistantState = AssistantState.Failed
            activeRequestId = null
            activeRequestType = null
            android.util.Log.e("AiAssistantViewModel", "startRecording failed to initialize")
        }
    }

    fun stopRecording() {
        if (assistantState == AssistantState.Recording) {
            audioRecorder.stop()

            val file = currentRecordingFile
            if (file != null && audioRecorder.isWavValid(file)) {
                audioRecorder.finalizeWav(file)
                assistantState = AssistantState.Publishing
                viewModelScope.launch {
                    val bytes = file.readBytes()
                    val result = mqttHelper.publishAudio(bytes, selectedLanguage)
                    if (result.isSuccess) {
                        assistantState = AssistantState.WaitingResponse
                        startResponseTimeout(activeRequestId)
                    } else {
                        assistantState = AssistantState.Failed
                        activeRequestId = null
                        android.util.Log.e("AiAssistantViewModel", "publishAudio failed")
                    }
                }
            } else {
                android.util.Log.e("AiAssistantViewModel", "Recording file is invalid or too small")
                assistantState = AssistantState.Failed
                activeRequestId = null
            }
        }
    }

    fun abortProcessing() {
        assistantState = AssistantState.Idle
        activeRequestId = null
        android.util.Log.d("AiAssistantViewModel", "abortProcessing: reset to Idle")
    }

    private fun startResponseTimeout(requestId: String?) {
        if (requestId == null) return
        viewModelScope.launch {
            kotlinx.coroutines.delay(25000L) // 25s
            if (activeRequestId == requestId && assistantState == AssistantState.WaitingResponse) {
                android.util.Log.e(
                    "AiAssistantViewModel",
                    "Response timeout reached for active request, transitioning to fallback"
                )
                assistantState = AssistantState.FallbackRunning
                runHttpFallback()
            }
        }
    }

    private fun runHttpFallback() {
        val requestId = activeRequestId ?: return
        val type = activeRequestType
        val fallbackIdempotencyKey = "fallback_$requestId"

        android.util.Log.d(
            "AiAssistantViewModel",
            "Starting HTTP Fallback for request: $requestId, type: $type"
        )

        viewModelScope.launch {
            val username = com.example.whisperandroid.util.DeviceUtils.getDeviceId(getApplication())
            val tm = com.example.whisperandroid.data.di.NetworkModule.tokenManager
            val token = tm.getAccessToken()
            val terminalId = tm.getTerminalId() ?: username

            if (token == null) {
                assistantState = AssistantState.Failed
                activeRequestId = null
                return@launch
            }

            if (type == RequestType.Chat) {
                val text = activeRequestText
                if (text != null) {
                    com.example.whisperandroid.data.di.NetworkModule.ragRepository.chat(
                        prompt = text,
                        language = selectedLanguage,
                        terminalId = terminalId,
                        uid = null,
                        token = token,
                        idempotencyKey = fallbackIdempotencyKey
                    ).collect { result ->
                        if (activeRequestId != requestId) return@collect
                        when (result) {
                            is com.example.whisperandroid.domain.repository.Resource.Success -> {
                                val data = result.data
                                if (data != null) {
                                    handleHttpChatResponse(data.response, false)
                                } else {
                                    assistantState = AssistantState.Failed
                                    activeRequestId = null
                                }
                            }
                            is com.example.whisperandroid.domain.repository.Resource.Error -> {
                                assistantState = AssistantState.Failed
                                activeRequestId = null
                            }
                            is com.example.whisperandroid.domain.repository.Resource.Loading -> {}
                        }
                    }
                } else {
                    assistantState = AssistantState.Failed
                    activeRequestId = null
                }
            } else if (type == RequestType.Audio) {
                val file = currentRecordingFile
                if (file != null && file.exists()) {
                    val transcribeAudioUseCase =
                        com.example.whisperandroid.data.di.NetworkModule.transcribeAudioUseCase
                    transcribeAudioUseCase.initiate(
                        audioFile = file,
                        token = token,
                        language = selectedLanguage,
                        macAddress = terminalId,
                        idempotencyKey = fallbackIdempotencyKey
                    ).collect { result ->
                        if (activeRequestId != requestId) return@collect
                        when (result) {
                            is com.example.whisperandroid.domain.repository.Resource.Success -> {
                                val taskId = result.data
                                if (taskId != null) {
                                    pollHttpTranscription(
                                        taskId,
                                        requestId,
                                        token,
                                        terminalId,
                                        fallbackIdempotencyKey
                                    )
                                } else {
                                    assistantState = AssistantState.Failed
                                    activeRequestId = null
                                }
                            }
                            is com.example.whisperandroid.domain.repository.Resource.Error -> {
                                assistantState = AssistantState.Failed
                                activeRequestId = null
                            }
                            is com.example.whisperandroid.domain.repository.Resource.Loading -> {}
                        }
                    }
                } else {
                    assistantState = AssistantState.Failed
                    activeRequestId = null
                }
            } else {
                assistantState = AssistantState.Failed
                activeRequestId = null
            }
        }
    }

    private fun pollHttpTranscription(
        taskId: String,
        requestId: String,
        token: String,
        terminalId: String,
        fallbackIdempotencyKey: String
    ) {
        viewModelScope.launch {
            var isCompleted = false
            var attempts = 0
            while (!isCompleted && attempts < 10) {
                if (activeRequestId != requestId) return@launch
                var currentSuccessText: String? = null
                var hasError = false

                com.example.whisperandroid.data.di.NetworkModule.transcribeAudioUseCase.getResult(
                    taskId = taskId,
                    token = token
                ).collect { result ->
                    when (result) {
                        is com.example.whisperandroid.domain.repository.Resource.Success -> {
                            currentSuccessText = result.data
                            isCompleted = true
                        }
                        is com.example.whisperandroid.domain.repository.Resource.Error -> {
                            hasError = true
                        }
                        is com.example.whisperandroid.domain.repository.Resource.Loading -> {}
                    }
                }

                if (isCompleted && currentSuccessText != null) {
                    val transcribedText = currentSuccessText!!

                    appendUserMessageIfNeeded(transcribedText)

                    com.example.whisperandroid.data.di.NetworkModule.ragRepository.chat(
                        prompt = transcribedText,
                        language = selectedLanguage,
                        terminalId = terminalId,
                        uid = null,
                        token = token,
                        idempotencyKey = fallbackIdempotencyKey + "_chat"
                    ).collect { chatResult ->
                        if (activeRequestId != requestId) return@collect
                        when (chatResult) {
                            is com.example.whisperandroid.domain.repository.Resource.Success -> {
                                val data = chatResult.data
                                if (data != null) {
                                    handleHttpChatResponse(data.response, false)
                                } else {
                                    assistantState = AssistantState.Failed
                                    activeRequestId = null
                                }
                            }
                            is com.example.whisperandroid.domain.repository.Resource.Error -> {
                                assistantState = AssistantState.Failed
                                activeRequestId = null
                            }
                            is com.example.whisperandroid.domain.repository.Resource.Loading -> {}
                        }
                    }
                } else if (hasError) {
                    assistantState = AssistantState.Failed
                    activeRequestId = null
                    return@launch
                } else {
                    attempts++
                    kotlinx.coroutines.delay(2000L)
                }
            }
            if (!isCompleted) {
                assistantState = AssistantState.Failed
                activeRequestId = null
            }
        }
    }

    private fun handleHttpChatResponse(responseText: String, isBlocked: Boolean) {
        assistantState = AssistantState.Completed
        activeRequestId = null

        if (isBlocked) {
            val userMessages = transcriptionResults.filter { it.role == MessageRole.USER }
            if (userMessages.isNotEmpty()) {
                transcriptionResults = transcriptionResults.toMutableList().apply {
                    val lastUserIndex = indexOfLast { it.role == MessageRole.USER }
                    if (lastUserIndex >= 0) removeAt(lastUserIndex)
                }
            }
            android.util.Log.d("AiAssistantViewModel", "HTTP Fallback: Guard blocked prompt")
        }

        val cleanMessage = when {
            responseText.isNotBlank() -> com.example.whisperandroid.util.parseMarkdownToText(
                responseText
            ).trim().removeSurrounding("\"")
            isBlocked -> "Halo! Saya Sensio, asisten rumah pintar Anda. Ada yang bisa saya bantu?"
            else -> "Maaf, terjadi kesalahan pada koneksi fallback."
        }

        transcriptionResults = transcriptionResults + TranscriptionMessage(
            text = cleanMessage,
            role = MessageRole.ASSISTANT
        )
    }

    private fun normalizeMessageForDedup(text: String): String {
        return text.trim().lowercase().replace(Regex("\\s+"), " ")
    }

    private fun appendUserMessageIfNeeded(text: String) {
        val normalized = normalizeMessageForDedup(text)
        val now = System.currentTimeMillis()

        // Consecutive deduplication: skip if same text as last USER message AND within short window
        val isConsecutiveDup = normalized == lastUserMessageNormalized &&
            (now - lastUserMessageAtMs) <= 1200L

        // Also skip if we are currently processing a request and the last bubble is already USER (safety for multi-source sync)
        val lastRole = transcriptionResults.lastOrNull()?.role
        val isProcessingDup = isProcessing &&
            lastRole == MessageRole.USER &&
            normalized == lastUserMessageNormalized

        if (!isConsecutiveDup && !isProcessingDup) {
            transcriptionResults = transcriptionResults + TranscriptionMessage(
                text = text,
                role = MessageRole.USER
            )
            lastUserMessageNormalized = normalized
            lastUserMessageAtMs = now
        } else {
            android.util.Log.d(
                "AiAssistantViewModel",
                "Skipping duplicate USER message: $text (consecutive=$isConsecutiveDup, " +
                    "processing=$isProcessingDup)"
            )
        }
    }

    override fun onCleared() {
        super.onCleared()
        mqttHelper.disconnect()
    }
}
