package com.example.teraluxapp.data.model

import com.google.gson.annotations.SerializedName

data class Teralux(
    @SerializedName("id") val id: String,
    @SerializedName("mac_address") val macAddress: String,
    @SerializedName("name") val name: String,
    @SerializedName("created_at") val createdAt: String,
    @SerializedName("updated_at") val updated_at: String
)

// Response DTO for single Teralux (matches backend)
typealias TeraluxResponseDTO = Teralux

data class TeraluxListResponse(
    @SerializedName("teralux") val teralux: List<Teralux>,
    @SerializedName("total") val total: Int
)

data class CreateTeraluxRequest(
    @SerializedName("mac_address") val macAddress: String,
    @SerializedName("name") val name: String
)

data class CreateTeraluxResponse(
    @SerializedName("teralux_id") val teraluxId: String
)
