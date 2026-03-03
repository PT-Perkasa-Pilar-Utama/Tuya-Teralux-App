package com.example.whisper_android.domain.usecase

import com.example.whisper_android.domain.repository.RagRepository
import com.example.whisper_android.domain.repository.Resource
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow

class TranslateTextUseCase(
    private val ragRepository: RagRepository
) {
    suspend fun initiate(
        text: String,
        targetLang: String,
        macAddress: String?,
        token: String
    ): Flow<Resource<String>> =
        flow {
            emit(Resource.Loading())
            ragRepository.translate(text, targetLang, macAddress, token).collect { result ->
                emit(result)
            }
        }

    suspend fun getResult(
        taskId: String,
        token: String
    ): Flow<Resource<String>> =
        flow {
            emit(Resource.Loading())
            ragRepository.pollTranslation(taskId, token).collect { result ->
                when (result) {
                    is Resource.Success -> {
                        val translatedText = result.data
                        if (translatedText != null) {
                            emit(Resource.Success(translatedText))
                        } else {
                            emit(Resource.Error("Translation completed but result is null"))
                        }
                    }
                    is Resource.Error -> {
                        emit(Resource.Error(result.message ?: "Translation fetch failed"))
                    }
                    is Resource.Loading -> {
                        emit(Resource.Loading())
                    }
                }
            }
        }
}
