package com.example.teraluxapp.data.model

import com.google.gson.annotations.SerializedName

data class Device(
    @SerializedName("id") val id: String,
    @SerializedName("name") val name: String,
    @SerializedName("category") val category: String,
    @SerializedName("product_name") val productName: String,
    @SerializedName("online") val online: Boolean,
    @SerializedName("icon") val icon: String,
    @SerializedName("status") val status: List<DeviceStatus>?,
    @SerializedName("ip") val ip: String?,
    @SerializedName("local_key") val localKey: String?,
    @SerializedName("gateway_id") val gatewayId: String?
)

data class DeviceStatus(
    @SerializedName("code") val code: String,
    @SerializedName("value") val value: Any? // Any? for generic JSON value
)

data class DeviceResponse(
    @SerializedName("devices") val devices: List<Device>,
    @SerializedName("total") val total: Int
)

data class SingleDeviceResponse(
     @SerializedName("device") val device: Device
)
