package com.example.whisper_android.domain.usecase

import com.example.whisper_android.domain.model.TerminalRegistration
import com.example.whisper_android.domain.repository.TerminalRepository

class GetTerminalByMacUseCase(
    private val repository: TerminalRepository
) {
    suspend operator fun invoke(macAddress: String): Result<TerminalRegistration?> {
        if (macAddress.isBlank()) {
            return Result.failure(
                IllegalArgumentException("MAC Address cannot be empty")
            )
        }
        return repository.getTerminalByMac(macAddress)
    }
}
