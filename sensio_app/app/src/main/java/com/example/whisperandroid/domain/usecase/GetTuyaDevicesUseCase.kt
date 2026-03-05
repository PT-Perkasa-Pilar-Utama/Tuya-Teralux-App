package com.example.whisperandroid.domain.usecase

import com.example.whisperandroid.data.remote.dto.TuyaDevicesResponseDto
import com.example.whisperandroid.domain.repository.TuyaRepository

class GetTuyaDevicesUseCase(
    private val repository: TuyaRepository
) {
    suspend operator fun invoke(): Result<TuyaDevicesResponseDto> {
        return repository.getDevices()
    }
}
