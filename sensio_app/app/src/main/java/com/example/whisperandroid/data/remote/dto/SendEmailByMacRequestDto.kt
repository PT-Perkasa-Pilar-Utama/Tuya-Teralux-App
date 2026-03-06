package com.example.whisperandroid.data.remote.dto

import com.google.gson.annotations.SerializedName

data class SendEmailByMacRequestDto(
    @SerializedName("subject") val subject: String,
    @SerializedName("template") val template: String? = null,
    @SerializedName("data") val data: Map<String, Any>? = null,
    @SerializedName("attachment_path") val attachmentPath: String? = null
)
