package com.example.whisper_android.data.remote.dto

import com.google.gson.annotations.SerializedName

/**
 * Standard response from Speech API endpoints.
 */
data class SpeechResponseDto<T>(
    @SerializedName("status") val status: Boolean,
    @SerializedName("message") val message: String,
    @SerializedName("data") val data: T? = null,
    @SerializedName("details") val details: String? = null
)

/**
 * Data for async transcription submission response.
 */
data class TranscriptionSubmissionData(
    @SerializedName("task_id") val taskId: String
)

/**
 * Nested status object in backend response.
 */
data class TranscriptionStatusWrapper(
    @SerializedName("status") val status: String,
    @SerializedName("result") val result: TranscriptionResultText? = null,
    @SerializedName("expires_at") val expiresAt: String? = null,
    @SerializedName("expires_in_seconds") val expiresInSeconds: Long? = null
)

data class TranscriptionResultText(
    @SerializedName("filename") val filename: String? = null,
    @SerializedName("transcription") val transcription: String,
    @SerializedName("refined_text") val refinedText: String? = null,
    @SerializedName("detected_language") val detectedLanguage: String? = null
)

/**
 * Data for transcription status check response.
 */
data class TranscriptionStatusData(
    @SerializedName("task_id") val taskId: String,
    @SerializedName("task_status") val taskStatus: TranscriptionStatusWrapper? = null
)

/**
 * RAG Request DTOs
 */
data class RAGRequestDto(
    @SerializedName("text") val text: String,
    @SerializedName("language") val language: String? = null
)

data class RAGSummaryRequestDto(
    @SerializedName("text") val text: String,
    @SerializedName("language") val language: String? = null,
    @SerializedName("context") val context: String? = null,
    @SerializedName("style") val style: String? = null
)

data class RAGSummaryResponseDto(
    @SerializedName("summary") val summary: String,
    @SerializedName("pdf_url") val pdfUrl: String? = null
)

/**
 * Nested status object for RAG tasks.
 */
data class RAGStatusDto(
    @SerializedName("status") val status: String,
    @SerializedName("result") val result: String? = null,
    @SerializedName("execution_result") val executionResult: RAGSummaryResponseDto? = null,
    @SerializedName("expires_at") val expiresAt: String? = null,
    @SerializedName("expires_in_seconds") val expiresInSeconds: Long? = null
)

