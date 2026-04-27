package com.example.whisperandroid.data.remote.api

import retrofit2.http.Body
import retrofit2.http.POST

interface CommonApi {
    @POST("/api/common/login")
    suspend fun login(@Body request: LoginRequest): retrofit2.Response<LoginResponse>
}

data class LoginRequest(
    val terminal_id: String
)

data class LoginResponse(
    val status: Boolean?,
    val message: String?,
    val data: LoginData?
)

data class LoginData(
    val terminal_id: String?,
    val access_token: String?,
    val message: String?
)
