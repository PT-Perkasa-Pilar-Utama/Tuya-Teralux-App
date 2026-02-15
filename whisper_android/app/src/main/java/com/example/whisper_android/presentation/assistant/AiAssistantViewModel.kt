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
            if (topic == "users/teralux/chat/answer") {
                val cleanMessage = parseMarkdownToText(message)
                transcriptionResults = transcriptionResults + TranscriptionMessage(
                    text = cleanMessage,
                    role = MessageRole.ASSISTANT
                )
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
