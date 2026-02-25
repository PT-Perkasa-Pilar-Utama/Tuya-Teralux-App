package com.example.whisper_android.domain.usecase

import com.example.whisper_android.domain.model.TerminalRegistration
import com.example.whisper_android.domain.repository.TerminalRepository

class RegisterTerminalUseCase(
    private val repository: TerminalRepository
) {
    suspend operator fun invoke(
        name: String,
        roomId: String,
        macAddress: String,
        deviceTypeId: String
    ): Result<TerminalRegistration> {
        if (name.isBlank()) return Result.failure(IllegalArgumentException("Name cannot be empty"))
        if (roomId.isBlank()) {
            return Result.failure(
                IllegalArgumentException("Room ID cannot be empty")
            )
        }
        if (deviceTypeId.isBlank()) {
            return Result.failure(
                IllegalArgumentException("Device Type ID cannot be empty")
            )
        }

        return repository.registerTerminal(name, roomId, macAddress, deviceTypeId)
    }
}
