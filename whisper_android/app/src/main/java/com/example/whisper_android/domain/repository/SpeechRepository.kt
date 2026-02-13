package com.example.whisper_android.domain.repository

import com.example.whisper_android.data.remote.dto.SpeechResponseDto
import com.example.whisper_android.data.remote.dto.TranscriptionResultText
import com.example.whisper_android.data.remote.dto.TranscriptionSubmissionData
import java.io.File
import kotlinx.coroutines.flow.Flow

interface SpeechRepository {
    suspend fun transcribeAudio(file: File, token: String, language: String): Flow<Resource<String>> // Returns Task ID
    suspend fun pollTranscription(taskId: String, token: String): Flow<Resource<TranscriptionResultText>>
}

// Sealed class for resource handling (if not already defined)
sealed class Resource<T>(val data: T? = null, val message: String? = null) {
    class Success<T>(data: T) : Resource<T>(data)
    class Error<T>(message: String, data: T? = null) : Resource<T>(data, message)
    class Loading<T>(val isLoading: Boolean = true) : Resource<T>(null)
}
