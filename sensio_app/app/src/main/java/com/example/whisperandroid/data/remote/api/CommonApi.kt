package com.example.whisperandroid.data.remote.api

import retrofit2.http.Body
import retrofit2.http.POST

interface CommonApi {
    @POST("/api/common/login")
    suspend fun login(@Body request: LoginRequest): retrofit2.Response<Unit>
}

data class LoginRequest(
    val terminal_id: String
)