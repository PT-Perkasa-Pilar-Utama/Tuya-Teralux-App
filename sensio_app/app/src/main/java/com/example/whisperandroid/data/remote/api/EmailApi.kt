package com.example.whisperandroid.data.remote.api

import com.example.whisperandroid.data.remote.dto.MailStatusDto
import com.example.whisperandroid.data.remote.dto.MailTaskResponseDto
import com.example.whisperandroid.data.remote.dto.SendEmailByMacRequestDto
import com.example.whisperandroid.data.remote.dto.SendEmailRequestDto
import com.example.whisperandroid.data.remote.dto.StandardResponseDto
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.Header
import retrofit2.http.POST
import retrofit2.http.Path

interface EmailApi {
    @POST("/api/mail/send/mac/{macAddress}")
    suspend fun sendEmailByMac(
        @Header("Authorization") token: String,
        @Path("macAddress") macAddress: String,
        @Body request: SendEmailByMacRequestDto
    ): StandardResponseDto<MailTaskResponseDto>

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
