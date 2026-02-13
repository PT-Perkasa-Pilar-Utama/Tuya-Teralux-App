package com.example.whisper_android.data.repository

import android.util.Log
import com.example.whisper_android.data.remote.api.RAGApi
import com.example.whisper_android.data.remote.dto.RAGRequestDto
import com.example.whisper_android.data.remote.dto.RAGSummaryRequestDto
import com.example.whisper_android.data.remote.dto.RAGSummaryResponseDto
import com.example.whisper_android.domain.repository.RagRepository
import com.example.whisper_android.domain.repository.Resource
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow

class RagRepositoryImpl(
    private val api: RAGApi
) : RagRepository {

    override suspend fun translate(text: String, targetLang: String, token: String): Flow<Resource<String>> = flow {
        emit(Resource.Loading())
        try {
            val response = api.translate(
                RAGRequestDto(text = text, language = targetLang),
                "Bearer $token"
            )
            val taskId = response.data?.taskId
            
            if (response.status && taskId != null) {
                emit(Resource.Success(taskId))
            } else {
                emit(Resource.Error(response.message))
            }
        } catch (e: Exception) {
            emit(Resource.Error("Translate request failed: ${e.message}"))
        }
    }

    override suspend fun pollTranslation(taskId: String, token: String): Flow<Resource<String>> = flow {
        emit(Resource.Loading())
        var attempts = 0
        val maxAttempts = 30 // 1 minute
        
        while (attempts < maxAttempts) {
            try {
                val response = api.getStatus(taskId, "Bearer $token")
                val statusData = response.data
                val status = statusData?.status?.lowercase()

                Log.d("RagRepo", "Polling Translation $taskId: $status")

                when (status) {
                    "completed" -> {
                        // For translation, the result is likely in 'result' string or similar
                        // RAGStatusDto(status, result=String?, executionResult=RAGSummaryResponseDto?)
                        val result = statusData.result
                        if (result != null) {
                            emit(Resource.Success(result))
                            return@flow
                        } else {
                             emit(Resource.Error("Completed but no translation result found"))
                             return@flow
                        }
                    }
                    "failed" -> {
                        emit(Resource.Error("Translation task failed"))
                        return@flow
                    }
                    else -> {
                        delay(2000)
                        attempts++
                    }
                }
            } catch (e: Exception) {
                Log.e("RagRepo", "Polling Translation error (attempt $attempts): ${e.message}")
                delay(2000)
                attempts++
            }
        }
        emit(Resource.Error("Translation timed out"))
    }

    override suspend fun generateSummary(text: String, style: String, context: String?, token: String): Flow<Resource<String>> = flow {
        emit(Resource.Loading())
        try {
            val response = api.summary(
                RAGSummaryRequestDto(text = text, style = style, context = context),
                "Bearer $token"
            )
            val taskId = response.data?.taskId
            
            if (response.status && taskId != null) {
                emit(Resource.Success(taskId))
            } else {
                emit(Resource.Error(response.message))
            }
        } catch (e: Exception) {
            emit(Resource.Error("Summary request failed: ${e.message}"))
        }
    }

    override suspend fun pollSummary(taskId: String, token: String): Flow<Resource<RAGSummaryResponseDto>> = flow {
        emit(Resource.Loading())
        var attempts = 0
        val maxAttempts = 60 // 2 minutes
        
        while (attempts < maxAttempts) {
            try {
                val response = api.getStatus(taskId, "Bearer $token")
                val statusData = response.data
                val status = statusData?.status?.lowercase()

                Log.d("RagRepo", "Polling Summary $taskId: $status")

                when (status) {
                    "completed" -> {
                        val result = statusData.executionResult
                        if (result != null) {
                            emit(Resource.Success<RAGSummaryResponseDto>(result))
                            return@flow
                        } else {
                            emit(Resource.Error("Completed but no summary result found"))
                            return@flow
                        }
                    }
                    "failed" -> {
                        emit(Resource.Error("Summary task failed"))
                        return@flow
                    }
                    else -> {
                        delay(2000)
                        attempts++
                    }
                }
            } catch (e: Exception) {
                Log.e("RagRepo", "Polling Summary error (attempt $attempts): ${e.message}")
                delay(2000)
                attempts++
            }
        }
        emit(Resource.Error("Summary timed out"))
    }
}
