package com.example.whisperandroid.domain.repository

import java.io.File
import kotlinx.coroutines.flow.Flow

sealed class UploadState {
    data class Loading(val message: String? = null) : UploadState()
    data class SessionStarted(val sessionId: String) : UploadState()
    data class Progress(
        val uploadedBytes: Long,
        val totalBytes: Long,
        val percent: Float
    ) : UploadState()
    data class Success(val sessionId: String) : UploadState()
    data class Error(val message: String) : UploadState()
}

interface UploadRepository {
    /**
     * Uploads a file in chunks.
     * Starts by creating a session, then uploads missing chunks.
     * Emits UploadState with progress.
     */
    fun uploadFile(
        file: File,
        token: String,
        chunkSizeMb: Int = 8, // Default 8MB chunks for performance
        sessionId: String? = null
    ): Flow<UploadState>

    /**
     * Gets the status of an existing upload session.
     */
    suspend fun getSessionStatus(
        sessionId: String,
        token: String
    ): Resource<com.example.whisperandroid.data.remote.dto.UploadSessionResponseDto>
}
