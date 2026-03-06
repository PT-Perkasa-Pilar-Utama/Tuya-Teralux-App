package com.example.whisperandroid.domain.usecase

import com.example.whisperandroid.domain.model.TerminalRegistration
import com.example.whisperandroid.domain.repository.TerminalRepository

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
