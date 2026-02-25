package com.example.whisper_android.data.remote.dto

import com.google.gson.annotations.SerializedName

data class MailTaskResponseDto(
    @SerializedName("task_id")
    val taskId: String,
    @SerializedName("task_status")
    val taskStatus: String
)

data class MailStatusDto(
    @SerializedName("status")
    val status: String,
    @SerializedName("result")
    val result: String? = null,
    @SerializedName("error")
    val error: String? = null,
    @SerializedName("started_at")
    val startedAt: String? = null,
    @SerializedName("duration_seconds")
    val durationSeconds: Double? = null,
    @SerializedName("expires_at")
    val expiresAt: String? = null,
    @SerializedName("expires_in_seconds")
    val expiresInSeconds: Long? = null
)
