package com.example.whisper_android.presentation.assistant

import android.app.Application
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.example.whisper_android.presentation.components.MessageRole
import com.example.whisper_android.presentation.components.TranscriptionMessage
import com.example.whisper_android.util.MqttHelper
import com.example.whisper_android.util.parseMarkdownToText
import kotlinx.coroutines.launch
import java.io.File

class AiAssistantViewModel(application: Application) : AndroidViewModel(application) {
    var transcriptionResults by mutableStateOf(listOf<TranscriptionMessage>())
        private set

    var isRecording by mutableStateOf(false)
        private set

    var isProcessing by mutableStateOf(false)
        private set

    private val mqttHelper = MqttHelper(application)

    init {
        mqttHelper.onMessageReceived = { topic, message ->
            android.util.Log.d("AiAssistantViewModel", "MQTT Message: topic=$topic, message=$message")
            try {
                when {
                    topic.endsWith("chat/answer") -> {
                        val json = org.json.JSONObject(message)
                        val data = json.optJSONObject("data")
                        val responseText = data?.optString("response") ?: json.optString("message")
                        
                        val cleanMessage = parseMarkdownToText(responseText)
                        transcriptionResults = transcriptionResults + TranscriptionMessage(
                            text = cleanMessage,
                            role = MessageRole.ASSISTANT
                        )
                    }
                    topic.endsWith("chat") -> {
                        // Extract prompt if it's JSON, otherwise use raw message
                        val prompt = try {
                            org.json.JSONObject(message).optString("prompt", message)
                        } catch (e: Exception) {
                            message
                        }

                        // Avoid duplicate if we just sent this message
                        val alreadyExists = transcriptionResults.any { 
                            it.role == MessageRole.USER && it.text == prompt 
                        }
                        if (!alreadyExists) {
                            transcriptionResults = transcriptionResults + TranscriptionMessage(
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
        mqttHelper.connect()
    }

    fun sendChat(text: String) {
        if (text.isNotBlank()) {
            transcriptionResults = transcriptionResults + TranscriptionMessage(text, MessageRole.USER)
            viewModelScope.launch {
                mqttHelper.publishChat(text)
            }
        }
    }

    fun startRecording() {
        isRecording = true
    }

    fun stopRecording(audioFile: File) {
        isRecording = false
        isProcessing = true
        viewModelScope.launch {
            if (audioFile.exists()) {
                val bytes = audioFile.readBytes()
                mqttHelper.publishAudio(bytes)
            }
            isProcessing = false
        }
    }

    override fun onCleared() {
        super.onCleared()
        mqttHelper.disconnect()
    }
}
