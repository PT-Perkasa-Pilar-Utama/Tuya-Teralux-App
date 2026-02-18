package com.example.whisper_android.domain.usecase

import com.example.whisper_android.domain.repository.EmailRepository

class SendEmailUseCase(
    private val emailRepository: EmailRepository
) {
    suspend operator fun invoke(to: String, subject: String, body: String, token: String): Result<Boolean> {
        return emailRepository.sendEmail(to, subject, body, token)
    }
}
