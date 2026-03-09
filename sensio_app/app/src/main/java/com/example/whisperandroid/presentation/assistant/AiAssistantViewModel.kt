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
        Idle, Recording, Processing
    }

    var assistantState by mutableStateOf(AssistantState.Idle)
        private set

    val isRecording: Boolean
        get() = assistantState == AssistantState.Recording

    val isProcessing: Boolean
        get() = assistantState == AssistantState.Processing

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

    var lastAssistantError by mutableStateOf<String?>(null)
        private set

    val shouldWakeWordListen: Boolean
        get() = assistantState == AssistantState.Idle &&
            mqttStatus == MqttHelper.MqttConnectionStatus.CONNECTED

    private var lastUserMessageNormalized: String? = null
    private var lastUserMessageAtMs: Long = 0L

    // Latency tracking
    private val requestStartedAtMs = mutableMapOf<String, Long>()

    private val mqttHelper = com.example.whisperandroid.data.di.NetworkModule.mqttHelper
    private val audioRecorder = AudioRecorder(application)
    private var currentRecordingFile: File? = null

    private fun transitionTo(newState: AssistantState, trigger: String, requestId: String? = null) {
        val oldState = assistantState
        if (oldState == newState && newState != AssistantState.Idle) return

        android.util.Log.d(
            "AiAssistantViewModel",
            "FSM Transition: $oldState -> $newState | " +
                "Trigger: $trigger | RID: ${requestId ?: activeRequestId}"
        )

        assistantState = newState

        if (newState == AssistantState.Idle) {
            val rid = activeRequestId
            if (rid != null) {
                requestStartedAtMs.remove(rid)
            }
            activeRequestId = null
            activeRequestType = null
            activeRequestText = null
        }
    }

    private fun completeRequest(
        requestId: String?,
        message: String?,
        elapsed: Long?,
        source: String
    ) {
        val currentRequestId = activeRequestId
        if (requestId != null && (currentRequestId == null || requestId != currentRequestId)) {
            android.util.Log.w(
                "AiAssistantViewModel",
                "completeRequest ignored (stale/unowned RID). " +
                    "Source: $source, Expected: $currentRequestId, Got: $requestId, State: $assistantState"
            )
            return
        }

        if (message != null) {
            transcriptionResults = transcriptionResults + TranscriptionMessage(
                text = message,
                role = MessageRole.ASSISTANT,
                requestId = requestId,
                finishedInMs = elapsed,
                source = source
            )
        }
        transitionTo(AssistantState.Idle, "completeRequest ($source)", requestId)
    }

    private fun failRequest(requestId: String?, error: String?) {
        val currentRequestId = activeRequestId
        if (requestId != null && (currentRequestId == null || requestId != currentRequestId)) {
            android.util.Log.w(
                "AiAssistantViewModel",
                "failRequest ignored (stale/unowned RID). " +
                    "Expected: $currentRequestId, Got: $requestId, State: $assistantState"
            )
            return
        }

        if (error != null) {
            lastAssistantError = error
        }
        transitionTo(AssistantState.Idle, "failRequest", requestId)
    }

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
                            val responseRequestId = try {
                                val parsed = com.google.gson.JsonParser
                                    .parseString(message).asJsonObject
                                if (parsed.has("request_id") &&
                                    !parsed.get("request_id").isJsonNull
                                ) {
                                    parsed.get("request_id").asString
                                } else if (parsed.has("data") && !parsed.get("data").isJsonNull) {
                                    val data = parsed.getAsJsonObject("data")
                                    if (data.has("request_id") &&
                                        !data.get("request_id").isJsonNull
                                    ) {
                                        data.get("request_id").asString
                                    } else {
                                        null
                                    }
                                } else {
                                    null
                                }
                            } catch (e: Exception) {
                                null
                            }

                            val currentRequestId = activeRequestId

                            // Strict early guards
                            if (currentRequestId == null) {
                                android.util.Log.w(
                                    "AiAssistantViewModel",
                                    "No active request, dropped chat/answer. Message: $message"
                                )
                                return@collect
                            }

                            if (responseRequestId != currentRequestId) {
                                android.util.Log.w(
                                    "AiAssistantViewModel",
                                    "MQTT response dropped (wrong or missing RID). " +
                                        "Expected: $currentRequestId, Got: $responseRequestId"
                                )
                                return@collect
                            }

                            val json = try {
                                com.google.gson.JsonParser.parseString(message).asJsonObject
                            } catch (e: Exception) {
                                null
                            }

                            val source = if (json != null) {
                                val data = if (json.has("data") && !json.get("data").isJsonNull) {
                                    json.getAsJsonObject("data")
                                } else {
                                    null
                                }
                                if (data != null && data.has("source") && !data.get("source").isJsonNull) {
                                    data.get("source").asString
                                } else {
                                    null
                                }
                            } else {
                                null
                            }

                            // Ignore ack-only sync/drop messages to prevent premature transition away from Processing
                            if (source == "MQTT_SYNC_DROP" || source == "MQTT_DUP_DROP") {
                                android.util.Log.d(
                                    "AiAssistantViewModel",
                                    "Ignored sync/dup drop from $source for RID: $currentRequestId"
                                )
                                return@collect
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

                            val elapsed = currentRequestId?.let { rid ->
                                requestStartedAtMs[rid]?.let { start ->
                                    System.currentTimeMillis() - start
                                }
                            }

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

                            completeRequest(currentRequestId, cleanMessage, elapsed, "mqtt")
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
                    failRequest(activeRequestId, "Error parsing response")
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

    fun clearLastError() {
        lastAssistantError = null
    }

    fun selectLanguage(language: String) {
        selectedLanguage = language
    }

    private fun isBusy(): Boolean {
        return assistantState == AssistantState.Recording ||
            assistantState == AssistantState.Processing
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
            transitionTo(AssistantState.Processing, "sendChat")

            viewModelScope.launch {
                val requestId = activeRequestId ?: return@launch
                requestStartedAtMs[requestId] = System.currentTimeMillis()
                val result = mqttHelper.publishChat(text, selectedLanguage, requestId)
                if (result.isSuccess) {
                    startResponseTimeout(requestId)
                } else {
                    failRequest(requestId, "Gagal mengirim pesan (MQTT error)")
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

        lastAssistantError = null
        val requestId = java.util.UUID.randomUUID().toString()
        activeRequestId = requestId
        activeRequestType = RequestType.Audio
        activeRequestText = null
        transitionTo(AssistantState.Recording, "startRecording")
        requestStartedAtMs[requestId] = System.currentTimeMillis()
        currentRecordingFile = file

        val result = audioRecorder.start(file)
        if (result !is AudioRecorder.RecorderResult.Success) {
            val error = when (result) {
                is AudioRecorder.RecorderResult.MicBusy ->
                    "Microphone is busy. Please try again in a moment."
                is AudioRecorder.RecorderResult.FileError ->
                    "Storage error: ${result.details}"
                else -> "Failed to initialize microphone."
            }
            failRequest(requestId, error)
            android.util.Log.e("AiAssistantViewModel", "startRecording failed: $lastAssistantError")
        }
    }

    fun stopRecording() {
        if (assistantState == AssistantState.Recording) {
            audioRecorder.stop()

            val file = currentRecordingFile
            if (file != null && audioRecorder.isWavValid(file)) {
                audioRecorder.finalizeWav(file)
                transitionTo(AssistantState.Processing, "stopRecording(valid)")
                val requestId = activeRequestId
                viewModelScope.launch {
                    val bytes = file.readBytes()
                    val result = mqttHelper.publishAudio(bytes, selectedLanguage, requestId)
                    if (result.isSuccess) {
                        startResponseTimeout(requestId)
                    } else {
                        failRequest(requestId, "Gagal mengirim audio (MQTT error)")
                        android.util.Log.e("AiAssistantViewModel", "publishAudio failed")
                    }
                }
            } else {
                android.util.Log.e("AiAssistantViewModel", "Recording file is invalid or too small")
                failRequest(activeRequestId, "Rekaman terlalu pendek atau tidak valid")
            }
        }
    }

    fun abortProcessing() {
        transitionTo(AssistantState.Idle, "abortProcessing")
    }

    private fun startResponseTimeout(requestId: String?) {
        if (requestId == null) return
        viewModelScope.launch {
            kotlinx.coroutines.delay(12000L) // 12s Standard Fallback Delay
            if (activeRequestId == requestId && assistantState == AssistantState.Processing) {
                android.util.Log.e(
                    "AiAssistantViewModel",
                    "Response timeout reached for active request, transitioning to fallback"
                )
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
            val tuyaUid = tm.getTuyaUid()

            if (token == null) {
                failRequest(requestId, "Authentication error: Token missing")
                return@launch
            }

            if (type == RequestType.Chat) {
                val text = activeRequestText
                if (text != null) {
                    com.example.whisperandroid.data.di.NetworkModule.ragRepository.chat(
                        prompt = text,
                        language = selectedLanguage,
                        terminalId = terminalId,
                        uid = tuyaUid,
                        token = token,
                        requestId = requestId,
                        idempotencyKey = fallbackIdempotencyKey
                    ).collect { resource ->
                        if (activeRequestId != requestId) return@collect
                        when (resource) {
                            is com.example.whisperandroid.domain.repository.Resource.Success -> {
                                val data = resource.data
                                if (data != null) {
                                    val responseText = data.response
                                    if (responseText != null) {
                                        handleHttpChatResponse(responseText, false, requestId)
                                    } else if (data.source == "HTTP_DUP_DROP") {
                                        android.util.Log.d("AiAssistantViewModel", "Fallback: Silent completion for HTTP_DUP_DROP")
                                        completeRequest(requestId, null, null, "http_dup")
                                    } else {
                                        failRequest(requestId, "Invalid fallback response (null body)")
                                    }
                                } else {
                                    failRequest(requestId, "Invalid fallback response")
                                }
                            }
                            is com.example.whisperandroid.domain.repository.Resource.Error -> {
                                failRequest(requestId, "Fallback API error")
                            }
                            is com.example.whisperandroid.domain.repository.Resource.Loading -> {}
                        }
                    }
                } else {
                    failRequest(requestId, "Cannot run fallback: text is missing")
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
                                    failRequest(requestId, "Fallback audio trigger failed")
                                }
                            }
                            is com.example.whisperandroid.domain.repository.Resource.Error -> {
                                failRequest(requestId, "Fallback audio API error")
                            }
                            is com.example.whisperandroid.domain.repository.Resource.Loading -> {}
                        }
                    }
                } else {
                    failRequest(requestId, "Audio file missing for fallback")
                }
            } else {
                failRequest(requestId, "Unknown request type for fallback")
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
        val username = com.example.whisperandroid.util.DeviceUtils.getDeviceId(getApplication())
        val tuyaUid = com.example.whisperandroid.data.di.NetworkModule.tokenManager.getTuyaUid()
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
                        uid = tuyaUid,
                        token = token,
                        requestId = requestId,
                        idempotencyKey = fallbackIdempotencyKey + "_chat"
                    ).collect { chatResult ->
                        if (activeRequestId != requestId) return@collect
                        when (chatResult) {
                            is com.example.whisperandroid.domain.repository.Resource.Success -> {
                                val data = chatResult.data
                                if (data != null) {
                                    val responseText = data.response
                                    if (responseText != null) {
                                        handleHttpChatResponse(responseText, false, requestId)
                                    } else if (data.source == "HTTP_DUP_DROP") {
                                        android.util.Log.d("AiAssistantViewModel", "Poll: Silent completion for HTTP_DUP_DROP")
                                        completeRequest(requestId, null, null, "http_dup")
                                    } else {
                                        failRequest(requestId, "Invalid poll response (null body)")
                                    }
                                } else {
                                    failRequest(requestId, "Invalid poll response")
                                }
                            }
                            is com.example.whisperandroid.domain.repository.Resource.Error -> {
                                failRequest(requestId, "Poll API error")
                            }
                            is com.example.whisperandroid.domain.repository.Resource.Loading -> {}
                        }
                    }
                } else if (hasError) {
                    failRequest(requestId, "Transcription poll failed")
                    return@launch
                } else {
                    attempts++
                    kotlinx.coroutines.delay(1500L) // 1.5s delay between polls
                }
            }
            if (!isCompleted) {
                failRequest(requestId, "Polling timeout")
            }
        }
    }

    private fun handleHttpChatResponse(
        responseText: String,
        isBlocked: Boolean,
        requestId: String?
    ) {
        val currentRequestId = activeRequestId
        if (requestId != null && currentRequestId != null && requestId != currentRequestId) {
            android.util.Log.w(
                "AiAssistantViewModel",
                "Stale HTTP response dropped. Expected: $currentRequestId, Got: $requestId"
            )
            return
        }

        val elapsed = currentRequestId?.let { rid ->
            requestStartedAtMs[rid]?.let { start ->
                System.currentTimeMillis() - start
            }
        }

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

        completeRequest(currentRequestId, cleanMessage, elapsed, "http_fallback")
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
