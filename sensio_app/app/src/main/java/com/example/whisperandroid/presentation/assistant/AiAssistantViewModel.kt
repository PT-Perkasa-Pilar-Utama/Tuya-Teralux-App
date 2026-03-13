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
    private var activeTimingLogger: AssistantTimingLogger? = null

    // Timing helper for structured latency logs
    private class AssistantTimingLogger(
        private val flow: String,
        private val requestId: String
    ) {
        private val requestStartMs = System.currentTimeMillis()
        private val checkpoints = mutableMapOf<String, Long>()
        private var lastStep: String? = null

        fun markStep(step: String, extraFields: Map<String, String> = emptyMap()) {
            val now = System.currentTimeMillis()
            val elapsedMs = now - requestStartMs
            val prevStep = lastStep
            val stepDeltaMs = if (prevStep != null && checkpoints[prevStep] != null) {
                now - checkpoints[prevStep]!!
            } else {
                0L
            }
            checkpoints[step] = now
            lastStep = step

            val fields = buildString {
                append("flow=$flow")
                append(" | request_id=$requestId")
                append(" | step=$step")
                append(" | elapsed_ms=$elapsedMs")
                if (prevStep != null && stepDeltaMs > 0) {
                    append(" | prev_step=$prevStep")
                    append(" | step_delta_ms=$stepDeltaMs")
                }
                extraFields.forEach { (k, v) -> append(" | $k=$v") }
            }
            android.util.Log.d("AssistantTrace", fields)
        }

        fun logDelta(fromStep: String, toStep: String, extraFields: Map<String, String> = emptyMap()) {
            val from = checkpoints[fromStep]
            val to = checkpoints[toStep] ?: System.currentTimeMillis()
            val deltaMs = if (from != null) to - from else 0L

            val fields = buildString {
                append("flow=$flow")
                append(" | request_id=$requestId")
                append(" | delta=$fromStep->$toStep")
                append(" | delta_ms=$deltaMs")
                extraFields.forEach { (k, v) -> append(" | $k=$v") }
            }
            android.util.Log.d("AssistantTrace", fields)
        }

        fun logTotal(metric: String, extraFields: Map<String, String> = emptyMap()) {
            val now = System.currentTimeMillis()
            val totalMs = now - requestStartMs

            val fields = buildString {
                append("flow=$flow")
                append(" | request_id=$requestId")
                append(" | metric=$metric")
                append(" | total_ms=$totalMs")
                extraFields.forEach { (k, v) -> append(" | $k=$v") }
            }
            android.util.Log.d("AssistantTrace", fields)
        }
    }

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

        // Log final timing
        activeTimingLogger?.markStep("request_completed", mapOf("source" to source, "elapsed_ms" to (elapsed ?: -1).toString()))
        activeTimingLogger?.logTotal("total_e2e_duration")
        activeTimingLogger = null

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

        // Log final timing
        activeTimingLogger?.markStep("request_failed", mapOf("error" to (error ?: "unknown")))
        activeTimingLogger?.logTotal("total_e2e_duration")
        activeTimingLogger = null

        if (error != null) {
            lastAssistantError = error
        }
        transitionTo(AssistantState.Idle, "failRequest", requestId)
    }

    /**
     * Completes a request with a friendly service-issue message.
     * Used for recoverable transport/network/server failures that should not show as errors.
     */
    private fun completeWithServiceIssue(requestId: String?, source: String) {
        val currentRequestId = activeRequestId
        if (requestId != null && (currentRequestId == null || requestId != currentRequestId)) {
            android.util.Log.w(
                "AiAssistantViewModel",
                "completeWithServiceIssue ignored (stale/unowned RID). " +
                    "Expected: $currentRequestId, Got: $requestId, State: $assistantState"
            )
            return
        }

        val elapsed = requestId?.let { rid ->
            requestStartedAtMs[rid]?.let { start ->
                System.currentTimeMillis() - start
            }
        }

        // Log final timing
        activeTimingLogger?.markStep("service_issue_completed", mapOf("source" to source, "elapsed_ms" to (elapsed ?: -1).toString()))
        activeTimingLogger?.logTotal("total_e2e_duration")
        activeTimingLogger = null

        // Friendly service-issue message (matches backend ServiceIssue skill)
        val serviceIssueMessage = when (selectedLanguage) {
            "en" -> "Sorry, the AI service or network is having trouble right now. Please try again shortly."
            else -> "Maaf, koneksi atau layanan AI sedang bermasalah. Coba lagi sebentar ya."
        }

        transcriptionResults = transcriptionResults + TranscriptionMessage(
            text = serviceIssueMessage,
            role = MessageRole.ASSISTANT,
            requestId = requestId,
            finishedInMs = elapsed,
            source = source
        )
        transitionTo(AssistantState.Idle, "completeWithServiceIssue ($source)", requestId)
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
                                activeTimingLogger?.markStep("response_dropped", mapOf("reason" to "stale_or_foreign_rid", "expected_rid" to (currentRequestId ?: "null"), "got_rid" to (responseRequestId ?: "null")))
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
                                activeTimingLogger?.markStep("response_dropped", mapOf("reason" to "ack_only_source", "source" to source))
                                android.util.Log.d(
                                    "AiAssistantViewModel",
                                    "Ignored sync/dup drop from $source for RID: $currentRequestId"
                                )
                                return@collect
                            }

                            activeTimingLogger?.markStep("mqtt_answer_received", mapOf("source" to (source ?: "unknown")))

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
                activeTimingLogger = AssistantTimingLogger(flow = "manual_chat_text", requestId = requestId)
                activeTimingLogger?.markStep("send_chat_invoked")
                activeTimingLogger?.markStep("request_id_assigned")
                
                val publishStartMs = System.currentTimeMillis()
                activeTimingLogger?.markStep("mqtt_publish_started")
                val result = mqttHelper.publishChat(text, selectedLanguage, requestId)
                val publishDurationMs = System.currentTimeMillis() - publishStartMs
                
                if (result.isSuccess) {
                    activeTimingLogger?.markStep("mqtt_publish_success", mapOf("publish_duration_ms" to publishDurationMs.toString()))
                    activeTimingLogger?.logDelta("mqtt_publish_started", "mqtt_publish_success", mapOf("metric" to "publish_duration"))
                    startResponseTimeout(requestId)
                } else {
                    activeTimingLogger?.markStep("mqtt_publish_failed", mapOf("publish_duration_ms" to publishDurationMs.toString()))
                    completeWithServiceIssue(requestId, "mqtt_publish_fail")
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
        activeTimingLogger = AssistantTimingLogger(flow = "manual_chat_audio", requestId = requestId)
        activeTimingLogger?.markStep("recording_started")
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
            activeTimingLogger?.markStep("recording_failed", mapOf("error" to error))
            failRequest(requestId, error)
            android.util.Log.e("AiAssistantViewModel", "startRecording failed: $lastAssistantError")
        }
    }

    fun stopRecording() {
        if (assistantState == AssistantState.Recording) {
            audioRecorder.stop()

            val file = currentRecordingFile
            activeTimingLogger?.markStep("recording_stopped")
            if (file != null && audioRecorder.isWavValid(file)) {
                audioRecorder.finalizeWav(file)
                activeTimingLogger?.markStep("wav_validated")
                activeTimingLogger?.logDelta("recording_started", "recording_stopped", mapOf("metric" to "recording_duration"))
                transitionTo(AssistantState.Processing, "stopRecording(valid)")
                val requestId = activeRequestId
                viewModelScope.launch {
                    val publishStartMs = System.currentTimeMillis()
                    activeTimingLogger?.markStep("mqtt_publish_started")
                    val bytes = file.readBytes()
                    val result = mqttHelper.publishAudio(bytes, selectedLanguage, requestId)
                    val publishDurationMs = System.currentTimeMillis() - publishStartMs
                    
                    if (result.isSuccess) {
                        activeTimingLogger?.markStep("mqtt_publish_success", mapOf("publish_duration_ms" to publishDurationMs.toString()))
                        activeTimingLogger?.logDelta("mqtt_publish_started", "mqtt_publish_success", mapOf("metric" to "publish_duration"))
                        startResponseTimeout(requestId)
                    } else {
                        activeTimingLogger?.markStep("mqtt_publish_failed", mapOf("publish_duration_ms" to publishDurationMs.toString()))
                        completeWithServiceIssue(requestId, "mqtt_audio_publish_fail")
                        android.util.Log.e("AiAssistantViewModel", "publishAudio failed")
                    }
                }
            } else {
                activeTimingLogger?.markStep("recording_invalid", mapOf("reason" to "too_short_or_invalid"))
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
        activeTimingLogger?.markStep("http_fallback_started", mapOf("type" to (type?.name ?: "unknown")))

        viewModelScope.launch {
            val fallbackInitStartMs = System.currentTimeMillis()
            val username = com.example.whisperandroid.util.DeviceUtils.getDeviceId(getApplication())
            val tm = com.example.whisperandroid.data.di.NetworkModule.tokenManager
            val token = tm.getAccessToken()
            val terminalId = tm.getTerminalId() ?: username
            val macAddress = tm.getMacAddress() ?: username
            val fallbackInitDurationMs = System.currentTimeMillis() - fallbackInitStartMs

            if (token == null) {
                activeTimingLogger?.markStep("http_fallback_failed", mapOf("reason" to "auth_error", "init_duration_ms" to fallbackInitDurationMs.toString()))
                failRequest(requestId, "Authentication error: Token missing")
                return@launch
            }

            if (type == RequestType.Chat) {
                val text = activeRequestText
                if (text != null) {
                    activeTimingLogger?.markStep("http_fallback_chat_initiated", mapOf("init_duration_ms" to fallbackInitDurationMs.toString()))
                    val httpChatStartMs = System.currentTimeMillis()
                    com.example.whisperandroid.data.di.NetworkModule.ragRepository.chat(
                        prompt = text,
                        language = selectedLanguage,
                        terminalId = terminalId,
                        uid = username,
                        token = token,
                        requestId = requestId,
                        idempotencyKey = fallbackIdempotencyKey
                    ).collect { resource ->
                        if (activeRequestId != requestId) return@collect
                        when (resource) {
                            is com.example.whisperandroid.domain.repository.Resource.Success -> {
                                val httpChatDurationMs = System.currentTimeMillis() - httpChatStartMs
                                val data = resource.data
                                if (data != null) {
                                    val responseText = data.response
                                    if (responseText != null) {
                                        activeTimingLogger?.markStep("http_fallback_chat_success", mapOf("http_chat_duration_ms" to httpChatDurationMs.toString()))
                                        handleHttpChatResponse(responseText, false, requestId)
                                    } else if (data.source == "HTTP_DUP_DROP") {
                                        activeTimingLogger?.markStep("http_fallback_chat_dup_drop", mapOf("http_chat_duration_ms" to httpChatDurationMs.toString()))
                                        android.util.Log.d("AiAssistantViewModel", "Fallback: Silent completion for HTTP_DUP_DROP")
                                        completeRequest(requestId, null, null, "http_dup")
                                    } else {
                                        activeTimingLogger?.markStep("http_fallback_chat_failed", mapOf("reason" to "null_response", "http_chat_duration_ms" to httpChatDurationMs.toString()))
                                        failRequest(requestId, "Invalid fallback response (null body)")
                                    }
                                } else {
                                    activeTimingLogger?.markStep("http_fallback_chat_failed", mapOf("reason" to "null_data", "http_chat_duration_ms" to httpChatDurationMs.toString()))
                                    failRequest(requestId, "Invalid fallback response")
                                }
                            }
                            is com.example.whisperandroid.domain.repository.Resource.Error -> {
                                activeTimingLogger?.markStep("http_fallback_chat_error")
                                completeWithServiceIssue(requestId, "http_fallback_error")
                            }
                            is com.example.whisperandroid.domain.repository.Resource.Loading -> {}
                        }
                    }
                } else {
                    activeTimingLogger?.markStep("http_fallback_failed", mapOf("reason" to "missing_text"))
                    failRequest(requestId, "Cannot run fallback: text is missing")
                }
            } else if (type == RequestType.Audio) {
                val file = currentRecordingFile
                if (file != null && file.exists()) {
                    activeTimingLogger?.markStep("http_fallback_audio_initiated", mapOf("init_duration_ms" to fallbackInitDurationMs.toString()))
                    val transcribeAudioUseCase =
                        com.example.whisperandroid.data.di.NetworkModule.transcribeAudioUseCase
                    val httpTranscribeStartMs = System.currentTimeMillis()
                    transcribeAudioUseCase.initiate(
                        audioFile = file,
                        token = token,
                        language = selectedLanguage,
                        macAddress = macAddress,
                        idempotencyKey = fallbackIdempotencyKey
                    ).collect { result ->
                        if (activeRequestId != requestId) return@collect
                        when (result) {
                            is com.example.whisperandroid.domain.repository.Resource.Success -> {
                                val httpTranscribeDurationMs = System.currentTimeMillis() - httpTranscribeStartMs
                                val taskId = result.data
                                if (taskId != null) {
                                    activeTimingLogger?.markStep("http_fallback_transcribe_success", mapOf("task_id" to taskId, "http_transcribe_duration_ms" to httpTranscribeDurationMs.toString()))
                                    pollHttpTranscription(
                                        taskId,
                                        requestId,
                                        token,
                                        terminalId,
                                        fallbackIdempotencyKey
                                    )
                                } else {
                                    activeTimingLogger?.markStep("http_fallback_transcribe_failed", mapOf("reason" to "null_task_id", "http_transcribe_duration_ms" to httpTranscribeDurationMs.toString()))
                                    failRequest(requestId, "Fallback audio trigger failed")
                                }
                            }
                            is com.example.whisperandroid.domain.repository.Resource.Error -> {
                                activeTimingLogger?.markStep("http_fallback_transcribe_error", mapOf("http_transcribe_duration_ms" to (System.currentTimeMillis() - httpTranscribeStartMs).toString()))
                                completeWithServiceIssue(requestId, "http_audio_fallback_error")
                            }
                            is com.example.whisperandroid.domain.repository.Resource.Loading -> {}
                        }
                    }
                } else {
                    activeTimingLogger?.markStep("http_fallback_failed", mapOf("reason" to "file_not_found"))
                    failRequest(requestId, "Audio file missing for fallback")
                }
            } else {
                activeTimingLogger?.markStep("http_fallback_failed", mapOf("reason" to "unknown_request_type"))
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
        viewModelScope.launch {
            val pollStartMs = System.currentTimeMillis()
            activeTimingLogger?.markStep("http_poll_started", mapOf("task_id" to taskId))
            
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
                    val pollDurationMs = System.currentTimeMillis() - pollStartMs
                    val transcribedText = currentSuccessText!!
                    activeTimingLogger?.markStep("http_poll_transcription_completed", mapOf("poll_duration_ms" to pollDurationMs.toString(), "attempts" to attempts.toString()))

                    appendUserMessageIfNeeded(transcribedText)

                    val httpPollChatStartMs = System.currentTimeMillis()
                    com.example.whisperandroid.data.di.NetworkModule.ragRepository.chat(
                        prompt = transcribedText,
                        language = selectedLanguage,
                        terminalId = terminalId,
                        uid = username,
                        token = token,
                        requestId = requestId,
                        idempotencyKey = fallbackIdempotencyKey + "_chat"
                    ).collect { chatResult ->
                        if (activeRequestId != requestId) return@collect
                        val httpPollChatDurationMs = System.currentTimeMillis() - httpPollChatStartMs
                        when (chatResult) {
                            is com.example.whisperandroid.domain.repository.Resource.Success -> {
                                val data = chatResult.data
                                if (data != null) {
                                    val responseText = data.response
                                    if (responseText != null) {
                                        activeTimingLogger?.markStep("http_poll_chat_success", mapOf("http_poll_chat_duration_ms" to httpPollChatDurationMs.toString()))
                                        handleHttpChatResponse(responseText, false, requestId)
                                    } else if (data.source == "HTTP_DUP_DROP") {
                                        activeTimingLogger?.markStep("http_poll_chat_dup_drop", mapOf("http_poll_chat_duration_ms" to httpPollChatDurationMs.toString()))
                                        android.util.Log.d("AiAssistantViewModel", "Poll: Silent completion for HTTP_DUP_DROP")
                                        completeRequest(requestId, null, null, "http_dup")
                                    } else {
                                        activeTimingLogger?.markStep("http_poll_chat_failed", mapOf("reason" to "null_response", "http_poll_chat_duration_ms" to httpPollChatDurationMs.toString()))
                                        failRequest(requestId, "Invalid poll response (null body)")
                                    }
                                } else {
                                    activeTimingLogger?.markStep("http_poll_chat_failed", mapOf("reason" to "null_data", "http_poll_chat_duration_ms" to httpPollChatDurationMs.toString()))
                                    failRequest(requestId, "Invalid poll response")
                                }
                            }
                            is com.example.whisperandroid.domain.repository.Resource.Error -> {
                                activeTimingLogger?.markStep("http_poll_chat_error", mapOf("http_poll_chat_duration_ms" to httpPollChatDurationMs.toString()))
                                completeWithServiceIssue(requestId, "http_poll_error")
                            }
                            is com.example.whisperandroid.domain.repository.Resource.Loading -> {}
                        }
                    }
                } else if (hasError) {
                    activeTimingLogger?.markStep("http_poll_error")
                    completeWithServiceIssue(requestId, "transcription_poll_fail")
                    return@launch
                } else {
                    attempts++
                    kotlinx.coroutines.delay(1500L) // 1.5s delay between polls
                }
            }
            if (!isCompleted) {
                completeWithServiceIssue(requestId, "polling_timeout")
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
