package com.example.whisper_android.domain.repository

import com.example.whisper_android.data.remote.dto.RAGSummaryResponseDto
import kotlinx.coroutines.flow.Flow

interface RagRepository {
    suspend fun translate(text: String, targetLang: String, token: String): Flow<Resource<String>> // Returns Task ID

    suspend fun pollTranslation(taskId: String, token: String): Flow<Resource<String>> // Returns Translated Text

    suspend fun generateSummary(
        text: String,
        style: String,
        language: String?,
        context: String?,
        token: String
    ): Flow<Resource<String>> // Returns Task ID

    suspend fun pollSummary(
        taskId: String,
        token: String
    ): Flow<Resource<RAGSummaryResponseDto>> // Returns Summary & PDF URL
}
