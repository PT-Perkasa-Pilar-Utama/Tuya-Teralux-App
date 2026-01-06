package com.example.teraluxapp.data.model

import com.google.gson.annotations.SerializedName

data class Device(
    @SerializedName("id") val id: String,
    @SerializedName("remote_id") val remoteId: String?,  // For IR devices
    @SerializedName("name") val name: String,
    @SerializedName("category") val category: String?,
    @SerializedName("remote_category") val remoteCategory: String?,  // For merged IR devices
    @SerializedName("product_name") val productName: String?,
    @SerializedName("remote_product_name") val remoteProductName: String?,  // For merged IR devices
    @SerializedName("online") val online: Boolean,
    @SerializedName("icon") val icon: String?,
    @SerializedName("custom_name") val customName: String?,
    @SerializedName("model") val model: String?,
    @SerializedName("ip") val ip: String?,
    @SerializedName("local_key") val localKey: String?,
    @SerializedName("gateway_id") val gatewayId: String?,
    @SerializedName("create_time") val createTime: Long?,
    @SerializedName("update_time") val updateTime: Long?,
    @SerializedName("status") val status: List<DeviceStatus>?,
    @SerializedName("collections") val collections: String? // Changed to String to match backend
) {
    fun getParsedCollections(): List<Device> {
        if (collections.isNullOrEmpty()) return emptyList()
        return try {
            val type = object : com.google.gson.reflect.TypeToken<List<Device>>() {}.type
            com.google.gson.Gson().fromJson(collections, type)
        } catch (e: Exception) {
            emptyList()
        }
    }
}

data class DeviceStatus(
    @SerializedName("code") val code: String,
    @SerializedName("value") val value: Any? // Any? for generic JSON value
)

data class DeviceResponse(
    @SerializedName("devices") val devices: List<Device>,
    @SerializedName("total_devices") val totalDevices: Int,
    @SerializedName("current_page_count") val currentPageCount: Int
)

data class SingleDeviceResponse(
     @SerializedName("device") val device: Device
)

data class DeviceListResponse(
    @SerializedName("devices") val devices: List<Device>,
    @SerializedName("total") val total: Int,
    @SerializedName("page") val page: Int,
    @SerializedName("per_page") val perPage: Int
)
