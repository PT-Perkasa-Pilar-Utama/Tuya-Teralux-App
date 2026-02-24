package com.example.whisper_android.domain.repository

interface TuyaRepository {
    suspend fun authenticate(): Result<String> // Returns access token
}
