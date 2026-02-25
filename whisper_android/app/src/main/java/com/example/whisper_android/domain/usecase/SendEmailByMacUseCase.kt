package com.example.whisper_android.domain.usecase

import com.example.whisper_android.domain.repository.EmailRepository
import com.example.whisper_android.domain.repository.Resource
import kotlinx.coroutines.flow.Flow

class SendEmailByMacUseCase(
    private val emailRepository: EmailRepository
) {
    suspend operator fun invoke(
        macAddress: String,
        subject: String,
        template: String,
        token: String,
        attachmentPath: String? = null,
        overrideEmails: List<String>? = null
    ): Flow<Resource<Boolean>> {
        return emailRepository.sendEmailByMac(
            macAddress,
            subject,
            template,
            token,
            attachmentPath,
            overrideEmails
        )
    }
}
