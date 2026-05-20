package com.example.whisperandroid.data.remote.dto

import com.google.gson.annotations.SerializedName

data class UploadRecordingUrlRequestDto(
    @SerializedName("filename") val filename: String,
    @SerializedName("content_type") val contentType: String
)

data class UploadRecordingUrlResponseDto(
    @SerializedName("upload_url") val uploadUrl: String,
    @SerializedName("object_key") val objectKey: String,
    @SerializedName("filename") val filename: String
)

data class FinalizeRecordingRequestDto(
    @SerializedName("filename") val filename: String,
    @SerializedName("object_key") val objectKey: String,
    @SerializedName("mac_address") val macAddress: String
)

data class RecordingResponseDto(
    @SerializedName("id") val id: String,
    @SerializedName("filename") val filename: String,
    @SerializedName("original_name") val originalName: String,
    @SerializedName("audio_url") val audioUrl: String,
    @SerializedName("created_at") val createdAt: String
)
