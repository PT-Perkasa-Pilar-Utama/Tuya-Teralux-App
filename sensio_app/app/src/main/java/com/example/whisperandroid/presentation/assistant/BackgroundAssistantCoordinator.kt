package com.example.whisperandroid.presentation.assistant

import android.app.Application
import android.media.MediaPlayer
import com.example.whisperandroid.R
import com.example.whisperandroid.data.di.NetworkModule
import com.example.whisperandroid.domain.repository.Resource
import com.example.whisperandroid.presentation.meeting.AudioRecorder
import com.example.whisperandroid.util.AppLog
import com.example.whisperandroid.util.DeviceUtils
import com.example.whisperandroid.util.MqttHelper
import com.example.whisperandroid.util.parseMarkdownToText
import com.google.gson.JsonParser
import java.io.File
import java.util.UUID
import kotlin.coroutines.resume
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.collect
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import kotlinx.coroutines.suspendCancellableCoroutine
import kotlinx.coroutines.withTimeout

class BackgroundAssistantCoordinator(
    private val application: Application,
    private val mqttHelperProvider: () -> com.example.whisperandroid.util.MqttHelper = { NetworkModule.mqttHelper },
    private val audioRecorderProvider: (android.app.Application) -> com.example.whisperandroid.presentation.meeting.AudioRecorder = { com.example.whisperandroid.presentation.meeting.AudioRecorder(it) }
) {
    private val TAG = "Coordinator"
    private var scope: CoroutineScope? = null
    private val _uiState = MutableStateFlow(BackgroundAssistantUiState())
    val uiState = _uiState.asStateFlow()

    private val mqttHelper by lazy { mqttHelperProvider() }
    private val audioRecorder by lazy { audioRecorderProvider(application) }
    private var currentRecordingFile: File? = null
    private var activeRequestId: String? = null
    private val requestStartedAtMs = mutableMapOf<String, Long>()
    private var selectedLanguage = "id"

    private var activeSessionJob: Job? = null
    private var listeningJob: Job? = null
    private var timeoutJob: Job? = null
    private var mqttCollectorJob: Job? = null
    private var fallbackJob: Job? = null
    private var greetingPlayer: MediaPlayer? = null

    // Timing logger instance for current session
    private var timingLogger: AssistantTimingLogger? = null

    var onDismissed: () -> Unit = {}

    val isSessionActive: Boolean
        get() = _uiState.value.state != BackgroundAssistantUiState.State.Hidden

    private companion object {
        const val GREETING_DURATION_MS = 1000L
        const val LISTENING_TICK_MS = 100L
        const val LISTENING_TIMEOUT_MS = 5000L
        const val PROCESSING_TIMEOUT_MS = 12000L
        // Note: RESULT and ERROR no longer auto-dismiss; user dismisses manually via outside tap
    }

    // Timing helper for structured latency logs
    private class AssistantTimingLogger(
        private val flow: String,
        private val requestId: String,
        private val sessionId: String?
    ) {
        private val sessionStartMs = System.currentTimeMillis()
        private val requestStartMs = System.currentTimeMillis()
        private val checkpoints = mutableMapOf<String, Long>()
        private var lastStep: String? = null

        fun markStep(step: String, extraFields: Map<String, String> = emptyMap()) {
            val now = System.currentTimeMillis()
            val elapsedMs = now - sessionStartMs
            val requestElapsedMs = now - requestStartMs
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
                if (sessionId != null) append(" | session_id=$sessionId")
                append(" | step=$step")
                append(" | elapsed_ms=$elapsedMs")
                append(" | request_elapsed_ms=$requestElapsedMs")
                if (prevStep != null && stepDeltaMs > 0) {
                    append(" | prev_step=$prevStep")
                    append(" | step_delta_ms=$stepDeltaMs")
                }
                extraFields.forEach { (k, v) -> append(" | $k=$v") }
            }
            AppLog.d("AssistantTrace", fields)
        }

        fun logDelta(fromStep: String, toStep: String, extraFields: Map<String, String> = emptyMap()) {
            val from = checkpoints[fromStep]
            val to = checkpoints[toStep] ?: System.currentTimeMillis()
            val deltaMs = if (from != null) to - from else 0L

            val fields = buildString {
                append("flow=$flow")
                append(" | request_id=$requestId")
                if (sessionId != null) append(" | session_id=$sessionId")
                append(" | delta=$fromStep->$toStep")
                append(" | delta_ms=$deltaMs")
                extraFields.forEach { (k, v) -> append(" | $k=$v") }
            }
            AppLog.d("AssistantTrace", fields)
        }

        fun logTotal(metric: String, extraFields: Map<String, String> = emptyMap()) {
            val now = System.currentTimeMillis()
            val totalMs = now - sessionStartMs
            val requestTotalMs = now - requestStartMs

            val fields = buildString {
                append("flow=$flow")
                append(" | request_id=$requestId")
                if (sessionId != null) append(" | session_id=$sessionId")
                append(" | metric=$metric")
                append(" | total_ms=$totalMs")
                append(" | request_total_ms=$requestTotalMs")
                extraFields.forEach { (k, v) -> append(" | $k=$v") }
            }
            AppLog.d("AssistantTrace", fields)
        }
    }

    fun start(serviceScope: CoroutineScope) {
        this.scope = serviceScope
        // Auto-connect MQTT in background
        serviceScope.launch {
            ensureMqttConnected()
        }
    }

    fun stop() {
        cancelPerSessionJobs()
        resetState()
        scope = null
        onDismissed = {}
    }

    fun onWakeDetected() {
        if (_uiState.value.state != BackgroundAssistantUiState.State.Hidden) {
            AppLog.w(TAG, "Ignoring wake word while session already active: ${_uiState.value.state}")
            return
        }
        AppLog.d(TAG, "Wake detected, starting real session")
        startRealSession()
    }

    private fun startRealSession() {
        cancelPerSessionJobs()
        activeSessionJob = scope?.launch {
            val sessionId = UUID.randomUUID().toString()
            val requestId = UUID.randomUUID().toString()
            activeRequestId = requestId
            requestStartedAtMs[requestId] = System.currentTimeMillis()

            // Initialize timing logger for this session
            timingLogger = AssistantTimingLogger(flow = "background", requestId = requestId, sessionId = sessionId)
            timingLogger?.markStep("wake_detected")
            timingLogger?.markStep("session_started")

            // 1. Greeting
            timingLogger?.markStep("greeting_started")
            playGreetingAudioAndAwait(sessionId)
            timingLogger?.markStep("greeting_finished")
            timingLogger?.logDelta("greeting_started", "greeting_finished", mapOf("metric" to "greeting_duration"))

            // 2. Listening
            startRecording(requestId)
        }
    }

    private fun startRecording(requestId: String) {
        val cacheDir = application.cacheDir
        val file = File(cacheDir, "bg_assistant_${System.currentTimeMillis()}.wav")
        currentRecordingFile = file

        timingLogger?.markStep("recording_started")
        transitionTo(BackgroundAssistantUiState.State.Listening, "startRecording")

        val result = audioRecorder.start(file)
        if (result !is AudioRecorder.RecorderResult.Success) {
            AppLog.e(TAG, "Mic initialization failed: $result")
            timingLogger?.markStep("recording_failed", mapOf("error" to "mic_init_failed"))
            failSession(requestId, "Microphone error")
            return
        }

        // Start mic level animation loop
        listeningJob = scope?.launch {
            val start = System.currentTimeMillis()
            while (isActive && System.currentTimeMillis() - start < LISTENING_TIMEOUT_MS) {
                _uiState.value = _uiState.value.copy(
                    micLevel = (0.3f + Math.random().toFloat() * 0.7f) // Animated pulse
                )
                delay(LISTENING_TICK_MS)
            }
            if (isActive) {
                AppLog.d(TAG, "Listening timeout, stopping")
                timingLogger?.markStep("listening_timeout_reached")
                timingLogger?.logDelta("recording_started", "listening_timeout_reached", mapOf("metric" to "listening_duration"))
                stopRecordingAndPublish(requestId)
            }
        }

        // Dedicated response collector (MQTT)
        startMqttCollector(requestId)
    }

    private fun stopRecordingAndPublish(requestId: String) {
        listeningJob?.cancel()
        audioRecorder.stop()

        timingLogger?.markStep("recording_stopped")
        timingLogger?.logDelta("recording_started", "recording_stopped", mapOf("metric" to "listening_duration"))

        val file = currentRecordingFile
        if (file != null && audioRecorder.isWavValid(file)) {
            audioRecorder.finalizeWav(file)
            timingLogger?.markStep("wav_validated")
            transitionTo(BackgroundAssistantUiState.State.Processing, "stopRecording(valid)")

            scope?.launch {
                ensureMqttConnected()
                val publishStartMs = System.currentTimeMillis()
                timingLogger?.markStep("mqtt_publish_started")
                val bytes = file.readBytes()
                val publishResult = mqttHelper.publishAudio(bytes, selectedLanguage, requestId)
                val publishDurationMs = System.currentTimeMillis() - publishStartMs
                if (publishResult.isSuccess) {
                    timingLogger?.markStep("mqtt_publish_success", mapOf("publish_duration_ms" to publishDurationMs.toString()))
                    timingLogger?.logDelta("mqtt_publish_started", "mqtt_publish_success", mapOf("metric" to "publish_duration"))
                    AppLog.d(TAG, "Audio published successfully, waiting for response")
                    startProcessingTimeout(requestId)
                } else {
                    timingLogger?.markStep("mqtt_publish_failed", mapOf("publish_duration_ms" to publishDurationMs.toString()))
                    AppLog.e(TAG, "Failed to publish audio via MQTT")
                    showServiceIssue(requestId, "mqtt_publish_fail")
                }
            }
        } else {
            timingLogger?.markStep("recording_invalid", mapOf("reason" to "too_short_or_invalid"))
            AppLog.e(TAG, "Recording invalid or too short")
            failSession(requestId, "Suara tidak terdengar")
        }
    }

    private fun startMqttCollector(requestId: String) {
        mqttCollectorJob?.cancel()
        mqttCollectorJob = scope?.launch {
            mqttHelper.messages.collect { (rawTopic, rawMessage) ->
                val topic = rawTopic as String
                val message = rawMessage as String

                if (topic.endsWith("chat/answer") || topic.endsWith("chat")) {
                    handleMqttMessage(topic, message, requestId)
                }
            }
        }
    }

    private fun handleMqttMessage(topic: String, message: String, currentRequestId: String) {
        try {
            val json = JsonParser.parseString(message).asJsonObject
            val responseRequestId = parseRequestId(json)

            if (topic.endsWith("chat")) {
                // Sync sync user prompt
                timingLogger?.markStep("mqtt_sync_received", mapOf("topic" to "chat"))
                val prompt = if (json.has("prompt")) json.get("prompt").asString else message
                if (responseRequestId == currentRequestId) {
                    _uiState.value = _uiState.value.copy(recognizedText = prompt.trim().removeSurrounding("\""))
                }
                return
            }

            // chat/answer handling
            if (responseRequestId == null || responseRequestId != currentRequestId) {
                timingLogger?.markStep("response_dropped", mapOf("reason" to "stale_or_foreign_rid", "expected_rid" to (currentRequestId ?: "null"), "got_rid" to (responseRequestId ?: "null")))
                AppLog.d(TAG, "Ignored response for foreign/stale RID: $responseRequestId (Current: $currentRequestId)")
                return
            }

            val source = parseSource(json)
            if (source == "MQTT_SYNC_DROP" || source == "MQTT_DUP_DROP") {
                timingLogger?.markStep("response_dropped", mapOf("reason" to "ack_only_source", "source" to (source ?: "null")))
                AppLog.d(TAG, "Ignored ack-only source: $source")
                return
            }

            timingLogger?.markStep("mqtt_answer_received", mapOf("source" to (source ?: "unknown")))

            val isBlocked = parseIsBlocked(json)
            val responseText = parseResponseText(json, message)

            val isValidationError = json.has("message") && json.get("message").asString == "Validation Error"

            val cleanMessage = when {
                isValidationError -> "Maaf, suara tidak terdengar jelas."
                responseText != null && responseText.isNotBlank() -> parseMarkdownToText(responseText).trim().removeSurrounding("\"")
                isBlocked -> "Halo! Saya Sensio, ada yang bisa saya bantu?"
                else -> null
            }

            if (cleanMessage != null) {
                timingLogger?.markStep("final_answer_ready")
                timingLogger?.logTotal("total_e2e_duration", mapOf("outcome" to "success"))
                showResult(currentRequestId, cleanMessage)
            }
        } catch (e: Exception) {
            timingLogger?.markStep("mqtt_parse_error", mapOf("error" to (e.message ?: "unknown")))
            timingLogger?.logTotal("total_e2e_duration", mapOf("outcome" to "parse_error"))
            AppLog.e(TAG, "Error parsing MQTT response", e)
        }
    }

    private fun showResult(requestId: String, text: String) {
        if (activeRequestId != requestId) return
        timeoutJob?.cancel()
        fallbackJob?.cancel()
        timingLogger?.markStep("result_shown")
        timingLogger?.logTotal("total_e2e_duration", mapOf("outcome" to "success"))
        transitionTo(BackgroundAssistantUiState.State.Result, "showResult")
        _uiState.value = _uiState.value.copy(assistantText = text)
        // No auto-dismiss: user dismisses manually via outside tap
    }

    private fun failSession(requestId: String, error: String) {
        if (activeRequestId != requestId) return
        timeoutJob?.cancel()
        fallbackJob?.cancel()
        timingLogger?.markStep("session_failed", mapOf("error" to error))
        timingLogger?.logTotal("total_e2e_duration", mapOf("outcome" to "failed", "error" to error))
        transitionTo(BackgroundAssistantUiState.State.Error, "failSession")
        _uiState.value = _uiState.value.copy(errorText = error)
        // No auto-dismiss: user dismisses manually via outside tap
    }

    /**
     * Shows a friendly service-issue message instead of an error state.
     * Used for recoverable transport/network/server failures.
     */
    private fun showServiceIssue(requestId: String, source: String) {
        if (activeRequestId != requestId) return
        timeoutJob?.cancel()
        fallbackJob?.cancel()
        timingLogger?.markStep("service_issue_shown", mapOf("source" to source))
        timingLogger?.logTotal("total_e2e_duration", mapOf("outcome" to "service_issue", "source" to source))
        transitionTo(BackgroundAssistantUiState.State.Result, "showServiceIssue ($source)")

        // Friendly service-issue message (matches backend ServiceIssue skill)
        val serviceIssueMessage = when (selectedLanguage) {
            "en" -> "Sorry, the AI service or network is having trouble right now. Please try again shortly."
            else -> "Maaf, koneksi atau layanan AI sedang bermasalah. Coba lagi sebentar ya."
        }

        _uiState.value = _uiState.value.copy(assistantText = serviceIssueMessage)
        // No auto-dismiss: user dismisses manually via outside tap
    }

    private fun startProcessingTimeout(requestId: String) {
        timeoutJob?.cancel()
        timeoutJob = scope?.launch {
            delay(PROCESSING_TIMEOUT_MS)
            if (activeRequestId == requestId && _uiState.value.state == BackgroundAssistantUiState.State.Processing) {
                timingLogger?.markStep("processing_timeout_reached", mapOf("threshold_ms" to PROCESSING_TIMEOUT_MS.toString()))
                AppLog.w(TAG, "Processing timeout, attempting HTTP fallback or failing")
                runHttpFallback(requestId)
            }
        }
    }

    private fun runHttpFallback(requestId: String) {
        val file = currentRecordingFile
        if (file == null || !file.exists()) {
            timingLogger?.markStep("http_fallback_failed", mapOf("reason" to "file_not_found"))
            failSession(requestId, "Batas waktu terlampaui")
            return
        }

        fallbackJob?.cancel()
        timingLogger?.markStep("http_fallback_started")
        fallbackJob = scope?.launch {
            try {
                val tm = NetworkModule.tokenManager
                val token = tm.getAccessToken() ?: return@launch failSession(requestId, "Auth error")
                val terminalId = tm.getTerminalId() ?: DeviceUtils.getDeviceId(application)
                val username = DeviceUtils.getDeviceId(application)
                val macAddress = tm.getMacAddress() ?: username
                val fallbackIdempotencyKey = "bg_$requestId"

                timingLogger?.markStep("http_transcribe_initiated")
                NetworkModule.transcribeAudioUseCase.initiate(
                    audioFile = file,
                    token = token,
                    language = selectedLanguage,
                    macAddress = macAddress,
                    idempotencyKey = fallbackIdempotencyKey
                ).collect { result ->
                    if (activeRequestId != requestId) return@collect
                    when (result) {
                        is Resource.Success -> {
                            val taskId = result.data
                            if (taskId != null) {
                                timingLogger?.markStep("http_transcribe_success", mapOf("task_id" to taskId))
                                pollTranscription(taskId, requestId, token, terminalId, username, fallbackIdempotencyKey)
                            } else {
                                timingLogger?.markStep("http_fallback_failed", mapOf("reason" to "null_task_id"))
                                failSession(requestId, "Gagal memproses (Fallback error)")
                            }
                        }
                        is Resource.Error -> {
                            timingLogger?.markStep("http_transcribe_error")
                            showServiceIssue(requestId, "http_transcribe_error")
                        }
                        is Resource.Loading -> {}
                    }
                }
            } catch (e: Exception) {
                timingLogger?.markStep("http_fallback_exception", mapOf("error" to (e.message ?: "unknown")))
                AppLog.e(TAG, "HTTP Fallback failed", e)
                showServiceIssue(requestId, "http_fallback_exception")
            }
        }
    }

    private fun pollTranscription(
        taskId: String,
        requestId: String,
        token: String,
        terminalId: String,
        username: String,
        fallbackIdempotencyKey: String
    ) {
        fallbackJob?.cancel()
        fallbackJob = scope?.launch {
            var isCompleted = false
            var attempts = 0
            val pollStartMs = System.currentTimeMillis()
            timingLogger?.markStep("http_poll_started", mapOf("task_id" to taskId))
            while (!isCompleted && attempts < 10) {
                if (activeRequestId != requestId) return@launch

                var currentSuccessText: String? = null
                var hasError = false

                NetworkModule.transcribeAudioUseCase.getResult(taskId, token).collect { result ->
                    when (result) {
                        is Resource.Success -> {
                            currentSuccessText = result.data
                            isCompleted = true
                        }
                        is Resource.Error -> {
                            hasError = true
                        }
                        is Resource.Loading -> {}
                    }
                }

                if (isCompleted && currentSuccessText != null) {
                    val transcribedText = currentSuccessText!!
                    val pollDurationMs = System.currentTimeMillis() - pollStartMs
                    timingLogger?.markStep("http_transcription_completed", mapOf("poll_duration_ms" to pollDurationMs.toString()))
                    _uiState.value = _uiState.value.copy(recognizedText = transcribedText)

                    timingLogger?.markStep("http_chat_started")
                    NetworkModule.ragRepository.chat(
                        prompt = transcribedText,
                        language = selectedLanguage,
                        terminalId = terminalId,
                        uid = username,
                        token = token,
                        requestId = requestId,
                        idempotencyKey = fallbackIdempotencyKey + "_chat"
                    ).collect { chatResult ->
                        if (activeRequestId != requestId) return@collect
                        when (chatResult) {
                            is Resource.Success -> {
                                val data = chatResult.data
                                if (data != null && data.response != null) {
                                    timingLogger?.markStep("http_chat_success")
                                    timingLogger?.logTotal("total_e2e_duration", mapOf("outcome" to "http_fallback_success"))
                                    showResult(requestId, parseMarkdownToText(data.response!!))
                                } else {
                                    timingLogger?.markStep("http_chat_failed", mapOf("reason" to "invalid_response"))
                                    timingLogger?.logTotal("total_e2e_duration", mapOf("outcome" to "http_fallback_failed", "reason" to "invalid_response"))
                                    failSession(requestId, "Gagal memproses (Invalid chat response)")
                                }
                            }
                            is Resource.Error -> {
                                timingLogger?.markStep("http_chat_error")
                                timingLogger?.logTotal("total_e2e_duration", mapOf("outcome" to "http_fallback_error"))
                                showServiceIssue(requestId, "http_chat_error")
                            }
                            is Resource.Loading -> {}
                        }
                    }
                } else if (hasError) {
                    timingLogger?.markStep("transcription_poll_error")
                    timingLogger?.logTotal("total_e2e_duration", mapOf("outcome" to "transcription_poll_error"))
                    showServiceIssue(requestId, "transcription_poll_error")
                    return@launch
                } else {
                    attempts++
                    delay(1500L)
                }
            }
            if (!isCompleted && activeRequestId == requestId) {
                timingLogger?.markStep("http_poll_timeout", mapOf("attempts" to attempts.toString()))
                timingLogger?.logTotal("total_e2e_duration", mapOf("outcome" to "http_poll_timeout", "attempts" to attempts.toString()))
                showServiceIssue(requestId, "polling_timeout")
            }
        }
    }

    private suspend fun playGreetingAudioAndAwait(sessionId: String) {
        _uiState.value = BackgroundAssistantUiState(
            state = BackgroundAssistantUiState.State.Greeting,
            sessionId = sessionId,
            startedAtMs = System.currentTimeMillis()
        )
        AppLog.d(TAG, "[Session] Greeting (Audio)")

        val completed = suspendCancellableCoroutine<Boolean> { continuation ->
            try {
                val player = MediaPlayer.create(application, R.raw.greeting_sensio_pro_assistant)
                if (player == null) {
                    AppLog.e(TAG, "Failed to create MediaPlayer for greeting")
                    if (continuation.isActive) continuation.resume(false)
                    return@suspendCancellableCoroutine
                }
                greetingPlayer = player
                player.setOnCompletionListener {
                    it.release()
                    if (greetingPlayer == it) greetingPlayer = null
                    if (continuation.isActive) continuation.resume(true)
                }
                player.setOnErrorListener { it, what, extra ->
                    AppLog.e(TAG, "MediaPlayer error: $what, $extra")
                    it.release()
                    if (greetingPlayer == it) greetingPlayer = null
                    if (continuation.isActive) continuation.resume(false)
                    true
                }
                player.start()

                continuation.invokeOnCancellation {
                    try {
                        if (player.isPlaying) player.stop()
                        player.release()
                    } catch (e: Exception) {
                        // Ignore
                    }
                    if (greetingPlayer == player) greetingPlayer = null
                }
            } catch (e: Exception) {
                AppLog.e(TAG, "Error playing greeting audio", e)
                if (continuation.isActive) continuation.resume(false)
            }
        }

        if (!completed) {
            AppLog.w(TAG, "Greeting audio failed or skipped, using fallback delay")
            delay(GREETING_DURATION_MS)
        }
    }

    private fun transitionTo(newState: BackgroundAssistantUiState.State, trigger: String) {
        val oldState = _uiState.value.state
        if (oldState == newState && newState != BackgroundAssistantUiState.State.Hidden) return
        AppLog.d(TAG, "FSM Transition: $oldState -> $newState | Trigger: $trigger")
        _uiState.value = _uiState.value.copy(state = newState)
    }

    private fun cancelPerSessionJobs() {
        activeSessionJob?.cancel()
        listeningJob?.cancel()
        timeoutJob?.cancel()
        mqttCollectorJob?.cancel()
        fallbackJob?.cancel()
        audioRecorder.stop()

        greetingPlayer?.let {
            try {
                if (it.isPlaying) it.stop()
                it.release()
            } catch (e: Exception) {
                AppLog.e(TAG, "Error releasing greeting player", e)
            }
        }
        greetingPlayer = null
    }

    private fun resetState() {
        activeRequestId = null
        requestStartedAtMs.clear()
        _uiState.value = BackgroundAssistantUiState(state = BackgroundAssistantUiState.State.Hidden)
    }

    fun dismissAndRearm() {
        cancelPerSessionJobs()
        resetState()
        onDismissed()
    }

    private suspend fun ensureMqttConnected() {
        val isConnected = try {
            mqttHelper.connectionStatus.value == MqttHelper.MqttConnectionStatus.CONNECTED
        } catch (e: Exception) {
            false
        }

        if (!isConnected) {
            val deviceId = DeviceUtils.getDeviceId(application)
            val pwdResult = NetworkModule.repository.fetchMqttPassword(deviceId)

            if (pwdResult.isSuccess) {
                val password = pwdResult.getOrNull()
                if (password != null) {
                    mqttHelper.connect(password)
                    try {
                        withTimeout(3000L) {
                            mqttHelper.connectionStatus.first { it == MqttHelper.MqttConnectionStatus.CONNECTED }
                        }
                    } catch (e: Exception) {
                        AppLog.e(TAG, "Failed to reconnect MQTT within timeout")
                    }
                }
            }
        }
    }

    private fun parseRequestId(json: com.google.gson.JsonObject): String? {
        return if (json.has("request_id") && !json.get("request_id").isJsonNull) {
            json.get("request_id").asString
        } else if (json.has("data") && !json.get("data").isJsonNull) {
            val data = json.getAsJsonObject("data")
            if (data.has("request_id") && !data.get("request_id").isJsonNull) {
                data.get("request_id").asString
            } else {
                null
            }
        } else {
            null
        }
    }

    private fun parseSource(json: com.google.gson.JsonObject): String? {
        val data = if (json.has("data") && !json.get("data").isJsonNull) json.getAsJsonObject("data") else null
        return if (data != null && data.has("source") && !data.get("source").isJsonNull) data.get("source").asString else null
    }

    private fun parseResponseText(json: com.google.gson.JsonObject, raw: String): String? {
        val data = if (json.has("data") && !json.get("data").isJsonNull) json.getAsJsonObject("data") else null
        return if (data != null && data.has("response") && !data.get("response").isJsonNull) {
            data.get("response").asString
        } else if (json.has("message") && !json.get("message").isJsonNull) {
            json.get("message").asString
        } else if (raw.contains("Response: \"")) {
            raw.substringAfter("Response: \"").substringBeforeLast("\"")
        } else {
            null
        }
    }

    private fun parseIsBlocked(json: com.google.gson.JsonObject): Boolean {
        val data = if (json.has("data") && !json.get("data").isJsonNull) json.getAsJsonObject("data") else null
        return data != null && data.has("is_blocked") && !data.get("is_blocked").isJsonNull && data.get("is_blocked").asBoolean
    }
}
