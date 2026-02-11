package com.example.whisper_android.domain.usecase

import com.example.whisper_android.domain.model.TeraluxRegistration
import com.example.whisper_android.domain.repository.TeraluxRepository

class RegisterTeraluxUseCase(
    private val repository: TeraluxRepository
) {
    suspend operator fun invoke(name: String, roomId: String, macAddress: String): Result<TeraluxRegistration> {
        if (name.isBlank()) return Result.failure(IllegalArgumentException("Name cannot be empty"))
        if (roomId.isBlank()) return Result.failure(IllegalArgumentException("Room ID cannot be empty"))
        
        return repository.registerTeralux(name, roomId, macAddress)
    }
}
