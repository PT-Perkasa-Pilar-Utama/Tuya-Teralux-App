package com.example.whisper_android.data.repository

import com.example.whisper_android.common.util.getErrorMessage
import com.example.whisper_android.data.remote.api.EmailApi
import com.example.whisper_android.data.remote.dto.SendEmailRequestDto
import com.example.whisper_android.domain.repository.EmailRepository

class EmailRepositoryImpl(
    private val api: EmailApi,
) : EmailRepository {
    override suspend fun sendEmail(
        to: String,
        subject: String,
        body: String,
        token: String,
    ): Result<Boolean> =
        try {
            val request =
                SendEmailRequestDto(
                    to = listOf(to),
                    subject = subject,
                    body = body,
                )
            val response = api.sendEmail("Bearer $token", request)

            if (response.status) {
                Result.success(true)
            } else {
                Result.failure(Exception(response.message))
            }
        } catch (e: retrofit2.HttpException) {
            Result.failure(Exception(e.getErrorMessage()))
        } catch (e: Exception) {
            Result.failure(e)
        }
}
