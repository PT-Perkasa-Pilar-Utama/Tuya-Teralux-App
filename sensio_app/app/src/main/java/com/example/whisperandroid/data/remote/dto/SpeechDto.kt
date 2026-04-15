package com.example.whisperandroid.data.remote.dto

import com.google.gson.annotations.SerializedName

/**
 * Standard response from Speech API endpoints.
 */
data class SpeechResponseDto<T>(
    @SerializedName("status") val status: Boolean,
    @SerializedName("message") val message: String,
    @SerializedName("data") val data: T? = null,
    @SerializedName("details") val details: Any? = null
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
    @SerializedName("detected_language") val detectedLanguage: String? = null,

    // ASR Quality Gate Metadata
    @SerializedName("transcript_valid") val transcriptValid: Boolean? = null,
    @SerializedName("transcript_rejection_reason") val transcriptRejectionReason: String? = null,
    @SerializedName("audio_class") val audioClass: String? = null,
    @SerializedName("provider_skipped") val providerSkipped: Boolean? = null,
    @SerializedName("provider_name") val providerName: String? = null
)

/**
 * RAG Request DTOs
 */
data class RAGRequestDto(
    @SerializedName("text") val text: String,
    @SerializedName("language") val language: String? = null,
    @SerializedName("mac_address") val macAddress: String? = null
)

data class RAGSummaryRequestDto(
    @SerializedName("text") val text: String,
    @SerializedName("language") val language: String? = null,
    @SerializedName("context") val context: String? = null,
    @SerializedName("style") val style: String? = null,
    @SerializedName("date") val date: String? = null,
    @SerializedName("location") val location: String? = null,
    @SerializedName("participants") val participants: List<String>? = null,
    @SerializedName("mac_address") val macAddress: String? = null
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
    @SerializedName("summary") val summary: String? = null,
    @SerializedName("pdf_url") val pdfUrl: String? = null,
    @SerializedName("agenda_context") val agendaContext: String? = null,
    @SerializedName("meeting_context") val meetingContext: String? = null,
    @SerializedName("language") val language: String? = null,
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
    @SerializedName("terminal_id") val terminalId: String,
    @SerializedName("uid") val uid: String? = null,
    @SerializedName("request_id") val requestId: String? = null
)

data class RAGChatResponseDto(
    @SerializedName("response") val response: String? = null,
    @SerializedName("is_control") val isControl: Boolean? = null,
    @SerializedName("is_blocked") val isBlocked: Boolean? = null,
    @SerializedName("redirect") val redirect: RedirectDto? = null,
    @SerializedName("request_id") val requestId: String? = null,
    @SerializedName("source") val source: String? = null,
    @SerializedName("instance_id") val instanceId: String? = null
)

data class RedirectDto(
    @SerializedName("endpoint") val endpoint: String,
    @SerializedName("method") val method: String,
    @SerializedName("body") val body: Any? = null
)

data class RAGControlRequestDto(
    @SerializedName("prompt") val prompt: String,
    @SerializedName("terminal_id") val terminalId: String
)

data class ControlResultDto(
    @SerializedName("message") val message: String,
    @SerializedName("device_id") val deviceId: String? = null,
    @SerializedName("command") val command: String? = null
)

/**
 * Unified Pipeline DTOs
 */
data class PipelineStatusDto(
    @SerializedName("task_id") val taskId: String,
    @SerializedName("overall_status") val overallStatus: String,
    @SerializedName("stages") val stages: Map<String, PipelineStageStatus>? = null,
    @SerializedName("started_at") val startedAt: String? = null,
    @SerializedName("duration_seconds") val durationSeconds: Double? = null,
    @SerializedName("expires_at") val expiresAt: String? = null,
    @SerializedName("expires_in_seconds") val expiresInSeconds: Long? = null
)

data class PipelineStageStatus(
    @SerializedName("status") val status: String,
    @SerializedName("result") val result: Any? = null,
    @SerializedName("error") val error: String? = null,
    @SerializedName("started_at") val startedAt: String? = null,
    @SerializedName("duration_seconds") val durationSeconds: Double? = null
)

/**
 * Chunk Upload Session DTOs
 */
data class CreateUploadSessionRequestDto(
    @SerializedName("file_name") val fileName: String,
    @SerializedName("total_size_bytes") val totalSizeBytes: Long,
    @SerializedName("mime_type") val mimeType: String? = null,
    @SerializedName("chunk_size_bytes") val chunkSizeByes: Int? = null
)

data class UploadSessionResponseDto(
    @SerializedName("session_id") val sessionId: String,
    @SerializedName("state") val state: String,
    @SerializedName("total_chunks") val totalChunks: Int,
    @SerializedName("chunk_size_bytes") val chunkSizeByes: Int,
    @SerializedName("total_size_bytes") val totalSizeBytes: Long,
    @SerializedName("received_bytes") val receivedBytes: Long? = 0,
    @SerializedName("missing_ranges") val missingRanges: List<String>? = null,
    @SerializedName("expires_at") val expiresAt: String
)

data class UploadChunkAckDto(
    @SerializedName("received_chunks") val receivedChunks: Int,
    @SerializedName("received_bytes") val receivedBytes: Long,
    @SerializedName("is_duplicate") val isDuplicate: Boolean,
    @SerializedName("state") val state: String
)

/**
 * Submit job by already uploaded session
 */
data class SubmitByUploadRequestDto(
    @SerializedName("session_id") val sessionId: String,
    @SerializedName("language") val language: String? = "id",
    @SerializedName("mac_address") val macAddress: String? = null,
    @SerializedName("diarize") val diarize: Boolean = false,
    @SerializedName("idempotency_key") val idempotencyKey: String? = null
)

data class PipelineSubmitByUploadRequestDto(
    @SerializedName("session_id") val sessionId: String,
    @SerializedName("language") val language: String? = "id",
    @SerializedName("target_language") val targetLanguage: String? = "en",
    @SerializedName("summarize") val summarize: Boolean = true,
    @SerializedName("refine") val refine: Boolean? = null,
    @SerializedName("diarize") val diarize: Boolean = false,
    @SerializedName("context") val context: String? = null,
    @SerializedName("style") val style: String? = null,
    @SerializedName("date") val date: String? = null,
    @SerializedName("location") val location: String? = null,
    @SerializedName("participants") val participants: List<String>? = null,
    @SerializedName("mac_address") val macAddress: String? = null,
    @SerializedName("idempotency_key") val idempotencyKey: String? = null
)
