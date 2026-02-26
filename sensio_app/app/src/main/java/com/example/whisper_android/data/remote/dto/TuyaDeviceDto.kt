package com.example.whisper_android.data.remote.dto

import com.google.gson.annotations.SerializedName

data class TuyaDevicesResponseDto(
    @SerializedName("devices") val devices: List<TuyaDeviceDto>,
    @SerializedName("total_devices") val totalDevices: Int,
    @SerializedName("current_page_count") val currentPageCount: Int,
    @SerializedName("page") val page: Int,
    @SerializedName("per_page") val perPage: Int,
    @SerializedName("total") val total: Int
)

data class TuyaDeviceDto(
    @SerializedName("id") val id: String,
    @SerializedName("remote_id") val remoteId: String? = null,
    @SerializedName("name") val name: String,
    @SerializedName("category") val category: String,
    @SerializedName("product_name") val productName: String,
    @SerializedName("online") val online: Boolean,
    @SerializedName("icon") val icon: String,
    @SerializedName("status") val status: List<TuyaDeviceStatusDto>,
    @SerializedName("custom_name") val customName: String? = null
)

data class TuyaDeviceStatusDto(
    @SerializedName("code") val code: String,
    @SerializedName("value") val value: Any
)
