package com.example.whisper_android.data.remote.dto

import com.google.gson.annotations.SerializedName

data class TuyaAuthResponseDto(
    @SerializedName("status") val status: Boolean,
    @SerializedName("message") val message: String,
    @SerializedName("data") val data: TuyaAuthDataDto?,
)

data class TuyaAuthDataDto(
    @SerializedName("access_token") val accessToken: String?,
    @SerializedName("expire_time") val expireTime: Int?,
    @SerializedName("refresh_token") val refreshToken: String?,
    @SerializedName("uid") val uid: String?,
)
