package com.example.whisper_android.data.remote.api

import com.example.whisper_android.data.remote.dto.RAGRequestDto
import com.example.whisper_android.data.remote.dto.RAGSummaryRequestDto
import com.example.whisper_android.data.remote.dto.SpeechResponseDto
import retrofit2.http.*

/**
 * Retrofit interface for RAG (Retrieval-Augmented Generation) services.
 */
interface RAGApi {

    @POST("/api/rag/translate")
    suspend fun translate(
        @Body request: RAGRequestDto,
        @Header("Authorization") token: String,
        @Header("X-API-KEY") apiKey: String = "teralux-api-key"
    ): SpeechResponseDto<String>

    @POST("/api/rag/summary")
    suspend fun summary(
        @Body request: RAGSummaryRequestDto,
        @Header("Authorization") token: String,
        @Header("X-API-KEY") apiKey: String = "teralux-api-key"
    ): SpeechResponseDto<String>
}
