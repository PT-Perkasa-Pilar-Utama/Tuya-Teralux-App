package com.example.whisperandroid.data.remote.api

import com.example.whisperandroid.data.remote.dto.PipelineStatusDto
import com.example.whisperandroid.data.remote.dto.PipelineSubmitByUploadRequestDto
import com.example.whisperandroid.data.remote.dto.SpeechResponseDto
import com.example.whisperandroid.data.remote.dto.TranscriptionSubmissionData
import okhttp3.MultipartBody
import retrofit2.http.Body
import retrofit2.http.DELETE
import retrofit2.http.GET
import retrofit2.http.Header
import retrofit2.http.Multipart
import retrofit2.http.POST
import retrofit2.http.Part
import retrofit2.http.Path

/**
 * Retrofit interface for the unified AI Pipeline (Transcribe -> Summarize).
 */
interface PipelineApi {
    @Multipart
    @POST("/api/models/pipeline/job")
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
        @Header("X-API-KEY") apiKey: String
    ): SpeechResponseDto<TranscriptionSubmissionData>

    @GET("/api/models/pipeline/status/{task_id}")
    suspend fun getPipelineStatus(
        @Path("task_id") taskId: String,
        @Header("Authorization") token: String,
        @Header("X-API-KEY") apiKey: String
    ): SpeechResponseDto<PipelineStatusDto>

    @POST("/api/models/pipeline/job/by-upload")
    suspend fun executePipelineByUpload(
        @retrofit2.http.Body request: PipelineSubmitByUploadRequestDto,
        @Header("Authorization") token: String,
        @Header("X-API-KEY") apiKey: String
    ): SpeechResponseDto<TranscriptionSubmissionData>

    @DELETE("/api/models/pipeline/status/{task_id}")
    suspend fun cancelPipelineTask(
        @Path("task_id") taskId: String,
        @Header("Authorization") token: String,
        @Header("X-API-KEY") apiKey: String
    ): SpeechResponseDto<Unit>
}
