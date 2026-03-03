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

    private val mqttHelper = com.example.whisper_android.data.di.NetworkModule.mqttHelper
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
                            val json = com.google.gson.JsonParser.parseString(message).asJsonObject
                            val data = if (json.has("data") && !json.get("data").isJsonNull) json.getAsJsonObject("data") else null
                            val responseText = if (data != null && data.has("response") && !data.get("response").isJsonNull) {
                                data.get("response").asString
                            } else if (json.has("message") && !json.get("message").isJsonNull) {
                                json.get("message").asString
                            } else {
                                message
                            }

                            val isValidationError = json.has("message") && !json.get("message").isJsonNull && json.get("message").asString == "Validation Error"
                            
                            val cleanMessage = if (isValidationError) {
                                "Maaf, suara tidak terdengar dengan jelas. Silakan coba lagi."
                            } else {
                                parseMarkdownToText(responseText).trim().removeSurrounding("\"")
                            }

                            val lastRole = transcriptionResults.lastOrNull()?.role
                            val isDuplicateAnswer = !isProcessing && lastRole == MessageRole.ASSISTANT
                            
                            if (!isDuplicateAnswer) {
                                transcriptionResults = transcriptionResults +
                                        TranscriptionMessage(
                                            text = cleanMessage,
                                            role = MessageRole.ASSISTANT
                                        )
                            }
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
                                    val jsonObj = com.google.gson.JsonParser.parseString(message).asJsonObject
                                    if (jsonObj.has("prompt") && !jsonObj.get("prompt").isJsonNull) {
                                        jsonObj.get("prompt").asString
                                    } else {
                                        message
                                    }
                                } catch (e: Exception) {
                                    message
                                }

                            val cleanPrompt = prompt.trim().removeSurrounding("\"")

                            if (cleanPrompt.isBlank()) {
                                return@collect
                            }

                            // Avoid duplicate if we just sent this message
                            val alreadyExists =
                                transcriptionResults.any {
                                    it.role == MessageRole.USER && it.text == cleanPrompt
                                }
                            
                            val lastRole = transcriptionResults.lastOrNull()?.role
                            val isDuplicateTranscription = isProcessing && lastRole == MessageRole.USER

                            if (!alreadyExists && !isDuplicateTranscription) {
                                transcriptionResults = transcriptionResults +
                                    TranscriptionMessage(
                                        text = cleanPrompt,
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
        
        // Auto-connect when the ViewModel is initialized
        reconnectMqtt()
    }

    fun reconnectMqtt() {
        viewModelScope.launch {
            android.util.Log.d("AiAssistantViewModel", "Manual MQTT Reconnection...")
            val username = com.example.whisper_android.util.DeviceUtils.getDeviceId(getApplication())
            val pwdResult = com.example.whisper_android.data.di.NetworkModule.repository.fetchMqttPassword(username)
            if (pwdResult.isSuccess) {
                mqttHelper.connect(pwdResult.getOrNull()!!)
            } else {
                android.util.Log.e("AiAssistantViewModel", "Failed to fetch MQTT password: ${pwdResult.exceptionOrNull()?.message}")
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
