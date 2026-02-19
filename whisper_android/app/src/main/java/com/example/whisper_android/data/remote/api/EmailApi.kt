package com.example.whisper_android.data.remote.api

import com.example.whisper_android.data.remote.dto.SendEmailRequestDto
import com.example.whisper_android.data.remote.dto.StandardResponseDto
import retrofit2.http.Body
import retrofit2.http.Header
import retrofit2.http.POST

interface EmailApi {
    @POST("/api/email/send")
    suspend fun sendEmail(
        @Header("Authorization") token: String,
        @Body request: SendEmailRequestDto,
    ): StandardResponseDto<Any>
}
