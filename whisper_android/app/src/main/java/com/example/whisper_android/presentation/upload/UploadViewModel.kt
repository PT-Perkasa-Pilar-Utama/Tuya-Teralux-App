package com.example.whisper_android.presentation.upload

import android.content.Context
import android.net.Uri
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.whisper_android.data.repository.SpeechRepository
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import java.io.File
import java.io.FileOutputStream

data class UploadUiState(
    val isLoading: Boolean = false,
    val isRecording: Boolean = false,
    val isPaused: Boolean = false,
    val isThinking: Boolean = false,
    val transcription: String = "",
    val refinedText: String = "",
    val summary: String = "",
    val displaySummary: String = "", // For typing effect
    val pdfUrl: String? = null,
    val summaryLanguage: String = "id",
    val error: String? = null,
    val availableFiles: List<File> = emptyList(),
    val showInternalPicker: Boolean = false
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

    fun handleMicClick() {
        val state = _uiState.value
        when {
            !state.isRecording -> startRecording()
            state.isPaused -> resumeRecording()
            else -> pauseRecording()
        }
    }

    fun handleMicStop() {
        if (_uiState.value.isRecording) {
            stopRecording()
        }
    }

    fun setSummaryLanguage(lang: String) {
        _uiState.update { it.copy(summaryLanguage = lang) }
    }

    fun handleFileSelected(uri: Uri, context: Context) {
        viewModelScope.launch {
            _uiState.update { it.copy(isThinking = true, error = null) }
            val file = copyUriToCache(uri, context)
            if (file != null) {
                currentRecordingFile = file
                uploadAndPoll()
            } else {
                _uiState.update { it.copy(isThinking = false, error = "Failed to access selected file") }
            }
        }
    }

    fun handleFileSelected(file: File) {
        viewModelScope.launch {
            _uiState.update { it.copy(isThinking = true, showInternalPicker = false, error = null) }
            currentRecordingFile = file
            uploadAndPoll()
        }
    }

    fun scanDownloadsFolder() {
        viewModelScope.launch(Dispatchers.IO) {
            _uiState.update { it.copy(isLoading = true) }
            val downloadsDir = android.os.Environment.getExternalStoragePublicDirectory(
                android.os.Environment.DIRECTORY_DOWNLOADS
            )
            val files = downloadsDir.listFiles { file ->
                val name = file.name.lowercase()
                file.isFile && (name.endsWith(".mp3") || name.endsWith(".m4a") || name.endsWith(".wav"))
            }?.toList() ?: emptyList()

            _uiState.update { it.copy(
                availableFiles = files.sortedByDescending { f -> f.lastModified() },
                showInternalPicker = true,
                isLoading = false
            ) }
        }
    }

    fun hideInternalPicker() {
        _uiState.update { it.copy(showInternalPicker = false) }
    }

    private suspend fun copyUriToCache(uri: Uri, context: Context): File? = withContext(Dispatchers.IO) {
        try {
            val extension = context.contentResolver.getType(uri)?.split("/")?.lastOrNull() ?: "m4a"
            val file = File(cacheDir, "upload_file_${System.currentTimeMillis()}.$extension")
            val copied = context.contentResolver.openInputStream(uri)?.use { input ->
                FileOutputStream(file).use { output ->
                    input.copyTo(output)
                    true
                }
            } ?: false
            
            if (copied) file else null
        } catch (e: Exception) {
            e.printStackTrace()
            null
        }
    }

    private fun startRecording() {
        viewModelScope.launch {
            try {
                val file = File(cacheDir, "upload_recording_${System.currentTimeMillis()}.m4a")
                currentRecordingFile = file
                recorder.start(file)
                _uiState.update { it.copy(isRecording = true, isPaused = false, error = null) }
            } catch (e: Exception) {
                _uiState.update { it.copy(error = "Failed to start recording: ${e.message}") }
            }
        }
    }

    private fun pauseRecording() {
        try {
            recorder.pause()
            _uiState.update { it.copy(isPaused = true) }
        } catch (e: Exception) {
            _uiState.update { it.copy(error = "Failed to pause: ${e.message}") }
        }
    }

    private fun resumeRecording() {
        try {
            recorder.resume()
            _uiState.update { it.copy(isPaused = false) }
        } catch (e: Exception) {
            _uiState.update { it.copy(error = "Failed to resume: ${e.message}") }
        }
    }

    private fun stopRecording() {
        viewModelScope.launch {
            try {
                recorder.stop()
                _uiState.update { it.copy(isRecording = false, isPaused = false, isThinking = true) }
                uploadAndPoll()
            } catch (e: Exception) {
                _uiState.update { it.copy(isRecording = false, isPaused = false, error = "Failed to stop recording: ${e.message}") }
            }
        }
    }

    private fun uploadAndPoll() {
        val file = currentRecordingFile ?: return
        val token = tokenManager.getAccessToken()
        if (token == null) {
            _uiState.update { it.copy(isThinking = false, error = "Authentication token missing. Please login again.") }
            return
        }

        viewModelScope.launch {
            repository.transcribeAudio(file, token).onSuccess { data ->
                pollTranscription(data.taskId)
            }.onFailure { e ->
                _uiState.update { it.copy(isThinking = false, error = "Upload failed: ${e.message}") }
                file.delete()
            }
        }
    }

    private fun pollTranscription(taskId: String) {
        val token = tokenManager.getAccessToken() ?: return
        pollingJob?.cancel()
        pollingJob = viewModelScope.launch {
            var completed = false
            var attempts = 0
            val maxAttempts = 60 // 2 minutes

            while (!completed && attempts < maxAttempts) {
                delay(2000)
                attempts++

                repository.getStatus(taskId, token).onSuccess { data ->
                    val statusStr = data.taskStatus?.status?.lowercase() ?: ""
                    if (statusStr == "completed") {
                        completed = true
                        val result = data.taskStatus?.result
                        val transcription = result?.transcription ?: ""
                        val refinedText = result?.refinedText ?: transcription
                        val detectedLang = result?.detectedLanguage ?: "unknown"

                        _uiState.update { it.copy(
                            transcription = transcription,
                            refinedText = refinedText
                        ) }

                        startTranslationPipeline(refinedText, detectedLang)
                    } else if (statusStr == "failed") {
                        completed = true
                        _uiState.update { it.copy(isThinking = false, error = "Transcription failed on server") }
                    }
                }.onFailure { e ->
                    if (attempts >= maxAttempts) {
                        _uiState.update { it.copy(isThinking = false, error = "Transcription polling timed out: ${e.message}") }
                    }
                }
            }
            currentRecordingFile?.delete()
        }
    }

    private fun startTranslationPipeline(text: String, detectedLang: String) {
        val token = tokenManager.getAccessToken() ?: return
        val targetLang = _uiState.value.summaryLanguage

        viewModelScope.launch {
            repository.translateAsync(text, targetLang, token).onSuccess { data ->
                pollTranslation(data.taskId, text) // Pass original text as fallback
            }.onFailure { e ->
                android.util.Log.e("UploadViewModel", "Translation submission failed: ${e.message}")
                // Fallback to summary with original text
                startSummaryPipeline(text)
            }
        }
    }

    private fun pollTranslation(taskId: String, originalText: String) {
        val token = tokenManager.getAccessToken() ?: return
        pollingJob?.cancel()
        pollingJob = viewModelScope.launch {
            var completed = false
            var attempts = 0
            val maxAttempts = 30 // 1 minute

            while (!completed && attempts < maxAttempts) {
                delay(2000)
                attempts++

                repository.getRagStatus(taskId, token).onSuccess { data ->
                    if (data.status == "done") {
                        completed = true
                        startSummaryPipeline(data.result ?: originalText)
                    } else if (data.status == "error") {
                        completed = true
                        startSummaryPipeline(originalText)
                    }
                }.onFailure { e ->
                    if (attempts >= maxAttempts) {
                        startSummaryPipeline(originalText)
                    }
                }
            }
        }
    }

    private fun startSummaryPipeline(text: String) {
        val token = tokenManager.getAccessToken() ?: return
        val targetLang = _uiState.value.summaryLanguage
        val request = com.example.whisper_android.data.remote.dto.RAGSummaryRequestDto(
            text = text,
            language = targetLang
        )

        viewModelScope.launch {
            repository.summaryAsync(request, token).onSuccess { data ->
                pollSummary(data.taskId)
            }.onFailure { e ->
                _uiState.update { it.copy(isThinking = false, error = "Summary submission failed: ${e.message}") }
            }
        }
    }

    private fun pollSummary(taskId: String) {
        val token = tokenManager.getAccessToken() ?: return
        pollingJob?.cancel()
        pollingJob = viewModelScope.launch {
            var completed = false
            var attempts = 0
            val maxAttempts = 60 // 2 minutes

            while (!completed && attempts < maxAttempts) {
                delay(2000)
                attempts++

                repository.getRagStatus(taskId, token).onSuccess { data ->
                    if (data.status == "done") {
                        completed = true
                        val result = data.executionResult
                        if (result != null) {
                            _uiState.update { it.copy(
                                summary = result.summary,
                                pdfUrl = result.pdfUrl,
                                isThinking = false
                            ) }
                            startTypingEffect(result.summary)
                        } else {
                            _uiState.update { it.copy(isThinking = false, error = "Summary result is empty") }
                        }
                    } else if (data.status == "error") {
                        completed = true
                        _uiState.update { it.copy(isThinking = false, error = "Summary generation failed: ${data.result}") }
                    }
                }.onFailure { e ->
                    if (attempts >= maxAttempts) {
                        _uiState.update { it.copy(isThinking = false, error = "Summary polling timed out: ${e.message}") }
                    }
                }
            }
        }
    }

    private fun startTypingEffect(text: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(displaySummary = "") }
            val sb = StringBuilder()
            for (char in text) {
                sb.append(char)
                _uiState.update { it.copy(displaySummary = sb.toString()) }
                delay(30)
            }
        }
    }

    fun clearLog() {
        _uiState.update { it.copy(
            transcription = "",
            refinedText = "",
            summary = "",
            displaySummary = ""
        ) }
    }

    override fun onCleared() {
        super.onCleared()
        recorder.stop()
        pollingJob?.cancel()
    }
}
