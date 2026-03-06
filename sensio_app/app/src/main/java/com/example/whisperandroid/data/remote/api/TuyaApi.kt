package com.example.whisperandroid.data.remote.api

import com.example.whisperandroid.data.remote.dto.TuyaAuthResponseDto
import retrofit2.http.GET
import retrofit2.http.Header

interface TuyaApi {
    @GET("/api/tuya/auth")
    suspend fun authenticate(
        @Header("X-API-KEY") apiKey: String
    ): TuyaAuthResponseDto

    @GET("/api/tuya/devices")
    suspend fun getDevices(
        @Header("Authorization") token: String,
        @Header("X-API-KEY") apiKey: String
    ): com.example.whisperandroid.data.remote.dto.StandardResponseDto<
        com.example.whisperandroid.data.remote.dto.TuyaDevicesResponseDto
        >
}
