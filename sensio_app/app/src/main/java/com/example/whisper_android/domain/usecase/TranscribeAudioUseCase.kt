package com.example.whisper_android.domain.usecase

import com.example.whisper_android.domain.repository.Resource
import com.example.whisper_android.domain.repository.SpeechRepository
import java.io.File
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow

class TranscribeAudioUseCase(
    private val speechRepository: SpeechRepository
) {
    suspend fun initiate(
        audioFile: File,
        token: String,
        language: String,
        macAddress: String? = null
    ): Flow<Resource<String>> =
        flow {
            emit(Resource.Loading())
            speechRepository.transcribeAudio(audioFile, token, language, macAddress).collect { result ->
                emit(result)
            }
        }

    suspend fun getResult(
        taskId: String,
        token: String
    ): Flow<Resource<String>> =
        flow {
            emit(Resource.Loading())
            speechRepository.pollTranscription(taskId, token).collect { result ->
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
                        emit(Resource.Error(result.message ?: "Transcription fetch failed"))
                    }
                    is Resource.Loading -> {
                        emit(Resource.Loading())
                    }
                }
            }
        }
}
