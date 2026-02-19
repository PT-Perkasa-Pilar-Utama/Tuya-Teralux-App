package com.example.whisper_android.domain.usecase

import com.example.whisper_android.domain.repository.TuyaRepository

class AuthenticateUseCase(
    private val repository: TuyaRepository,
) {
    suspend operator fun invoke(): Result<String> = repository.authenticate()
}
