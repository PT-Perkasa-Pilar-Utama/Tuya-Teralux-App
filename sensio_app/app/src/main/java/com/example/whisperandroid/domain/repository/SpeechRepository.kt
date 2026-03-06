package com.example.whisperandroid.domain.repository

import com.example.whisperandroid.data.remote.dto.SpeechResponseDto
import com.example.whisperandroid.data.remote.dto.TranscriptionResultText
import com.example.whisperandroid.data.remote.dto.TranscriptionStatusDto
import java.io.File
import kotlinx.coroutines.flow.Flow

interface SpeechRepository {
    suspend fun transcribeAudio(
        file: File,
        token: String,
        language: String,
        macAddress: String? = null,
        idempotencyKey: String? = null
    ): Flow<Resource<String>> // Returns Task ID

    suspend fun pollTranscription(
        taskId: String,
        token: String
    ): Flow<Resource<TranscriptionResultText>>

    suspend fun getTranscriptionStatus(
        taskId: String,
        token: String
    ): SpeechResponseDto<TranscriptionStatusDto>

    /**
     * Submits a transcription job using an already uploaded session ID.
     */
    suspend fun transcribeByUpload(
        sessionId: String,
        token: String,
        language: String,
        macAddress: String?,
        idempotencyKey: String?,
        diarize: Boolean = false
    ): Flow<Resource<String>>
}

// Sealed class for resource handling (if not already defined)
sealed class Resource<T>(
    val data: T? = null,
    val message: String? = null
) {
    class Success<T>(
        data: T
    ) : Resource<T>(data)

    class Error<T>(
        message: String,
        data: T? = null
    ) : Resource<T>(data, message)

    class Loading<T>(
        val isLoading: Boolean = true
    ) : Resource<T>(null)
}
