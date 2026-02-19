package com.example.whisper_android.data.remote.dto

import com.google.gson.annotations.SerializedName

data class TeraluxResponseDto(
    @SerializedName("status") val status: Boolean,
    @SerializedName("message") val message: String,
    @SerializedName("data") val data: TeraluxDataDto?,
)

data class TeraluxDataDto(
    @SerializedName("teralux_id") val teraluxId: String?, // From Register response
    @SerializedName("id") val id: String?, // From Get response
    @SerializedName("name") val name: String?,
    @SerializedName("room_id") val roomId: String?,
    @SerializedName("mac_address") val macAddress: String?,
)
