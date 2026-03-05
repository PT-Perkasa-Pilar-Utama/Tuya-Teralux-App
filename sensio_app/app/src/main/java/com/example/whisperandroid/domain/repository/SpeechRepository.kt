package com.example.whisperandroid.domain.repository

import com.example.whisperandroid.data.remote.dto.TranscriptionResultText
import java.io.File
import kotlinx.coroutines.flow.Flow

interface SpeechRepository {
    suspend fun transcribeAudio(
        file: File,
        token: String,
        language: String,
        macAddress: String? = null
    ): Flow<Resource<String>> // Returns Task ID

    suspend fun pollTranscription(
        taskId: String,
        token: String
    ): Flow<Resource<TranscriptionResultText>>
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
