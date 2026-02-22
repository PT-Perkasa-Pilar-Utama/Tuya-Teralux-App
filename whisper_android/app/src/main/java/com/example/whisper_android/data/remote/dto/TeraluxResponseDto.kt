package com.example.whisper_android.data.remote.dto

import com.google.gson.annotations.SerializedName

data class TeraluxResponseDto(
    @SerializedName("status") val status: Boolean,
    @SerializedName("message") val message: String,
    @SerializedName("data") val data: TeraluxDataDto?
)

data class TeraluxDataDto(
    @SerializedName("teralux_id") val teraluxId: String? = null, // From Register response
    @SerializedName("teralux") val teralux: TeraluxItemDto? = null, // From GetByID/MAC response
    // Maintain flat fields for immediate backward compatibility in case they are used
    @SerializedName("id") val id: String? = null,
    @SerializedName("name") val name: String? = null,
    @SerializedName("room_id") val roomId: String? = null,
    @SerializedName("mac_address") val macAddress: String? = null
)

data class TeraluxItemDto(
    @SerializedName("id") val id: String,
    @SerializedName("mac_address") val macAddress: String,
    @SerializedName("room_id") val roomId: String,
    @SerializedName("name") val name: String,
    @SerializedName("created_at") val createdAt: String? = null,
    @SerializedName("updated_at") val updatedAt: String? = null
)
