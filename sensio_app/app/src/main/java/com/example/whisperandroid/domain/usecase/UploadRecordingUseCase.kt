package com.example.whisperandroid.domain.usecase

import com.example.whisperandroid.domain.repository.RecordingUploadRepository
import com.example.whisperandroid.domain.repository.RecordingUploadState
import java.io.File
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow

sealed class RecordingProcessState {
    object Idle : RecordingProcessState()
    data class Loading(val message: String) : RecordingProcessState()
    data class Uploading(val progress: Int) : RecordingProcessState()
    data class Finalizing(val objectKey: String) : RecordingProcessState()
    data class Success(val recordingId: String, val filename: String, val audioUrl: String) : RecordingProcessState()
    data class Error(val message: String) : RecordingProcessState()
}

class UploadRecordingUseCase(
    private val repository: RecordingUploadRepository
) {
    operator fun invoke(
        file: File,
        token: String,
        macAddress: String,
        contentType: String = "audio/wav"
    ): Flow<RecordingProcessState> = flow {
            repository.uploadRecording(file, token, macAddress, contentType)
                .collect { state ->
                    when (state) {
                        is RecordingUploadState.Loading -> {
                            emit(RecordingProcessState.Loading(state.message))
                        }
                        is RecordingUploadState.Progress -> {
                            emit(RecordingProcessState.Uploading(state.percent.toInt()))
                        }
                        is RecordingUploadState.Finalizing -> {
                            emit(RecordingProcessState.Finalizing(state.objectKey))
                        }
                        is RecordingUploadState.Success -> {
                            emit(
                                RecordingProcessState.Success(
                                    recordingId = state.recording.id,
                                    filename = state.recording.filename,
                                    audioUrl = state.recording.audioUrl
                                )
                            )
                        }
                        is RecordingUploadState.Error -> {
                            emit(RecordingProcessState.Error(state.message))
                        }
                    }
                }
        }
}