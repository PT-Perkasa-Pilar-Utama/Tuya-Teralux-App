package com.example.whisper_android.domain.usecase

import com.example.whisper_android.domain.repository.EmailRepository
import com.example.whisper_android.domain.repository.Resource
import kotlinx.coroutines.flow.Flow

class SendEmailUseCase(
    private val emailRepository: EmailRepository
) {
    suspend operator fun invoke(
        to: String,
        subject: String,
        template: String,
        token: String,
        attachmentPath: String? = null
    ): Flow<Resource<Boolean>> {
        return emailRepository.sendEmail(to, subject, template, token, attachmentPath)
    }
}
