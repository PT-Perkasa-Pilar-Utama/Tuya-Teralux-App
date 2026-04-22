package com.example.whisperandroid.data.repository

import android.content.Context
import android.provider.Settings
import com.example.whisperandroid.data.remote.api.CommonApi
import com.example.whisperandroid.data.remote.api.LoginRequest
import com.example.whisperandroid.presentation.splash.SplashUiState

class LoginRepositoryImpl(
    private val commonApi: CommonApi,
    private val context: Context
) {
    fun getAndroidId(): String {
        return Settings.Secure.getString(context.contentResolver, Settings.Secure.ANDROID_ID)
    }

    suspend fun login(): Result<SplashUiState> {
        return try {
            val androidId = getAndroidId()
            val response = commonApi.login(LoginRequest(terminal_id = androidId))
            if (response.isSuccessful) {
                Result.success(SplashUiState.Authenticated)
            } else {
                when (response.code()) {
                    404 -> Result.success(SplashUiState.NotRegistered)
                    401 -> Result.success(SplashUiState.Unauthorized)
                    else -> Result.success(SplashUiState.Error("Server error: ${response.code()}"))
                }
            }
        } catch (e: Exception) {
            Result.success(SplashUiState.Error("Network error: ${e.message}"))
        }
    }
}