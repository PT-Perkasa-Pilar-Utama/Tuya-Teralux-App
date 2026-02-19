package com.example.whisper_android.data.remote.dto

import com.google.gson.annotations.SerializedName

data class SendEmailRequestDto(
    @SerializedName("to") val to: List<String>,
    @SerializedName("subject") val subject: String,
    @SerializedName("body") val body: String,
)
