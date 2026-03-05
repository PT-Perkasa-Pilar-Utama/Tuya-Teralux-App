package com.example.whisperandroid.data.remote.api

import com.example.whisperandroid.data.remote.dto.SpeechResponseDto
import com.example.whisperandroid.data.remote.dto.TranscriptionStatusDto
import com.example.whisperandroid.data.remote.dto.TranscriptionSubmissionData
import okhttp3.MultipartBody
import retrofit2.http.GET
import retrofit2.http.Header
import retrofit2.http.Multipart
import retrofit2.http.POST
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
}
