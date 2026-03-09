package com.example.whisperandroid.presentation.assistant

import android.app.Application
import com.example.whisperandroid.data.di.NetworkModule
import com.example.whisperandroid.presentation.meeting.AudioRecorder
import com.example.whisperandroid.util.AppLog
import com.example.whisperandroid.util.MqttHelper
import java.io.File
import java.util.UUID
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch
import kotlinx.coroutines.withTimeout

class BackgroundAssistantCoordinator(
    private val application: Application
) {
    private val TAG = "Coordinator"
    private var scope: CoroutineScope? = null
    private val _uiState = MutableStateFlow(BackgroundAssistantUiState())
    val uiState = _uiState.asStateFlow()

    private val mqttHelper = NetworkModule.mqttHelper
    private val audioRecorder = AudioRecorder(application)
    private var currentRecordingFile: File? = null
    private var activeRequestId: String? = null
    private val requestStartedAtMs = mutableMapOf<String, Long>()

    private var dismissJob: Job? = null
    private var greetingJob: Job? = null
    private var listeningTimeoutJob: Job? = null
    private var responseTimeoutJob: Job? = null
    private var mqttJob: Job? = null

    private var activeSessionJob: Job? = null
    private companion object {
        const val GREETING_DURATION_MS = 1000L
        const val LISTENING_DURATION_MS = 3000L
        const val LISTENING_TICK_MS = 120L
        const val PROCESSING_DURATION_MS = 2200L
        const val RESULT_DURATION_MS = 5000L
    }

    private val dummyResponses = listOf(
        "Tentu, lampu ruang tamu sudah dinyalakan." to "Nyalakan lampu ruang tamu",
        "Suhu AC diatur ke 22 derajat." to "Atur AC ke 22 derajat",
        "Pintu depan sudah terkunci aman." to "Kunci pintu depan",
        "Menampilkan ringkasan rapat terakhir Anda." to "Buka ringkasan rapat",
        "Halo! Ada yang bisa saya bantu hari ini?" to "Halo Sensio"
    )

    fun start(serviceScope: CoroutineScope) {
        this.scope = serviceScope
    }

    fun stop() {
        cancelActiveSession()
        scope = null
        onDismissed = {}
        _uiState.value = BackgroundAssistantUiState(state = BackgroundAssistantUiState.State.Hidden)
    }

    fun onWakeDetected() {
        AppLog.d(TAG, "Wake detected, starting session")
        runDummySession()
    }

    private fun runDummySession() {
        cancelActiveSession()
        activeSessionJob = scope?.launch {
            val sessionId = UUID.randomUUID().toString()
            val (assistantResp, userPrompt) = dummyResponses.random()

            // 1. Greeting
            _uiState.value = BackgroundAssistantUiState(
                state = BackgroundAssistantUiState.State.Greeting,
                sessionId = sessionId,
                startedAtMs = System.currentTimeMillis()
            )
            AppLog.d(TAG, "[Session] Greeting")
            delay(GREETING_DURATION_MS)

            // 2. Listening
            _uiState.value = _uiState.value.copy(
                state = BackgroundAssistantUiState.State.Listening,
                recognizedText = ""
            )
            AppLog.d(TAG, "[Session] Listening")

            val words = userPrompt.split(" ")
            val listeningStart = System.currentTimeMillis()
            while (System.currentTimeMillis() - listeningStart < LISTENING_DURATION_MS) {
                val elapsed = System.currentTimeMillis() - listeningStart
                val progress = elapsed.toFloat() / LISTENING_DURATION_MS.toFloat()
                val wordCount = (progress * words.size).toInt().coerceIn(0, words.size)
                val currentText = words.take(wordCount).joinToString(" ")

                _uiState.value = _uiState.value.copy(
                    recognizedText = currentText,
                    micLevel = (0.2f + Math.random().toFloat() * 0.8f) // Simulated pulse
                )
                delay(LISTENING_TICK_MS)
            }
            _uiState.value = _uiState.value.copy(recognizedText = userPrompt, micLevel = 0f)

            // 3. Processing
            _uiState.value = _uiState.value.copy(state = BackgroundAssistantUiState.State.Processing)
            AppLog.d(TAG, "[Session] Processing")
            delay(PROCESSING_DURATION_MS)

            // 4. Result
            _uiState.value = _uiState.value.copy(
                state = BackgroundAssistantUiState.State.Result,
                assistantText = assistantResp
            )
            AppLog.i(TAG, "[Session] Result shown: $assistantResp")
            delay(RESULT_DURATION_MS)

            // 5. Dismiss
            dismissAndRearm()
        }
    }

    private fun cancelActiveSession() {
        activeSessionJob?.cancel()
        activeSessionJob = null
        resetState()
    }

    fun dismissAndRearm() {
        cancelActiveSession()
        _uiState.value = BackgroundAssistantUiState(state = BackgroundAssistantUiState.State.Hidden)
        onDismissed()
    }

    private fun resetState() {
        activeRequestId = null
        requestStartedAtMs.clear()
        dismissJob?.cancel()
        greetingJob?.cancel()
        listeningTimeoutJob?.cancel()
        responseTimeoutJob?.cancel()
    }

    private suspend fun ensureMqttConnected() {
        val isConnected = try {
            mqttHelper.connectionStatus.value == MqttHelper.MqttConnectionStatus.CONNECTED
        } catch (e: Exception) {
            false
        }

        if (!isConnected) {
            val deviceId = com.example.whisperandroid.util.DeviceUtils.getDeviceId(application)
            AppLog.d(TAG, "MQTT disconnected, fetching password for $deviceId")
            val pwdResult = NetworkModule.repository.fetchMqttPassword(deviceId)

            if (pwdResult.isSuccess) {
                val password = pwdResult.getOrNull()
                if (password != null) {
                    AppLog.d(TAG, "Attempting MQTT reconnect before publish")
                    mqttHelper.connect(password)
                    // Wait up to 3 seconds for connection
                    try {
                        withTimeout(3000L) {
                            mqttHelper.connectionStatus.first { it == MqttHelper.MqttConnectionStatus.CONNECTED }
                        }
                    } catch (e: Exception) {
                        AppLog.e(TAG, "Failed to reconnect MQTT within timeout")
                    }
                }
            } else {
                AppLog.e(TAG, "Failed to fetch MQTT password: ${pwdResult.exceptionOrNull()?.message}")
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

    var onDismissedAt: Long = 0L
    var onDismissed: () -> Unit = {}
}
