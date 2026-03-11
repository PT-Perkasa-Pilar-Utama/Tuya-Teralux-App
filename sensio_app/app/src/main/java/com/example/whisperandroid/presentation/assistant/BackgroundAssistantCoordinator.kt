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
    private var dismissTimerJob: Job? = null
    private var greetingPlayer: MediaPlayer? = null

    var onDismissed: () -> Unit = {}

    val isSessionActive: Boolean
        get() = _uiState.value.state != BackgroundAssistantUiState.State.Hidden

    private companion object {
        const val GREETING_DURATION_MS = 1000L
        const val LISTENING_TICK_MS = 100L
        const val LISTENING_TIMEOUT_MS = 10000L
        const val PROCESSING_TIMEOUT_MS = 12000L
        const val RESULT_DURATION_MS = 6000L
        const val ERROR_DURATION_MS = 3500L
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

            // 1. Greeting
            playGreetingAudioAndAwait()

            // 2. Listening
            startRecording(requestId)
        }
    }

    private fun startRecording(requestId: String) {
        val cacheDir = application.cacheDir
        val file = File(cacheDir, "bg_assistant_${System.currentTimeMillis()}.wav")
        currentRecordingFile = file

        transitionTo(BackgroundAssistantUiState.State.Listening, "startRecording")

        val result = audioRecorder.start(file)
        if (result !is AudioRecorder.RecorderResult.Success) {
            AppLog.e(TAG, "Mic initialization failed: $result")
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
                stopRecordingAndPublish(requestId)
            }
        }

        // Dedicated response collector (MQTT)
        startMqttCollector(requestId)
    }

    private fun stopRecordingAndPublish(requestId: String) {
        listeningJob?.cancel()
        audioRecorder.stop()

        val file = currentRecordingFile
        if (file != null && audioRecorder.isWavValid(file)) {
            audioRecorder.finalizeWav(file)
            transitionTo(BackgroundAssistantUiState.State.Processing, "stopRecording(valid)")

            scope?.launch {
                ensureMqttConnected()
                val bytes = file.readBytes()
                val publishResult = mqttHelper.publishAudio(bytes, selectedLanguage, requestId)
                if (publishResult.isSuccess) {
                    AppLog.d(TAG, "Audio published successfully, waiting for response")
                    startProcessingTimeout(requestId)
                } else {
                    AppLog.e(TAG, "Failed to publish audio via MQTT")
                    failSession(requestId, "Gagal mengirim perintah")
                }
            }
        } else {
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
                val prompt = if (json.has("prompt")) json.get("prompt").asString else message
                if (responseRequestId == currentRequestId) {
                    _uiState.value = _uiState.value.copy(recognizedText = prompt.trim().removeSurrounding("\""))
                }
                return
            }

            // chat/answer handling
            if (responseRequestId == null || responseRequestId != currentRequestId) {
                AppLog.d(TAG, "Ignored response for foreign/stale RID: $responseRequestId (Current: $currentRequestId)")
                return
            }

            val source = parseSource(json)
            if (source == "MQTT_SYNC_DROP" || source == "MQTT_DUP_DROP") {
                AppLog.d(TAG, "Ignored ack-only source: $source")
                return
            }

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
                showResult(currentRequestId, cleanMessage)
            }
        } catch (e: Exception) {
            AppLog.e(TAG, "Error parsing MQTT response", e)
        }
    }

    private fun showResult(requestId: String, text: String) {
        if (activeRequestId != requestId) return
        timeoutJob?.cancel()
        fallbackJob?.cancel()
        transitionTo(BackgroundAssistantUiState.State.Result, "showResult")
        _uiState.value = _uiState.value.copy(assistantText = text)

        dismissTimerJob?.cancel()
        dismissTimerJob = scope?.launch {
            delay(RESULT_DURATION_MS)
            dismissAndRearm()
        }
    }

    private fun failSession(requestId: String, error: String) {
        if (activeRequestId != requestId) return
        timeoutJob?.cancel()
        fallbackJob?.cancel()
        transitionTo(BackgroundAssistantUiState.State.Error, "failSession")
        _uiState.value = _uiState.value.copy(errorText = error)

        dismissTimerJob?.cancel()
        dismissTimerJob = scope?.launch {
            delay(ERROR_DURATION_MS)
            dismissAndRearm()
        }
    }

    private fun startProcessingTimeout(requestId: String) {
        timeoutJob?.cancel()
        timeoutJob = scope?.launch {
            delay(PROCESSING_TIMEOUT_MS)
            if (activeRequestId == requestId && _uiState.value.state == BackgroundAssistantUiState.State.Processing) {
                AppLog.w(TAG, "Processing timeout, attempting HTTP fallback or failing")
                runHttpFallback(requestId)
            }
        }
    }

    private fun runHttpFallback(requestId: String) {
        val file = currentRecordingFile
        if (file == null || !file.exists()) {
            failSession(requestId, "Batas waktu terlampaui")
            return
        }

        fallbackJob?.cancel()
        fallbackJob = scope?.launch {
            try {
                val tm = NetworkModule.tokenManager
                val token = tm.getAccessToken() ?: return@launch failSession(requestId, "Auth error")
                val terminalId = tm.getTerminalId() ?: DeviceUtils.getDeviceId(application)
                val username = DeviceUtils.getDeviceId(application)
                val fallbackIdempotencyKey = "bg_$requestId"

                NetworkModule.transcribeAudioUseCase.initiate(
                    audioFile = file,
                    token = token,
                    language = selectedLanguage,
                    macAddress = terminalId,
                    idempotencyKey = fallbackIdempotencyKey
                ).collect { result ->
                    if (activeRequestId != requestId) return@collect
                    when (result) {
                        is Resource.Success -> {
                            val taskId = result.data
                            if (taskId != null) {
                                pollTranscription(taskId, requestId, token, terminalId, username, fallbackIdempotencyKey)
                            } else {
                                failSession(requestId, "Gagal memproses (Fallback error)")
                            }
                        }
                        is Resource.Error -> {
                            failSession(requestId, "Gagal memproses (Transcribe API error)")
                        }
                        is Resource.Loading -> {}
                    }
                }
            } catch (e: Exception) {
                AppLog.e(TAG, "HTTP Fallback failed", e)
                failSession(requestId, "Koneksi bermasalah")
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
                    _uiState.value = _uiState.value.copy(recognizedText = transcribedText)

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
                                    showResult(requestId, parseMarkdownToText(data.response!!))
                                } else {
                                    failSession(requestId, "Gagal memproses (Invalid chat response)")
                                }
                            }
                            is Resource.Error -> {
                                failSession(requestId, "Gagal memproses (Chat API error)")
                            }
                            is Resource.Loading -> {}
                        }
                    }
                } else if (hasError) {
                    failSession(requestId, "Gagal mengenali suara")
                    return@launch
                } else {
                    attempts++
                    delay(1500L)
                }
            }
            if (!isCompleted && activeRequestId == requestId) {
                failSession(requestId, "Batas waktu polling terlampaui")
            }
        }
    }

    private suspend fun playGreetingAudioAndAwait() {
        val sessionId = UUID.randomUUID().toString()
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
        dismissTimerJob?.cancel()
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
