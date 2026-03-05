package com.example.whisperandroid.data.remote.api

import com.example.whisperandroid.data.remote.dto.SpeechResponseDto
import com.example.whisperandroid.data.remote.dto.TranscriptionStatusDto
import com.example.whisperandroid.data.remote.dto.TranscriptionSubmissionData
import com.example.whisperandroid.data.remote.dto.CreateUploadSessionRequestDto
import com.example.whisperandroid.data.remote.dto.UploadSessionResponseDto
import com.example.whisperandroid.data.remote.dto.UploadChunkAckDto
import com.example.whisperandroid.data.remote.dto.SubmitByUploadRequestDto
import okhttp3.MultipartBody
import retrofit2.http.GET
import retrofit2.http.Header
import retrofit2.http.Multipart
import retrofit2.http.POST
import retrofit2.http.PUT
import retrofit2.http.Part
import retrofit2.http.Path

/**
 * Retrofit interface for Speech Transcription services.
 */
interface SpeechApi {
    @Multipart
    @POST("/api/speech/transcribe")
    suspend fun transcribeAudio(
        @Part audio: MultipartBody.Part,
        @Part language: MultipartBody.Part,
        @Part macAddress: MultipartBody.Part? = null,
        @Header("Authorization") token: String,
        @Header("Idempotency-Key") idempotencyKey: String? = null,
        @Header("X-API-KEY") apiKey: String = "REDACTED_SECRET"
    ): SpeechResponseDto<TranscriptionSubmissionData>

    @GET("/api/speech/transcribe/{transcribe_id}")
    suspend fun getTranscriptionStatus(
        @Path("transcribe_id") taskId: String,
        @Header("Authorization") token: String,
        @Header("X-API-KEY") apiKey: String = "REDACTED_SECRET"
    ): SpeechResponseDto<TranscriptionStatusDto>

    // --- Chunk Upload Endpoints ---

    @POST("/api/speech/uploads/sessions")
    suspend fun createUploadSession(
        @retrofit2.http.Body request: CreateUploadSessionRequestDto,
        @Header("Authorization") token: String,
        @Header("X-API-KEY") apiKey: String = "REDACTED_SECRET"
    ): SpeechResponseDto<UploadSessionResponseDto>

    @PUT("/api/speech/uploads/sessions/{id}/chunks/{index}")
    suspend fun uploadChunk(
        @Path("id") sessionId: String,
        @Path("index") chunkIndex: Int,
        @retrofit2.http.Body chunk: okhttp3.RequestBody,
        @Header("Authorization") token: String,
        @Header("X-API-KEY") apiKey: String = "REDACTED_SECRET"
    ): SpeechResponseDto<UploadChunkAckDto>

    @GET("/api/speech/uploads/sessions/{id}")
    suspend fun getUploadSessionStatus(
        @Path("id") sessionId: String,
        @Header("Authorization") token: String,
        @Header("X-API-KEY") apiKey: String = "REDACTED_SECRET"
    ): SpeechResponseDto<UploadSessionResponseDto>

    @POST("/api/speech/transcribe/by-upload")
    suspend fun transcribeByUpload(
        @retrofit2.http.Body request: SubmitByUploadRequestDto,
        @Header("Authorization") token: String,
        @Header("X-API-KEY") apiKey: String = "REDACTED_SECRET"
    ): SpeechResponseDto<TranscriptionSubmissionData>
}
