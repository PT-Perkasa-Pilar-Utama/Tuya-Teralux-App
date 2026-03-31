package com.example.whisperandroid.data.repository

import com.example.whisperandroid.common.util.getErrorMessage
import com.example.whisperandroid.data.local.TokenManager
import com.example.whisperandroid.data.remote.api.TuyaApi
import com.example.whisperandroid.domain.repository.TuyaRepository

class TuyaRepositoryImpl(
    private val api: TuyaApi,
    private val tokenManager: TokenManager,
    private val apiKey: String
) : TuyaRepository {
    override suspend fun authenticate(): Result<String> =
        try {
            val response = api.authenticate(apiKey)
            if (response.status && response.data?.accessToken != null) {
                tokenManager.saveAccessToken(response.data.accessToken)
                response.data.uid?.takeIf { it.isNotBlank() }?.let { tokenManager.saveTuyaUid(it) }
                Result.success(response.data.accessToken)
            } else {
                Result.failure(Exception(response.message))
            }
        } catch (e: retrofit2.HttpException) {
            Result.failure(Exception(e.getErrorMessage()))
        } catch (e: Exception) {
            Result.failure(e)
        }

    override suspend fun getDevices():
        Result<com.example.whisperandroid.data.remote.dto.TuyaDevicesResponseDto> =
        try {
            val token = tokenManager.getAccessToken() ?: return Result.failure(
                Exception("No access token found")
            )
            val response = api.getDevices("Bearer $token")
            if (response.status && response.data != null) {
                Result.success(response.data)
            } else {
                Result.failure(Exception(response.message))
            }
        } catch (e: retrofit2.HttpException) {
            Result.failure(Exception(e.getErrorMessage()))
        } catch (e: Exception) {
            Result.failure(e)
        }
}
