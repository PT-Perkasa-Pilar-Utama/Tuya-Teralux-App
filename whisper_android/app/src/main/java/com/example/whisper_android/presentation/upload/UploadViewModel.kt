package com.example.whisper_android.presentation.upload

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.whisper_android.data.repository.SpeechRepository
import com.example.whisper_android.presentation.components.MessageRole
import com.example.whisper_android.presentation.components.TranscriptionMessage
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import java.io.File

data class UploadUiState(
    val isRecording: Boolean = false,
    val isProcessing: Boolean = false,
    val transcriptionResults: List<TranscriptionMessage> = emptyList(),
    val error: String? = null
)

class UploadViewModel(
    private val repository: SpeechRepository,
    private val recorder: AudioRecorder,
    private val tokenManager: com.example.whisper_android.data.local.TokenManager,
    private val cacheDir: File
) : ViewModel() {

    private val _uiState = MutableStateFlow(UploadUiState())
    val uiState: StateFlow<UploadUiState> = _uiState.asStateFlow()

    private var currentRecordingFile: File? = null
    private var pollingJob: Job? = null

    fun toggleRecording() {
        if (_uiState.value.isRecording) {
            stopRecording()
        } else {
            startRecording()
        }
    }

    private fun startRecording() {
        viewModelScope.launch {
            try {
                val file = File(cacheDir, "upload_recording_${System.currentTimeMillis()}.m4a")
                currentRecordingFile = file
                recorder.start(file)
                _uiState.update { it.copy(isRecording = true, error = null) }
            } catch (e: Exception) {
                _uiState.update { it.copy(error = "Failed to start recording: ${e.message}") }
            }
        }
    }

    private fun stopRecording() {
        viewModelScope.launch {
            try {
                recorder.stop()
                _uiState.update { it.copy(isRecording = false, isProcessing = true) }
                uploadAndPoll()
            } catch (e: Exception) {
                _uiState.update { it.copy(isRecording = false, error = "Failed to stop recording: ${e.message}") }
            }
        }
    }

    private fun uploadAndPoll() {
        val file = currentRecordingFile ?: return
        val token = tokenManager.getAccessToken()
        if (token == null) {
            _uiState.update { it.copy(isProcessing = false, error = "Authentication token missing. Please login again.") }
            return
        }

        viewModelScope.launch {
            repository.transcribeAudio(file, token).onSuccess { data ->
                startPolling(data.taskId)
            }.onFailure { e ->
                _uiState.update { it.copy(isProcessing = false, error = "Upload failed: ${e.message}") }
                file.delete()
            }
        }
    }

    private fun startPolling(taskId: String) {
        val token = tokenManager.getAccessToken() ?: return
        pollingJob?.cancel()
        pollingJob = viewModelScope.launch {
            var completed = false
            var attempts = 0
            val maxAttempts = 60 // 2 minutes with 2s delay

            while (!completed && attempts < maxAttempts) {
                delay(2000)
                attempts++

                repository.getStatus(taskId, token).onSuccess { data ->
                    val statusStr = data.taskStatus?.status?.lowercase() ?: ""
                    android.util.Log.d("UploadViewModel", "Task $taskId Status: $statusStr")
                    
                    when (statusStr) {
                        "completed" -> {
                            completed = true
                            val result = data.taskStatus?.result
                            val transcription = result?.transcription ?: ""
                            val translatedText = result?.translatedText
                            val detectedLang = result?.detectedLanguage ?: "unknown"
                            
                            val displayText = if (!translatedText.isNullOrEmpty()) translatedText else transcription
                            val finalText = if (displayText.isNotEmpty()) displayText else "No text transcribed"

                            android.util.Log.d("UploadViewModel", "Transcription: $transcription, Translated: $translatedText, Lang: $detectedLang")
                            
                            _uiState.update { state ->
                                val newResults = state.transcriptionResults.toMutableList()
                                newResults.add(TranscriptionMessage(finalText, MessageRole.USER)) 
                                // Add dummy agent response as requested
                                newResults.add(TranscriptionMessage("Agent: I've processed your audio. Is there anything else?", MessageRole.ASSISTANT))
                                state.copy(
                                    isProcessing = false,
                                    transcriptionResults = newResults
                                )
                            }
                            currentRecordingFile?.delete()
                        }
                        "failed" -> {
                            completed = true
                            _uiState.update { it.copy(isProcessing = false, error = "Transcription failed on server") }
                            currentRecordingFile?.delete()
                        }
                        else -> {
                            // Still processing (e.g. "pending", "processing")
                        }
                    }
                }.onFailure { e ->
                    // Optionally handle polling error, but keep trying
                    if (attempts >= maxAttempts) {
                        _uiState.update { it.copy(isProcessing = false, error = "Polling timed out: ${e.message}") }
                    }
                }
            }
        }
    }

    fun clearLog() {
        _uiState.update { it.copy(transcriptionResults = emptyList()) }
    }

    override fun onCleared() {
        super.onCleared()
        recorder.stop()
        pollingJob?.cancel()
    }
}
