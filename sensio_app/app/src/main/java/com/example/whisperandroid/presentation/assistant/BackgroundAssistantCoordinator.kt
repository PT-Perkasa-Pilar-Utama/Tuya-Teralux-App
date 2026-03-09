package com.example.whisperandroid.presentation.assistant

import android.app.Application
import android.util.Log
import com.example.whisperandroid.data.di.NetworkModule
import com.example.whisperandroid.presentation.meeting.AudioRecorder
import com.example.whisperandroid.util.MqttHelper
import com.example.whisperandroid.util.parseMarkdownToText
import com.google.gson.JsonParser
import kotlinx.coroutines.*
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.first
import java.io.File
import java.util.UUID

class BackgroundAssistantCoordinator(
    private val application: Application
) {
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
        Log.d("BGAssistantCoord", "Wake detected, starting dummy session")
        runDummySession()
    }

    private fun runDummySession() {
        cancelActiveSession()
        activeSessionJob = scope?.launch {
            val sessionId = UUID.randomUUID().toString()
            val (assistantResp, userPrompt) = dummyResponses.random()

            // 1. Greeting (600ms)
            _uiState.value = BackgroundAssistantUiState(
                state = BackgroundAssistantUiState.State.Greeting,
                sessionId = sessionId,
                startedAtMs = System.currentTimeMillis()
            )
            delay(600L)

            // 2. Listening (2000ms)
            _uiState.value = _uiState.value.copy(
                state = BackgroundAssistantUiState.State.Listening,
                recognizedText = ""
            )
            
            val words = userPrompt.split(" ")
            val listeningStart = System.currentTimeMillis()
            while (System.currentTimeMillis() - listeningStart < 2000L) {
                val elapsed = System.currentTimeMillis() - listeningStart
                val progress = elapsed.toFloat() / 2000f
                val wordCount = (progress * words.size).toInt().coerceIn(0, words.size)
                val currentText = words.take(wordCount).joinToString(" ")
                
                _uiState.value = _uiState.value.copy(
                    recognizedText = currentText,
                    micLevel = (0.2f + Math.random().toFloat() * 0.8f) // Simulated pulse
                )
                delay(100L)
            }
            _uiState.value = _uiState.value.copy(recognizedText = userPrompt, micLevel = 0f)

            // 3. Processing (1500ms)
            _uiState.value = _uiState.value.copy(state = BackgroundAssistantUiState.State.Processing)
            delay(1500L)

            // 4. Result (4000ms)
            _uiState.value = _uiState.value.copy(
                state = BackgroundAssistantUiState.State.Result,
                assistantText = assistantResp
            )
            delay(4000L)

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
            Log.d("BGAssistantCoord", "MQTT disconnected, fetching password for $deviceId")
            val pwdResult = NetworkModule.repository.fetchMqttPassword(deviceId)
            
            if (pwdResult.isSuccess) {
                val password = pwdResult.getOrNull()
                if (password != null) {
                    Log.d("BGAssistantCoord", "Attempting MQTT reconnect before publish")
                    mqttHelper.connect(password)
                    // Wait up to 3 seconds for connection
                    try {
                        withTimeout(3000L) {
                            mqttHelper.connectionStatus.first { it == MqttHelper.MqttConnectionStatus.CONNECTED }
                        }
                    } catch (e: Exception) {
                        Log.e("BGAssistantCoord", "Failed to reconnect MQTT within timeout")
                    }
                }
            } else {
                Log.e("BGAssistantCoord", "Failed to fetch MQTT password: ${pwdResult.exceptionOrNull()?.message}")
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
            } else null
        } else null
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
        } else null
    }

    private fun parseIsBlocked(json: com.google.gson.JsonObject): Boolean {
        val data = if (json.has("data") && !json.get("data").isJsonNull) json.getAsJsonObject("data") else null
        return data != null && data.has("is_blocked") && !data.get("is_blocked").isJsonNull && data.get("is_blocked").asBoolean
    }

    var onDismissedAt: Long = 0L
    var onDismissed: () -> Unit = {}
}
