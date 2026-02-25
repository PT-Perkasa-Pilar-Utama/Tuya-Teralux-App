package com.example.whisper_android.data.repository

import com.example.whisper_android.common.util.getErrorMessage
import com.example.whisper_android.data.local.TokenManager
import com.example.whisper_android.data.remote.api.TuyaApi
import com.example.whisper_android.domain.repository.TuyaRepository

class TuyaRepositoryImpl(
    private val api: TuyaApi,
    private val apiKey: String,
    private val tokenManager: TokenManager
) : TuyaRepository {
    override suspend fun authenticate(): Result<String> =
        try {
            val response = api.authenticate(apiKey)
            if (response.status && response.data?.accessToken != null) {
                tokenManager.saveAccessToken(response.data.accessToken)
                Result.success(response.data.accessToken)
            } else {
                Result.failure(Exception(response.message))
            }
        } catch (e: retrofit2.HttpException) {
            Result.failure(Exception(e.getErrorMessage()))
        } catch (e: Exception) {
            Result.failure(e)
        }
}
