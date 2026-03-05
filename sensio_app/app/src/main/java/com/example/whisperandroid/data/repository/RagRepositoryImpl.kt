package com.example.whisperandroid.data.repository

import android.util.Log
import com.example.whisperandroid.data.remote.api.RAGApi
import com.example.whisperandroid.data.remote.dto.RAGRequestDto
import com.example.whisperandroid.data.remote.dto.RAGSummaryRequestDto
import com.example.whisperandroid.data.remote.dto.RAGSummaryResponseDto
import com.example.whisperandroid.domain.repository.RagRepository
import com.example.whisperandroid.domain.repository.Resource
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow

class RagRepositoryImpl(
    private val api: RAGApi
) : RagRepository {
    override suspend fun translate(
        text: String,
        targetLang: String,
        macAddress: String?,
        token: String
    ): Flow<Resource<String>> =
        flow {
            emit(Resource.Loading())
            try {
                val response =
                    api.translate(
                        RAGRequestDto(text = text, language = targetLang, macAddress = macAddress),
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

    override suspend fun pollTranslation(
        taskId: String,
        token: String
    ): Flow<Resource<String>> =
        flow {
            emit(Resource.Loading())
            try {
                val response = api.getStatus(taskId, "Bearer $token")
                val statusData = response.data
                val status = statusData?.status?.lowercase()

                Log.d("RagRepo", "Check Translation $taskId: $status")

                when (status) {
                    "completed" -> {
                        val result = statusData.result
                        if (result != null) {
                            emit(Resource.Success(result))
                        } else {
                            emit(Resource.Error("Completed but no translation result found"))
                        }
                    }

                    "failed" -> {
                        emit(Resource.Error(statusData.error ?: "Translation task failed"))
                    }

                    else -> {
                        emit(Resource.Loading())
                    }
                }
            } catch (e: Exception) {
                Log.e("RagRepo", "Check Translation error: ${e.message}")
                emit(Resource.Error("Check error: ${e.message}"))
            }
        }

    override suspend fun generateSummary(
        text: String,
        style: String,
        language: String?,
        context: String?,
        macAddress: String?,
        token: String
    ): Flow<Resource<String>> =
        flow {
            emit(Resource.Loading())
            try {
                val response =
                    api.summary(
                        RAGSummaryRequestDto(
                            text = text,
                            style = style,
                            language = language,
                            context = context,
                            macAddress = macAddress
                        ),
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

    override suspend fun pollSummary(
        taskId: String,
        token: String
    ): Flow<Resource<RAGSummaryResponseDto>> =
        flow {
            emit(Resource.Loading())
            try {
                val response = api.getStatus(taskId, "Bearer $token")
                val statusData = response.data
                val status = statusData?.status?.lowercase()

                Log.d("RagRepo", "Check Summary $taskId: $status")

                when (status) {
                    "completed" -> {
                        val resultJson = statusData.executionResult
                        if (resultJson != null) {
                            try {
                                val summaryResult = com.google.gson.Gson().fromJson(
                                    resultJson,
                                    RAGSummaryResponseDto::class.java
                                )
                                emit(Resource.Success(summaryResult))
                            } catch (e: Exception) {
                                emit(
                                    Resource.Error(
                                        "Failed to parse summary result: ${e.message}"
                                    )
                                )
                            }
                        } else {
                            emit(Resource.Error("Completed but no summary result found"))
                        }
                    }

                    "failed" -> {
                        emit(Resource.Error(statusData.error ?: "Summary task failed"))
                    }

                    else -> {
                        emit(Resource.Loading())
                    }
                }
            } catch (e: Exception) {
                Log.e("RagRepo", "Check Summary error: ${e.message}")
                emit(Resource.Error("Check error: ${e.message}"))
            }
        }
}
