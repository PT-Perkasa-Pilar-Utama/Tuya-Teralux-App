package com.example.whisper_android.data.remote.dto

import com.google.gson.annotations.SerializedName

data class TeraluxRequestDto(
    @SerializedName("name") val name: String,
    @SerializedName("room_id") val roomId: String,
    @SerializedName("mac_address") val macAddress: String,
)
