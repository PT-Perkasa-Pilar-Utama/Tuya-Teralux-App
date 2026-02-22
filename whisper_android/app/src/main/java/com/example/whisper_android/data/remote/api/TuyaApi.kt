package com.example.whisper_android.data.remote.api

import com.example.whisper_android.data.remote.dto.TuyaAuthResponseDto
import retrofit2.http.GET
import retrofit2.http.Header

interface TuyaApi {
    @GET("/api/tuya/auth")
    suspend fun authenticate(
        @Header("X-API-KEY") apiKey: String
    ): TuyaAuthResponseDto
}
