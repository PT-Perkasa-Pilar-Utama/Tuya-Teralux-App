package com.example.whisperandroid.domain.usecase

import com.example.whisperandroid.domain.repository.EmailRepository
import com.example.whisperandroid.domain.repository.Resource
import kotlinx.coroutines.flow.Flow

class SendEmailUseCase(
    private val emailRepository: EmailRepository
) {
    suspend operator fun invoke(
        to: List<String>,
        subject: String,
        template: String,
        token: String,
        attachmentPath: String? = null
    ): Flow<Resource<Boolean>> {
        return emailRepository.sendEmail(to, subject, template, token, attachmentPath)
    }
}
