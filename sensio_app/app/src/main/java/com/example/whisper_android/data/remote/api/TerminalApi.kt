package com.example.whisper_android.data.remote.api

import com.example.whisper_android.data.remote.dto.TerminalRequestDto
import com.example.whisper_android.data.remote.dto.TerminalResponseDto
import retrofit2.http.Body
import retrofit2.http.Header
import retrofit2.http.POST

interface TerminalApi {
    @POST("/api/terminal")
    suspend fun registerTerminal(
        @Header("X-API-KEY") apiKey: String,
        @Body request: TerminalRequestDto
    ): TerminalResponseDto

    @retrofit2.http.GET("/api/terminal/mac/{mac}")
    suspend fun getTerminalByMac(
        @Header("X-API-KEY") apiKey: String,
        @retrofit2.http.Path("mac") mac: String
    ): TerminalResponseDto
}
