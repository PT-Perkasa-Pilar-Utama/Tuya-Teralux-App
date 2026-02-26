package com.example.whisper_android.domain.usecase

import com.example.whisper_android.domain.repository.TuyaRepository
import com.example.whisper_android.data.remote.dto.TuyaDevicesResponseDto

class GetTuyaDevicesUseCase(
    private val repository: TuyaRepository
) {
    suspend operator fun invoke(): Result<TuyaDevicesResponseDto> {
        return repository.getDevices()
    }
}
