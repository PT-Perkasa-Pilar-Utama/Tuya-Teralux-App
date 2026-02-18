package com.example.whisper_android.domain.repository

interface EmailRepository {
    suspend fun sendEmail(to: String, subject: String, body: String, token: String): Result<Boolean>
}
