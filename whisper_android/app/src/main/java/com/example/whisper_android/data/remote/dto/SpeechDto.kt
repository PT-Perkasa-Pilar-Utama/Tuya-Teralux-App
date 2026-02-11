package com.example.whisper_android.data.remote.dto

import com.google.gson.annotations.SerializedName

/**
 * Standard response from Speech API endpoints.
 */
data class SpeechResponseDto<T>(
    @SerializedName("status") val status: Boolean,
    @SerializedName("message") val message: String,
    @SerializedName("data") val data: T? = null
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
    @SerializedName("result") val result: TranscriptionResultText? = null
)

data class TranscriptionResultText(
    @SerializedName("transcription") val transcription: String,
    @SerializedName("translated_text") val translatedText: String? = null,
    @SerializedName("detected_language") val detectedLanguage: String? = null
)

/**
 * Data for transcription status check response.
 */
data class TranscriptionStatusData(
    @SerializedName("task_id") val taskId: String,
    @SerializedName("task_status") val taskStatus: TranscriptionStatusWrapper? = null
)
