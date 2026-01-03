package com.example.teraluxapp.data.model

import com.google.gson.annotations.SerializedName

data class Teralux(
    @SerializedName("id") val id: String,
    @SerializedName("mac_address") val macAddress: String,
    @SerializedName("room_id") val roomId: String,
    @SerializedName("name") val name: String,
    @SerializedName("created_at") val createdAt: String,
    @SerializedName("updated_at") val updated_at: String,
    @SerializedName("devices") val devices: List<Device>? = null
)

// Response DTO for single Teralux (matches backend)
typealias TeraluxResponseDTO = Teralux

data class TeraluxListResponse(
    @SerializedName("teralux") val teralux: List<Teralux>,
    @SerializedName("total") val total: Int
)

data class CreateTeraluxRequest(
    @SerializedName("mac_address") val macAddress: String,
    @SerializedName("room_id") val roomId: String,
    @SerializedName("name") val name: String
)

data class CreateTeraluxResponse(
    @SerializedName("teralux_id") val teraluxId: String
)

data class UpdateTeraluxRequest(
    @SerializedName("room_id") val roomId: String? = null,
    @SerializedName("name") val name: String? = null
)

data class CreateDeviceRequest(
    @SerializedName("teralux_id") val teraluxId: String,
    @SerializedName("name") val name: String
)

data class CreateDeviceResponse(
    @SerializedName("device_id") val deviceId: String
)
