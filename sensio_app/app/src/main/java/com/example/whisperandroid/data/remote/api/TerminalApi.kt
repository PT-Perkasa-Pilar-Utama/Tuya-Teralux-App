package com.example.whisperandroid.data.remote.api

import com.example.whisperandroid.data.remote.dto.TerminalRequestDto
import com.example.whisperandroid.data.remote.dto.TerminalResponseDto
import com.example.whisperandroid.data.remote.dto.UpdateTerminalRequestDto
import retrofit2.http.Body
import retrofit2.http.Header
import retrofit2.http.POST
import retrofit2.http.PUT
import retrofit2.http.Path

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

    @retrofit2.http.PUT("/api/terminal/{id}")
    suspend fun updateTerminal(
        @Header("Authorization") token: String,
        @Path("id") id: String,
        @Body request: UpdateTerminalRequestDto
    ): TerminalResponseDto

    @retrofit2.http.GET("/api/mqtt/credentials/{username}")
    suspend fun getMqttCredentials(
        @retrofit2.http.Header("Authorization") token: String,
        @retrofit2.http.Path("username") username: String
    ): com.example.whisperandroid.data.remote.dto.MqttCredentialsResponseDto
}
