package com.example.whisperandroid.data.remote.dto

import com.google.gson.annotations.SerializedName

data class AiEngineProfileResponseDto(
    val status: Boolean,
    val message: String,
    val data: AiEngineProfileDataDto?
)

data class AiEngineProfileDataDto(
    @SerializedName("terminal_id") val terminalId: String,
    val profile: String?,
    val source: String,
    @SerializedName("effective_provider") val effectiveProvider: String?,
    @SerializedName("effective_mode") val effectiveMode: String
)

data class UpdateAiEngineProfileRequestDto(
    val profile: String?
)
