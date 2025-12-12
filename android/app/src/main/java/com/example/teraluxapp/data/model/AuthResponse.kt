package com.example.teraluxapp.data.model

import com.google.gson.annotations.SerializedName

data class AuthResponse(
    @SerializedName("access_token") val accessToken: String,
    @SerializedName("expire_time") val expireTime: Int,
    @SerializedName("refresh_token") val refreshToken: String,
    @SerializedName("uid") val uid: String
)
