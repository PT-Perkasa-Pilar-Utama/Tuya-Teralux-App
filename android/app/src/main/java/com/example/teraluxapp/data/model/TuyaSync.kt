package com.example.teraluxapp.data.model

import com.google.gson.annotations.SerializedName

data class TuyaSyncDeviceDTO(
    @SerializedName("id") val id: String,
    @SerializedName("remote_id") val remoteId: String?,
    @SerializedName("online") val online: Boolean,
    @SerializedName("create_time") val createTime: Long,
    @SerializedName("update_time") val updateTime: Long
)

data class TuyaSyncResponse(
    @SerializedName("devices") val devices: List<TuyaSyncDeviceDTO>?,
    @SerializedName("total") val total: Int?
)
