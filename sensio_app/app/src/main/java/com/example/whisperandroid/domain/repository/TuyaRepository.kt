package com.example.whisperandroid.domain.repository

interface TuyaRepository {
    suspend fun authenticate(): Result<String> // Returns access token
    suspend fun getDevices():
        Result<com.example.whisperandroid.data.remote.dto.TuyaDevicesResponseDto>
}
