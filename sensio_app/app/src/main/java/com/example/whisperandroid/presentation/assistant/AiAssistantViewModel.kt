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

    private var lastUserMessageNormalized: String? = null
    private var lastUserMessageAtMs: Long = 0L

    // Latency tracking
    private val requestStartedAtMs = mutableMapOf<String, Long>()
    private var activeTimingLogger: AssistantTimingLogger? = null

    // Fallback gate to prevent overlapping fallback loops for the same request
    private var fallbackRunningForRequestId: String? = null

    // Track recent request IDs to handle race conditions (e.g., retry with new ID before old response arrives)
    private val recentRequestIds = ArrayDeque<String>()

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


    // Helper to track request ID for recent request matching
    private fun trackRequestId(requestId: String) {
        recentRequestIds.add(requestId)
        if (recentRequestIds.size > 5) {
            recentRequestIds.removeFirst()
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

        // Reset fallback gate on terminal completion
        if (requestId == fallbackRunningForRequestId) {
            fallbackRunningForRequestId = null
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

        // Log final timing
        activeTimingLogger?.markStep("request_failed", mapOf("error" to (error ?: "unknown")))
        activeTimingLogger?.logTotal("total_e2e_duration")
        activeTimingLogger = null

        // Reset fallback gate on terminal completion
        if (requestId == fallbackRunningForRequestId) {
            fallbackRunningForRequestId = null
        }

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

        // Reset fallback gate on terminal completion
        if (requestId == fallbackRunningForRequestId) {
            fallbackRunningForRequestId = null
        }

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
            mqttHelper.connectionStatus.collect { status ->
                android.util.Log.d("AiAssistantViewModel", "MQTT Status Changed: $status")
                mqttStatus = status
            }
        }

        // MQTT connection is handled by Foreground Service, not ViewModel
    }

    fun reconnectMqtt() {
        viewModelScope.launch {
            if (mqttHelper.connectionStatus.value == MqttHelper.MqttConnectionStatus.CONNECTED ||
                mqttHelper.connectionStatus.value == MqttHelper.MqttConnectionStatus.CONNECTING
            ) {
                return@launch
            }
            android.util.Log.d("AiAssistantViewModel", "Manual MQTT Reconnection...")
            // connect() now fetches credentials internally, password is never stored
            mqttHelper.connect()
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
            trackRequestId(activeRequestId!!)  // Track for recent request matching
            activeRequestType = RequestType.Chat
            activeRequestText = text
            transitionTo(AssistantState.Processing, "sendChat")

            viewModelScope.launch {
                val requestId = activeRequestId ?: return@launch
                requestStartedAtMs[requestId] = System.currentTimeMillis()
                activeTimingLogger = AssistantTimingLogger(flow = "manual_chat_text", requestId = requestId)
                activeTimingLogger?.markStep("send_chat_invoked")
                activeTimingLogger?.markStep("request_id_assigned")

                val username = com.example.whisperandroid.util.DeviceUtils.getDeviceId(getApplication())
                val tm = com.example.whisperandroid.data.di.NetworkModule.tokenManager
                val token = tm.getAccessToken()
                val terminalId = tm.getTerminalId() ?: username

                if (token == null) {
                    activeTimingLogger?.markStep("http_chat_failed", mapOf("reason" to "auth_error"))
                    failRequest(requestId, "Authentication error: Token missing")
                    return@launch
                }

                activeTimingLogger?.markStep("http_chat_started")
                val httpChatStartMs = System.currentTimeMillis()

                com.example.whisperandroid.data.di.NetworkModule.ragRepository.chat(
                    prompt = text,
                    language = selectedLanguage,
                    terminalId = terminalId,
                    uid = username,
                    token = token,
                    requestId = requestId,
                    idempotencyKey = requestId
                ).collect { resource ->
                    if (activeRequestId != requestId) return@collect
                    when (resource) {
                        is com.example.whisperandroid.domain.repository.Resource.Success -> {
                            val httpChatDurationMs = System.currentTimeMillis() - httpChatStartMs
                            val data = resource.data
                            if (data != null) {
                                val parsedResult = AssistantResponseParser.parseHttpAssistantResult(data)
                                if (parsedResult.isDupInProgress) {
                                    activeTimingLogger?.markStep("http_chat_in_progress", mapOf("http_chat_duration_ms" to httpChatDurationMs.toString()))
                                    android.util.Log.d("AiAssistantViewModel", "HTTP Chat: Request in progress, starting bounded retry")
                                    launchBoundedRetryForInProgress(
                                        requestId = requestId,
                                        text = text,
                                        selectedLanguage = selectedLanguage,
                                        terminalId = terminalId,
                                        username = username,
                                        token = token,
                                        fallbackIdempotencyKey = requestId,
                                        httpChatStartMs = httpChatStartMs
                                    )
                                } else if (parsedResult.isDupCached) {
                                    activeTimingLogger?.markStep("http_chat_dup_drop", mapOf("http_chat_duration_ms" to httpChatDurationMs.toString()))
                                    android.util.Log.d("AiAssistantViewModel", "HTTP Chat: Silent completion for cached duplicate")
                                    completeRequest(requestId, null, null, "http_dup")
                                } else {
                                    activeTimingLogger?.markStep("http_chat_success", mapOf("http_chat_duration_ms" to httpChatDurationMs.toString()))
                                    handleHttpChatResponse(parsedResult, requestId)
                                }
                            } else {
                                activeTimingLogger?.markStep("http_chat_failed", mapOf("reason" to "null_data"))
                                failRequest(requestId, "Invalid response")
                            }
                        }
                        is com.example.whisperandroid.domain.repository.Resource.Error -> {
                            activeTimingLogger?.markStep("http_chat_error", mapOf("error" to (resource.message ?: "unknown")))
                            val errorMsg = resource.message
                            if (errorMsg != null && (errorMsg.contains("Maaf") || errorMsg.contains("Sorry"))) {
                                completeRequest(requestId, errorMsg, null, "http_error_msg")
                            } else {
                                completeWithServiceIssue(requestId, "http_chat_error")
                            }
                        }
                        is com.example.whisperandroid.domain.repository.Resource.Loading -> {}
                    }
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
        trackRequestId(requestId)  // Track for recent request matching
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
                    val username = com.example.whisperandroid.util.DeviceUtils.getDeviceId(getApplication())
                    val tm = com.example.whisperandroid.data.di.NetworkModule.tokenManager
                    val token = tm.getAccessToken()
                    val terminalId = tm.getTerminalId() ?: username
                    val macAddress = tm.getMacAddress() ?: username

                    if (token == null) {
                        activeTimingLogger?.markStep("http_transcribe_failed", mapOf("reason" to "auth_error"))
                        failRequest(requestId, "Authentication error: Token missing")
                        return@launch
                    }

                    activeTimingLogger?.markStep("http_transcribe_started")
                    val transcribeAudioUseCase =
                        com.example.whisperandroid.data.di.NetworkModule.transcribeAudioUseCase
                    val httpTranscribeStartMs = System.currentTimeMillis()
                    transcribeAudioUseCase.initiate(
                        audioFile = file,
                        token = token,
                        language = selectedLanguage,
                        macAddress = macAddress,
                        idempotencyKey = requestId ?: ""
                    ).collect { result ->
                        if (activeRequestId != requestId) return@collect
                        when (result) {
                            is com.example.whisperandroid.domain.repository.Resource.Success -> {
                                val httpTranscribeDurationMs = System.currentTimeMillis() - httpTranscribeStartMs
                                val taskId = result.data
                                if (taskId != null) {
                                    activeTimingLogger?.markStep("http_transcribe_success", mapOf("task_id" to taskId, "http_transcribe_duration_ms" to httpTranscribeDurationMs.toString()))
                                    pollHttpTranscription(
                                        taskId,
                                        requestId ?: "",
                                        token,
                                        terminalId,
                                        requestId ?: ""
                                    )
                                } else {
                                    activeTimingLogger?.markStep("http_transcribe_failed", mapOf("reason" to "null_task_id"))
                                    failRequest(requestId, "Audio trigger failed")
                                }
                            }
                            is com.example.whisperandroid.domain.repository.Resource.Error -> {
                                activeTimingLogger?.markStep("http_transcribe_error", mapOf("error" to (result.message ?: "unknown")))
                                completeWithServiceIssue(requestId, "http_transcribe_error")
                            }
                            is com.example.whisperandroid.domain.repository.Resource.Loading -> {}
                        }
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
            kotlinx.coroutines.delay(60000L) // Wait up to 60s before HTTP fallback (extended for slow backend responses)
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

        // Prevent overlapping fallback loops for the same request
        if (fallbackRunningForRequestId == requestId) {
            android.util.Log.w("AiAssistantViewModel", "Fallback already running for RID: $requestId")
            return
        }
        fallbackRunningForRequestId = requestId

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
                                    val parsedResult = AssistantResponseParser.parseHttpAssistantResult(data)
                                    if (parsedResult.isDupInProgress) {
                                        // IDEMPOTENCY_IN_PROGRESS: First request still processing
                                        // Launch bounded retry loop to re-check HTTP chat status
                                        activeTimingLogger?.markStep("http_fallback_chat_in_progress", mapOf("http_chat_duration_ms" to httpChatDurationMs.toString()))
                                        android.util.Log.d("AiAssistantViewModel", "Fallback: Request in progress, starting bounded retry")
                                        launchBoundedRetryForInProgress(
                                            requestId = requestId,
                                            text = text,
                                            selectedLanguage = selectedLanguage,
                                            terminalId = terminalId,
                                            username = username,
                                            token = token,
                                            fallbackIdempotencyKey = fallbackIdempotencyKey,
                                            httpChatStartMs = httpChatStartMs
                                        )
                                    } else if (parsedResult.isDupCached) {
                                        // IDEMPOTENCY_CACHED: Duplicate with completed response, silent finish
                                        activeTimingLogger?.markStep("http_fallback_chat_dup_drop", mapOf("http_chat_duration_ms" to httpChatDurationMs.toString()))
                                        android.util.Log.d("AiAssistantViewModel", "Fallback: Silent completion for cached duplicate")
                                        completeRequest(requestId, null, null, "http_dup")
                                    } else {
                                        activeTimingLogger?.markStep("http_fallback_chat_success", mapOf("http_chat_duration_ms" to httpChatDurationMs.toString()))
                                        handleHttpChatResponse(parsedResult, requestId)
                                    }
                                } else {
                                    activeTimingLogger?.markStep("http_fallback_chat_failed", mapOf("reason" to "null_data", "http_chat_duration_ms" to httpChatDurationMs.toString()))
                                    failRequest(requestId, "Invalid fallback response")
                                }
                            }
                            is com.example.whisperandroid.domain.repository.Resource.Error -> {
                                activeTimingLogger?.markStep("http_fallback_chat_error", mapOf("error" to (resource.message ?: "unknown")))
                                // If error contains a backend message (e.g. from RagRepositoryImpl), show it directly
                                val errorMsg = resource.message
                                if (errorMsg != null && (errorMsg.contains("Maaf") || errorMsg.contains("Sorry"))) {
                                    completeRequest(requestId, errorMsg, null, "http_fallback_error_msg")
                                } else {
                                    completeWithServiceIssue(requestId, "http_fallback_error")
                                }
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
                                    val parsedResult = AssistantResponseParser.parseHttpAssistantResult(data)
                                    if (parsedResult.isDupInProgress) {
                                        // IDEMPOTENCY_IN_PROGRESS: First request still processing, continue waiting
                                        activeTimingLogger?.markStep("http_poll_chat_in_progress", mapOf("http_poll_chat_duration_ms" to httpPollChatDurationMs.toString()))
                                        android.util.Log.d("AiAssistantViewModel", "Poll: Request in progress, continue waiting")
                                        // Do NOT completeRequest - continue polling for real result
                                    } else if (parsedResult.isDupCached) {
                                        // IDEMPOTENCY_CACHED: Duplicate with completed response, silent finish
                                        activeTimingLogger?.markStep("http_poll_chat_dup_drop", mapOf("http_poll_chat_duration_ms" to httpPollChatDurationMs.toString()))
                                        android.util.Log.d("AiAssistantViewModel", "Poll: Silent completion for cached duplicate")
                                        completeRequest(requestId, null, null, "http_dup")
                                    } else {
                                        activeTimingLogger?.markStep("http_poll_chat_success", mapOf("http_poll_chat_duration_ms" to httpPollChatDurationMs.toString()))
                                        handleHttpChatResponse(parsedResult, requestId)
                                    }
                                } else {
                                    activeTimingLogger?.markStep("http_poll_chat_failed", mapOf("reason" to "null_data", "http_poll_chat_duration_ms" to httpPollChatDurationMs.toString()))
                                    failRequest(requestId, "Invalid poll response")
                                }
                            }
                            is com.example.whisperandroid.domain.repository.Resource.Error -> {
                                activeTimingLogger?.markStep("http_poll_chat_error", mapOf("http_poll_chat_duration_ms" to httpPollChatDurationMs.toString(), "error" to (chatResult.message ?: "unknown")))
                                // Preserve backend error message
                                val errorMsg = chatResult.message
                                if (errorMsg != null && (errorMsg.contains("Maaf") || errorMsg.contains("Sorry"))) {
                                    completeRequest(requestId, errorMsg, null, "http_poll_error_msg")
                                } else {
                                    completeWithServiceIssue(requestId, "http_poll_error")
                                }
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

    /**
     * Launches a bounded retry loop for HTTP chat when response is IDEMPOTENCY_IN_PROGRESS.
     * Re-checks chat status with exponential backoff until:
     * - Final payload received (!isDupInProgress)
     * - Request no longer active (activeRequestId changed)
     * - Retry budget exhausted (then fails gracefully with service issue message)
     */
    private fun launchBoundedRetryForInProgress(
        requestId: String,
        text: String,
        selectedLanguage: String,
        terminalId: String,
        username: String,
        token: String,
        fallbackIdempotencyKey: String,
        httpChatStartMs: Long
    ) {
        viewModelScope.launch {
            val maxRetries = 4
            val baseDelayMs = 1000L
            var attempt = 0
            val retryStartMs = System.currentTimeMillis()

            activeTimingLogger?.markStep("http_retry_loop_started", mapOf("max_retries" to maxRetries.toString(), "base_delay_ms" to baseDelayMs.toString()))

            while (attempt < maxRetries) {
                if (activeRequestId != requestId) {
                    activeTimingLogger?.markStep("http_retry_loop_abandoned", mapOf("reason" to "request_changed", "attempt" to (attempt + 1).toString()))
                    android.util.Log.d("AiAssistantViewModel", "Retry loop abandoned: request changed")
                    return@launch
                }

                attempt++
                activeTimingLogger?.markStep("http_retry_attempt", mapOf("attempt" to attempt.toString(), "max" to maxRetries.toString()))
                android.util.Log.d("AiAssistantViewModel", "Retry attempt $attempt/$maxRetries for RID: $requestId")

                // Re-check chat status
                val retryChatStartMs = System.currentTimeMillis()
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
                            val data = resource.data
                            if (data != null) {
                                val parsedResult = AssistantResponseParser.parseHttpAssistantResult(data)
                                val retryDurationMs = System.currentTimeMillis() - retryChatStartMs

                                if (!parsedResult.isDupInProgress && !parsedResult.isDupCached) {
                                    // Final response received
                                    activeTimingLogger?.markStep("http_retry_success", mapOf("attempt" to attempt.toString(), "retry_duration_ms" to retryDurationMs.toString()))
                                    activeTimingLogger?.markStep("http_fallback_chat_success", mapOf("http_chat_duration_ms" to (System.currentTimeMillis() - httpChatStartMs).toString()))
                                    handleHttpChatResponse(parsedResult, requestId)
                                    return@collect
                                } else if (parsedResult.isDupCached) {
                                    // Cached response received
                                    activeTimingLogger?.markStep("http_retry_cached", mapOf("attempt" to attempt.toString(), "retry_duration_ms" to retryDurationMs.toString()))
                                    activeTimingLogger?.markStep("http_fallback_chat_dup_drop", mapOf("http_chat_duration_ms" to (System.currentTimeMillis() - httpChatStartMs).toString()))
                                    completeRequest(requestId, null, null, "http_dup")
                                    return@collect
                                } else {
                                    // Still in progress, continue retry
                                    activeTimingLogger?.markStep("http_retry_still_in_progress", mapOf("attempt" to attempt.toString(), "retry_duration_ms" to retryDurationMs.toString()))
                                    android.util.Log.d("AiAssistantViewModel", "Retry $attempt: Still in progress")
                                }
                            } else {
                                activeTimingLogger?.markStep("http_retry_failed", mapOf("attempt" to attempt.toString(), "reason" to "null_data"))
                                failRequest(requestId, "Invalid retry response")
                                return@collect
                            }
                        }
                        is com.example.whisperandroid.domain.repository.Resource.Error -> {
                            activeTimingLogger?.markStep("http_retry_error", mapOf("attempt" to attempt.toString(), "error" to (resource.message ?: "unknown")))
                            val errorMsg = resource.message
                            if (errorMsg != null && (errorMsg.contains("Maaf") || errorMsg.contains("Sorry"))) {
                                completeRequest(requestId, errorMsg, null, "http_retry_error_msg")
                            } else {
                                completeWithServiceIssue(requestId, "http_retry_error")
                            }
                            return@collect
                        }
                        is com.example.whisperandroid.domain.repository.Resource.Loading -> {}
                    }
                }

                // If we reach here, still in progress - apply backoff before next retry
                if (attempt < maxRetries) {
                    val delayMs = baseDelayMs * (1L shl (attempt - 1)) // Exponential backoff: 1s, 2s, 4s, ...
                    activeTimingLogger?.markStep("http_retry_backoff", mapOf("attempt" to attempt.toString(), "delay_ms" to delayMs.toString()))
                    kotlinx.coroutines.delay(delayMs)
                }
            }

            // Retry budget exhausted - fail gracefully
            val totalRetryDurationMs = System.currentTimeMillis() - retryStartMs
            activeTimingLogger?.markStep("http_retry_exhausted", mapOf("total_retries" to attempt.toString(), "total_retry_duration_ms" to totalRetryDurationMs.toString()))
            android.util.Log.w("AiAssistantViewModel", "Retry budget exhausted for RID: $requestId after $attempt attempts")
            completeWithServiceIssue(requestId, "http_retry_timeout")
        }
    }

    private fun handleHttpChatResponse(
        result: ParsedAssistantChatResult,
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

        if (result.isBlocked) {
            val userMessages = transcriptionResults.filter { it.role == MessageRole.USER }
            if (userMessages.isNotEmpty()) {
                transcriptionResults = transcriptionResults.toMutableList().apply {
                    val lastUserIndex = indexOfLast { it.role == MessageRole.USER }
                    if (lastUserIndex >= 0) removeAt(lastUserIndex)
                }
            }
            android.util.Log.d("AiAssistantViewModel", "HTTP Fallback: Guard blocked prompt")
        }

        val cleanMessage = AssistantResponseParser.getCleanMessage(result, selectedLanguage)

        // Handle empty terminal payload - convert to service-issue instead of silent complete
        if (cleanMessage == null && !result.isBlocked && !result.isControl) {
            activeTimingLogger?.markStep("empty_terminal_payload", mapOf("source" to (result.source ?: "http_fallback")))
            android.util.Log.w(
                "AiAssistantViewModel",
                "Empty terminal payload in HTTP fallback, completing with service-issue. RID: $currentRequestId"
            )
            completeWithServiceIssue(currentRequestId, "empty_terminal_payload_http")
        } else {
            completeRequest(currentRequestId, cleanMessage, elapsed, "http_fallback")
        }
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
