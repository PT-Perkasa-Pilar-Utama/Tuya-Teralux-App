package com.example.whisperandroid.domain.usecase

import com.example.whisperandroid.domain.repository.TuyaRepository

class AuthenticateUseCase(
    private val repository: TuyaRepository
) {
    suspend operator fun invoke(): Result<String> = repository.authenticate()
}
