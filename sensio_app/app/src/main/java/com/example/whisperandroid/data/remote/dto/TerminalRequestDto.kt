package com.example.whisperandroid.data.remote.dto

import com.google.gson.annotations.SerializedName

data class TerminalRequestDto(
    @SerializedName("name") val name: String,
    @SerializedName("room_id") val roomId: String,
    @SerializedName("mac_address") val macAddress: String,
    @SerializedName("device_type_id") val deviceTypeId: String
)
