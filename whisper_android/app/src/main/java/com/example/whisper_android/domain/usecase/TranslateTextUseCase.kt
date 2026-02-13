package com.example.whisper_android.domain.usecase

import com.example.whisper_android.domain.repository.RagRepository
import com.example.whisper_android.domain.repository.Resource
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow

class TranslateTextUseCase(
    private val ragRepository: RagRepository
) {
    suspend operator fun invoke(text: String, targetLang: String, token: String): Flow<Resource<String>> = flow {
        emit(Resource.Loading())
        
        var taskId: String? = null
        ragRepository.translate(text, targetLang, token).collect { result ->
            when (result) {
                is Resource.Success -> taskId = result.data
                is Resource.Error -> {
                    emit(Resource.Error(result.message ?: "Translation request failed"))
                    return@collect
                }
                is Resource.Loading -> emit(Resource.Loading())
            }
        }
        
        if (taskId == null) return@flow

        // Start Polling
        ragRepository.pollTranslation(taskId!!, token).collect { result ->
            when (result) {
                is Resource.Success -> {
                    val translatedText = result.data
                    if (translatedText != null) {
                        emit(Resource.Success(translatedText))
                    } else {
                        emit(Resource.Error("Translation completed but result is null"))
                    }
                }
                is Resource.Error -> emit(Resource.Error(result.message ?: "Translation polling failed"))
                is Resource.Loading -> emit(Resource.Loading())
            }
        }
    }
}
