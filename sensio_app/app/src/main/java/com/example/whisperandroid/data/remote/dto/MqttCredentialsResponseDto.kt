package com.example.whisperandroid.data.remote.dto

import com.google.gson.annotations.SerializedName

data class MqttCredentialsResponseDto(
    @SerializedName("status") val status: Boolean,
    @SerializedName("message") val message: String,
    @SerializedName("data") val data: MqttCredentialsDataDto?
)

data class MqttCredentialsDataDto(
    @SerializedName("username") val username: String,
    @SerializedName("password") val password: String
)
