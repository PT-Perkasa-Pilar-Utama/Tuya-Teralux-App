package com.example.whisperandroid.data.remote.api

import com.example.whisperandroid.data.remote.dto.TerminalRequestDto
import com.example.whisperandroid.data.remote.dto.TerminalResponseDto
import com.example.whisperandroid.data.remote.dto.UpdateTerminalRequestDto
import retrofit2.http.Body
import retrofit2.http.GET
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

    @GET("/api/terminal/mac/{mac}")
    suspend fun getTerminalByMac(
        @Header("X-API-KEY") apiKey: String,
        @Path("mac") mac: String
    ): TerminalResponseDto

    @GET("/api/terminal/mac/{mac}/ai-engine-profile")
    suspend fun getAiEngineProfileByMac(
        @Header("X-API-KEY") apiKey: String,
        @Path("mac") mac: String
    ): com.example.whisperandroid.data.remote.dto.AiEngineProfileResponseDto

    @PUT("/api/terminal/{id}")
    suspend fun updateTerminal(
        @Header("Authorization") token: String,
        @Path("id") id: String,
        @Body request: com.example.whisperandroid.data.remote.dto.UpdateTerminalRequestDto
    ): com.example.whisperandroid.data.remote.dto.StandardResponseDto<Unit>

    @retrofit2.http.GET("/api/mqtt/users/{username}")
    suspend fun getMqttCredentials(
        @retrofit2.http.Header("Authorization") token: String,
        @retrofit2.http.Path("username") username: String
    ): com.example.whisperandroid.data.remote.dto.MqttCredentialsResponseDto
}
