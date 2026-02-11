package com.example.whisper_android.data.repository

import com.example.whisper_android.data.remote.api.TuyaApi
import com.example.whisper_android.domain.repository.TuyaRepository
import com.example.whisper_android.data.local.TokenManager

class TuyaRepositoryImpl(
    private val api: TuyaApi,
    private val apiKey: String,
    private val tokenManager: TokenManager
) : TuyaRepository {
    override suspend fun authenticate(): Result<String> {
        return try {
            val response = api.authenticate(apiKey)
            if (response.status && response.data?.accessToken != null) {
                tokenManager.saveAccessToken(response.data.accessToken)
                Result.success(response.data.accessToken)
            } else {
                Result.failure(Exception(response.message))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}
