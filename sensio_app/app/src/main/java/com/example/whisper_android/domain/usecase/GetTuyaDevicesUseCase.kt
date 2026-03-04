package com.example.whisper_android.domain.usecase

import com.example.whisper_android.data.remote.dto.TuyaDevicesResponseDto
import com.example.whisper_android.domain.repository.TuyaRepository

class GetTuyaDevicesUseCase(
    private val repository: TuyaRepository
) {
    suspend operator fun invoke(): Result<TuyaDevicesResponseDto> {
        return repository.getDevices()
    }
}
