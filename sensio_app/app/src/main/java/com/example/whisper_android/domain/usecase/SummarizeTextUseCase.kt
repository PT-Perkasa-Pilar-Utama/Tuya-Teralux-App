package com.example.whisper_android.domain.usecase

import com.example.whisper_android.data.remote.dto.RAGSummaryResponseDto
import com.example.whisper_android.domain.repository.RagRepository
import com.example.whisper_android.domain.repository.Resource
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow

class SummarizeTextUseCase(
    private val ragRepository: RagRepository
) {
    suspend fun initiate(
        text: String,
        language: String?,
        style: String,
        macAddress: String?,
        token: String
    ): Flow<Resource<String>> =
        flow {
            emit(Resource.Loading())
            ragRepository.generateSummary(text, style, language, null, macAddress, token).collect { result ->
                emit(result)
            }
        }

    suspend fun getResult(
        taskId: String,
        token: String
    ): Flow<Resource<RAGSummaryResponseDto>> =
        flow {
            emit(Resource.Loading())
            ragRepository.pollSummary(taskId, token).collect { result ->
                when (result) {
                    is Resource.Success -> {
                        val summaryData = result.data
                        if (summaryData != null) {
                            emit(Resource.Success(summaryData))
                        } else {
                            emit(Resource.Error("Summary completed but data is null"))
                        }
                    }
                    is Resource.Error -> {
                        emit(Resource.Error(result.message ?: "Summary fetch failed"))
                    }
                    is Resource.Loading -> {
                        emit(Resource.Loading())
                    }
                }
            }
        }
}
