package com.example.whisper_android.domain.usecase

import com.example.whisper_android.domain.model.TeraluxRegistration
import com.example.whisper_android.domain.repository.TeraluxRepository

class GetTeraluxByMacUseCase(
    private val repository: TeraluxRepository
) {
    suspend operator fun invoke(macAddress: String): Result<TeraluxRegistration?> {
        if (macAddress.isBlank()) {
            return Result.failure(
                IllegalArgumentException("MAC Address cannot be empty")
            )
        }
        return repository.getTeraluxByMac(macAddress)
    }
}
