package com.example.whisper_android.data.remote.api

import com.example.whisper_android.data.remote.dto.MailStatusDto
import com.example.whisper_android.data.remote.dto.MailTaskResponseDto
import com.example.whisper_android.data.remote.dto.SendEmailRequestDto
import com.example.whisper_android.data.remote.dto.StandardResponseDto
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.Header
import retrofit2.http.POST
import retrofit2.http.Path

interface EmailApi {
    @POST("/api/mail/send")
    suspend fun sendEmail(
        @Header("Authorization") token: String,
        @Body request: SendEmailRequestDto
    ): StandardResponseDto<MailTaskResponseDto>

    @GET("/api/mail/status/{taskId}")
    suspend fun getEmailStatus(
        @Header("Authorization") token: String,
        @Path("taskId") taskId: String
    ): StandardResponseDto<MailStatusDto>
}
