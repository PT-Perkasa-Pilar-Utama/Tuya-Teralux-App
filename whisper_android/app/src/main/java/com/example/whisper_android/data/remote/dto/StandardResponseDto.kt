package com.example.whisper_android.data.remote.dto

import com.google.gson.annotations.SerializedName

/**
 * Generic standard response matching backend's dtos.StandardResponse
 */
data class StandardResponseDto<T>(
    @SerializedName("status") val status: Boolean,
    @SerializedName("message") val message: String,
    @SerializedName("data") val data: T? = null,
    @SerializedName("details") val details: Any? = null
)
