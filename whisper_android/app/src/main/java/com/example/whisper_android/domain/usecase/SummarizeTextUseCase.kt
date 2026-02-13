package com.example.whisper_android.domain.usecase

import com.example.whisper_android.data.remote.dto.RAGSummaryResponseDto
import com.example.whisper_android.domain.repository.RagRepository
import com.example.whisper_android.domain.repository.Resource
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow

class SummarizeTextUseCase(
    private val ragRepository: RagRepository
) {
    suspend operator fun invoke(text: String, style: String, token: String): Flow<Resource<RAGSummaryResponseDto>> = flow {
        emit(Resource.Loading())
        
        var taskId: String? = null
        ragRepository.generateSummaryAsync(text, style, null, token).collect { result ->
            when (result) {
                is Resource.Success -> taskId = result.data
                is Resource.Error -> {
                    emit(Resource.Error(result.message ?: "Summary request failed"))
                    return@collect
                }
                is Resource.Loading -> emit(Resource.Loading())
            }
        }
        
        if (taskId == null) return@flow

        // Start Polling
        ragRepository.pollSummary(taskId!!, token).collect { result ->
            when (result) {
                is Resource.Success -> {
                    val summaryData = result.data
                    if (summaryData != null) {
                        emit(Resource.Success(summaryData))
                    } else {
                        emit(Resource.Error("Summary completed but data is null"))
                    }
                }
                is Resource.Error -> emit(Resource.Error(result.message ?: "Summary polling failed"))
                is Resource.Loading -> emit(Resource.Loading())
            }
        }
    }
}
