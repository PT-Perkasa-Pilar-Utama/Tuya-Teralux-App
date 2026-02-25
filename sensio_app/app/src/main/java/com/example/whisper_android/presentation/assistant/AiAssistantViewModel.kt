package com.example.whisper_android.presentation.assistant

import android.app.Application
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.example.whisper_android.presentation.components.MessageRole
import com.example.whisper_android.presentation.components.TranscriptionMessage
import com.example.whisper_android.presentation.meeting.AudioRecorder
import com.example.whisper_android.util.MqttHelper
import com.example.whisper_android.util.parseMarkdownToText
import java.io.File
import kotlinx.coroutines.launch

class AiAssistantViewModel(
    application: Application
) : AndroidViewModel(application) {
    var transcriptionResults by mutableStateOf(listOf<TranscriptionMessage>())
        private set

    var isRecording by mutableStateOf(false)
        private set

    var isProcessing by mutableStateOf(false)
        private set

    var selectedLanguage by mutableStateOf("id")
        private set

    var mqttStatus by mutableStateOf(MqttHelper.MqttConnectionStatus.DISCONNECTED)
        private set

    private val mqttHelper = MqttHelper(application)
    private val audioRecorder = AudioRecorder(application)
    private var currentRecordingFile: File? = null

    init {
        mqttHelper.onMessageReceived = { topic, message ->
            viewModelScope.launch {
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
                            val json = org.json.JSONObject(message)
                            val data = json.optJSONObject("data")
                            val responseText = data?.optString("response") ?: json.optString(
                                "message",
                                message
                            )

                            val cleanMessage = parseMarkdownToText(responseText)
                            transcriptionResults = transcriptionResults +
                                TranscriptionMessage(
                                    text = cleanMessage,
                                    role = MessageRole.ASSISTANT
                                )
                            isProcessing = false
                            android.util.Log.d("AiAssistantViewModel", "isProcessing set to false")
                        }

                        topic.endsWith("chat") -> {
                            android.util.Log.d(
                                "AiAssistantViewModel",
                                "Received chat (sync): $message"
                            )
                            // Extract prompt if it's JSON, otherwise use raw message
                            val prompt =
                                try {
                                    val jsonObj = org.json.JSONObject(message)
                                    if (jsonObj.has("prompt")) {
                                        jsonObj.getString("prompt")
                                    } else {
                                        message
                                    }
                                } catch (e: Exception) {
                                    message
                                }

                            // Avoid duplicate if we just sent this message
                            val alreadyExists =
                                transcriptionResults.any {
                                    it.role == MessageRole.USER && it.text == prompt
                                }
                            if (!alreadyExists) {
                                transcriptionResults = transcriptionResults +
                                    TranscriptionMessage(
                                        text = prompt,
                                        role = MessageRole.USER
                                    )
                            }
                        }

                        topic.endsWith("whisper/answer") -> {
                            // Handle whisper answer (e.g. task ID)
                            android.util.Log.d("AiAssistantViewModel", "Whisper Task: $message")
                        }
                    }
                } catch (e: Exception) {
                    android.util.Log.e("AiAssistantViewModel", "Error parsing MQTT message", e)
                }
            }
        }

        mqttHelper.onConnectionStatusChanged = { status ->
            viewModelScope.launch {
                android.util.Log.d("AiAssistantViewModel", "MQTT Status Changed: $status")
                mqttStatus = status
            }
        }

        startConnectionMonitoring()
    }

    private fun startConnectionMonitoring() {
        viewModelScope.launch {
            while (true) {
                if (mqttStatus == MqttHelper.MqttConnectionStatus.DISCONNECTED ||
                    mqttStatus == MqttHelper.MqttConnectionStatus.FAILED
                ) {
                    android.util.Log.d("AiAssistantViewModel", "Attempting MQTT Reconnection...")
                    mqttHelper.connect()
                }
                kotlinx.coroutines.delay(5000) // Retry every 5 seconds
            }
        }
    }

    fun selectLanguage(language: String) {
        selectedLanguage = language
    }

    fun sendChat(text: String) {
        if (text.isNotBlank()) {
            android.util.Log.d("AiAssistantViewModel", "sendChat: $text")
            // Avoid duplicate if we just sent this
            val alreadyExists =
                transcriptionResults.any {
                    it.role == MessageRole.USER && it.text == text
                }
            if (!alreadyExists) {
                transcriptionResults = transcriptionResults + TranscriptionMessage(
                    text,
                    MessageRole.USER
                )
            }
            isProcessing = true
            viewModelScope.launch {
                mqttHelper.publishChat(text, selectedLanguage)
            }
        }
    }

    fun startRecording(file: File) {
        isRecording = true
        currentRecordingFile = file
        audioRecorder.start(file)
    }

    fun stopRecording() {
        if (isRecording) {
            isRecording = false
            isProcessing = true
            audioRecorder.stop()

            val file = currentRecordingFile
            if (file != null && file.exists()) {
                audioRecorder.finalizeWav(file)
                viewModelScope.launch {
                    val bytes = file.readBytes()
                    mqttHelper.publishAudio(bytes, selectedLanguage)
                }
            }
        }
    }

    fun abortProcessing() {
        isProcessing = false
        android.util.Log.d("AiAssistantViewModel", "abortProcessing: isProcessing set to false")
    }

    override fun onCleared() {
        super.onCleared()
        mqttHelper.disconnect()
    }
}
