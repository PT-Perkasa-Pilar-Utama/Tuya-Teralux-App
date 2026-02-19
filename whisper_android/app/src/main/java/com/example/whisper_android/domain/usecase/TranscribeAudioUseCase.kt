package com.example.whisper_android.domain.usecase

import com.example.whisper_android.domain.repository.Resource
import com.example.whisper_android.domain.repository.SpeechRepository
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow
import java.io.File

class TranscribeAudioUseCase(
    private val speechRepository: SpeechRepository,
) {
    suspend operator fun invoke(
        audioFile: File,
        token: String,
        language: String,
    ): Flow<Resource<String>> =
        flow {
            emit(Resource.Loading())

            var taskId: String? = null
            speechRepository.transcribeAudio(audioFile, token, language).collect { result ->
                when (result) {
                    is Resource.Success -> {
                        taskId = result.data
                    }

                    is Resource.Error -> {
                        emit(Resource.Error(result.message ?: "Transcription request failed"))
                        return@collect
                    }

                    is Resource.Loading -> {
                        emit(Resource.Loading())
                    }
                }
            }

            if (taskId == null) return@flow

            // Start Polling
            speechRepository.pollTranscription(taskId!!, token).collect { result ->
                when (result) {
                    is Resource.Success -> {
                        val text = result.data?.transcription
                        if (text != null) {
                            emit(Resource.Success(text))
                        } else {
                            emit(Resource.Error("Transcription completed but text is null"))
                        }
                    }

                    is Resource.Error -> {
                        emit(Resource.Error(result.message ?: "Transcription polling failed"))
                    }

                    is Resource.Loading -> {
                        emit(Resource.Loading())
                    }
                }
            }
        }
}
