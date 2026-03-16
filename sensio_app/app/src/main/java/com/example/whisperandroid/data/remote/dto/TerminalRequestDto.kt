package com.example.whisperandroid.data.remote.dto

import com.google.gson.annotations.SerializedName

data class TerminalRequestDto(
    @SerializedName("name") val name: String,
    @SerializedName("room_id") val roomId: String,
    @SerializedName("mac_address") val macAddress: String,
    @SerializedName("device_type_id") val deviceTypeId: String
)

data class UpdateTerminalRequestDto(
    @SerializedName("name") val name: String? = null,
    @SerializedName("room_id") val roomId: String? = null,
    @SerializedName("mac_address") val macAddress: String? = null,
    @SerializedName("device_type_id") val deviceTypeId: String? = null,
    @SerializedName("ai_provider") val aiProvider: String? = null
)
