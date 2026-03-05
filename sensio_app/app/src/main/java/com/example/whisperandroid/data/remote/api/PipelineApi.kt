package com.example.whisperandroid.data.remote.api

import com.example.whisperandroid.data.remote.dto.SpeechResponseDto
import com.example.whisperandroid.data.remote.dto.TranscriptionSubmissionData
import com.example.whisperandroid.data.remote.dto.PipelineStatusDto
import okhttp3.MultipartBody
import retrofit2.http.*

/**
 * Retrofit interface for the unified AI Pipeline (Transcribe -> Summarize).
 */
interface PipelineApi {
    @Multipart
    @POST("/api/pipeline/job")
    suspend fun executePipeline(
        @Part audio: MultipartBody.Part,
        @Part("language") language: String? = null,
        @Part("target_language") targetLanguage: String? = null,
        @Part("summarize") summarize: Boolean = true,
        @Part("refine") refine: Boolean? = null,
        @Part("diarize") diarize: Boolean = false,
        @Part("context") context: String? = null,
        @Part("style") style: String? = null,
        @Part("date") date: String? = null,
        @Part("location") location: String? = null,
        @Part("participants") participants: String? = null,
        @Part("mac_address") macAddress: String? = null,
        @Header("Authorization") token: String,
        @Header("Idempotency-Key") idempotencyKey: String? = null,
        @Header("X-API-KEY") apiKey: String = "REDACTED_SECRET"
    ): SpeechResponseDto<TranscriptionSubmissionData>

    @GET("/api/pipeline/status/{task_id}")
    suspend fun getPipelineStatus(
        @Path("task_id") taskId: String,
        @Header("Authorization") token: String,
        @Header("X-API-KEY") apiKey: String = "REDACTED_SECRET"
    ): SpeechResponseDto<PipelineStatusDto>
}
