package com.example.whisper_android.domain.repository

import kotlinx.coroutines.flow.Flow

interface EmailRepository {
    suspend fun sendEmail(
        to: String,
        subject: String,
        template: String,
        token: String,
        attachmentPath: String? = null
    ): Flow<Resource<Boolean>> // true = success (task submitted + completed)

    suspend fun pollEmailStatus(
        taskId: String,
        token: String
    ): Flow<Resource<Boolean>> // true = completed successfully
}
