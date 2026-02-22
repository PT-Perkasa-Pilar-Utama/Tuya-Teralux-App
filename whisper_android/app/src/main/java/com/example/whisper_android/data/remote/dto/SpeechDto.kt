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
    @SerializedName("task_id") val taskId: String,
    @SerializedName("task_status") val taskStatus: String? = null,
    @SerializedName("recording_id") val recordingId: String? = null
)

/**
 * Nested status object in backend response.
 */
data class TranscriptionStatusDto(
    @SerializedName("status") val status: String,
    @SerializedName("result") val result: TranscriptionResultText? = null,
    @SerializedName("error") val error: String? = null,
    @SerializedName("trigger") val trigger: String? = null,
    @SerializedName("started_at") val startedAt: String? = null,
    @SerializedName("duration_seconds") val durationSeconds: Double? = null,
    @SerializedName("expires_at") val expiresAt: String? = null,
    @SerializedName("expires_in_seconds") val expiresInSeconds: Long? = null
)

data class TranscriptionResultText(
    @SerializedName("transcription") val transcription: String,
    @SerializedName("refined_text") val refinedText: String? = null,
    @SerializedName("detected_language") val detectedLanguage: String? = null
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
    @SerializedName("error") val error: String? = null,
    @SerializedName("trigger") val trigger: String? = null,
    @SerializedName("started_at") val startedAt: String? = null,
    @SerializedName("duration_seconds") val durationSeconds: Double? = null,
    @SerializedName("execution_result") val executionResult: com.google.gson.JsonElement? = null,
    @SerializedName("expires_at") val expiresAt: String? = null,
    @SerializedName("expires_in_seconds") val expiresInSeconds: Long? = null
)

/**
 * RAG Chat and Control DTOs
 */
data class RAGChatRequestDto(
    @SerializedName("prompt") val prompt: String,
    @SerializedName("language") val language: String? = null,
    @SerializedName("teralux_id") val teraluxId: String,
    @SerializedName("uid") val uid: String? = null
)

data class RAGChatResponseDto(
    @SerializedName("response") val response: String,
    @SerializedName("is_control") val isControl: Boolean,
    @SerializedName("redirect") val redirect: RedirectDto? = null
)

data class RedirectDto(
    @SerializedName("endpoint") val endpoint: String,
    @SerializedName("method") val method: String,
    @SerializedName("body") val body: Any? = null
)

data class RAGControlRequestDto(
    @SerializedName("prompt") val prompt: String,
    @SerializedName("teralux_id") val teraluxId: String
)

data class ControlResultDto(
    @SerializedName("message") val message: String,
    @SerializedName("device_id") val deviceId: String? = null,
    @SerializedName("command") val command: String? = null
)
