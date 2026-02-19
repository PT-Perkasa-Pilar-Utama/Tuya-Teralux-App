package com.example.whisper_android.data.remote.api

import com.example.whisper_android.data.remote.dto.TeraluxRequestDto
import com.example.whisper_android.data.remote.dto.TeraluxResponseDto
import retrofit2.http.Body
import retrofit2.http.Header
import retrofit2.http.POST

interface TeraluxApi {
    @POST("/api/teralux")
    suspend fun registerTeralux(
        @Header("X-API-KEY") apiKey: String,
        @Body request: TeraluxRequestDto,
    ): TeraluxResponseDto

    @retrofit2.http.GET("/api/teralux/mac/{mac}")
    suspend fun getTeraluxByMac(
        @Header("X-API-KEY") apiKey: String,
        @retrofit2.http.Path("mac") mac: String,
    ): TeraluxResponseDto
}
