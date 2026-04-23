package com.example.whisperandroid.data.remote.dto

import com.google.gson.annotations.SerializedName

data class SendEmailRequestDto(
    @SerializedName("to") val to: List<String>,
    @SerializedName("subject") val subject: String,
    @SerializedName("template") val template: String? = "test",
    @SerializedName("attachment_path") val attachmentPath: String? = null,
    @SerializedName("data") val data: Map<String, Any>? = null
)

data class SendMailByMacRequestDto(
    @SerializedName("subject") val subject: String,
    @SerializedName("template") val template: String? = "test",
    @SerializedName("attachment_path") val attachmentPath: String? = null
)
