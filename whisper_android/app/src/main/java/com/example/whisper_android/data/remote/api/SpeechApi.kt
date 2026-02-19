package com.example.whisper_android.data.remote.api

import com.example.whisper_android.data.remote.dto.SpeechResponseDto
import com.example.whisper_android.data.remote.dto.TranscriptionStatusData
import com.example.whisper_android.data.remote.dto.TranscriptionSubmissionData
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
        @Header("Authorization") token: String,
        @Header("X-API-KEY") apiKey: String = "REDACTED_SECRET",
    ): SpeechResponseDto<TranscriptionSubmissionData>

    @GET("/api/speech/transcribe/{transcribe_id}")
    suspend fun getTranscriptionStatus(
        @Path("transcribe_id") taskId: String,
        @Header("Authorization") token: String,
        @Header("X-API-KEY") apiKey: String = "REDACTED_SECRET",
    ): SpeechResponseDto<TranscriptionStatusData>
}
