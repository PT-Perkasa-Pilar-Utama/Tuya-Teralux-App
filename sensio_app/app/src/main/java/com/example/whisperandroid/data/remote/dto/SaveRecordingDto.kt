package com.example.whisperandroid.data.remote.dto

import com.google.gson.annotations.SerializedName

data class SaveRecordingRequestDto(
    @SerializedName("s3_key") val s3Key: String,
    @SerializedName("booking_id") val bookingId: String,
    @SerializedName("password_hash") val passwordHash: String
)

data class SaveRecordingResponseDto(
    @SerializedName("recording_id") val recordingId: String,
    @SerializedName("status") val status: String
)
