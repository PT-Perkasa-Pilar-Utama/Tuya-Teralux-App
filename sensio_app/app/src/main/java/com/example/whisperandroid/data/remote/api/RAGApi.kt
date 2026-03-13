package com.example.whisperandroid.data.remote.api

import com.example.whisperandroid.data.remote.dto.RAGRequestDto
import com.example.whisperandroid.data.remote.dto.RAGStatusDto
import com.example.whisperandroid.data.remote.dto.RAGSummaryRequestDto
import com.example.whisperandroid.data.remote.dto.SpeechResponseDto
import com.example.whisperandroid.data.remote.dto.TranscriptionSubmissionData
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.Header
import retrofit2.http.POST
import retrofit2.http.Path

/**
 * Retrofit interface for RAG (Retrieval-Augmented Generation) services.
 */
interface RAGApi {
    @POST("/api/models/rag/translate")
    suspend fun translate(
        @Body request: RAGRequestDto,
        @Header("Authorization") token: String,
        @Header("Idempotency-Key") idempotencyKey: String? = null,
        @Header("X-API-KEY") apiKey: String
    ): SpeechResponseDto<TranscriptionSubmissionData>

    @POST("/api/models/rag/summary")
    suspend fun summary(
        @Body request: RAGSummaryRequestDto,
        @Header("Authorization") token: String,
        @Header("Idempotency-Key") idempotencyKey: String? = null,
        @Header("X-API-KEY") apiKey: String
    ): SpeechResponseDto<TranscriptionSubmissionData>

    @GET("/api/models/rag/{task_id}")
    suspend fun getStatus(
        @Path("task_id") taskId: String,
        @Header("Authorization") token: String,
        @Header("X-API-KEY") apiKey: String
    ): SpeechResponseDto<RAGStatusDto>

    @POST("/api/models/rag/chat")
    suspend fun chat(
        @Body request: com.example.whisperandroid.data.remote.dto.RAGChatRequestDto,
        @Header("Authorization") token: String,
        @Header("Idempotency-Key") idempotencyKey: String? = null,
        @Header("X-API-KEY") apiKey: String
    ): SpeechResponseDto<com.example.whisperandroid.data.remote.dto.RAGChatResponseDto>
}
