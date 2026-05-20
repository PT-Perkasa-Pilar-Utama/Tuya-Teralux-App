package com.example.whisperandroid.domain.repository

import com.example.whisperandroid.data.remote.dto.RecordingResponseDto
import java.io.File
import kotlinx.coroutines.flow.Flow

sealed class RecordingUploadState {
    data class Loading(val message: String) : RecordingUploadState()
    data class Progress(val uploadedBytes: Long, val totalBytes: Long, val percent: Float) : RecordingUploadState()
    data class Finalizing(val objectKey: String) : RecordingUploadState()
    data class Success(val recording: RecordingResponseDto) : RecordingUploadState()
    data class Error(val message: String) : RecordingUploadState()
}

interface RecordingUploadRepository {
    fun uploadRecording(
        file: File,
        token: String,
        macAddress: String,
        contentType: String
    ): Flow<RecordingUploadState>
}
