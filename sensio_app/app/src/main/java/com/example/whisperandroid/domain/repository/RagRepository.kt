package com.example.whisperandroid.domain.repository

import com.example.whisperandroid.data.remote.dto.RAGSummaryResponseDto
import kotlinx.coroutines.flow.Flow

interface RagRepository {
    suspend fun translate(text: String, targetLang: String, macAddress: String?, token: String): Flow<Resource<String>> // Returns Task ID

    suspend fun pollTranslation(taskId: String, token: String): Flow<Resource<String>> // Returns Translated Text

    suspend fun generateSummary(
        text: String,
        style: String,
        language: String?,
        context: String?,
        macAddress: String?,
        token: String
    ): Flow<Resource<String>> // Returns Task ID

    suspend fun pollSummary(
        taskId: String,
        token: String
    ): Flow<Resource<RAGSummaryResponseDto>> // Returns Summary & PDF URL
}
