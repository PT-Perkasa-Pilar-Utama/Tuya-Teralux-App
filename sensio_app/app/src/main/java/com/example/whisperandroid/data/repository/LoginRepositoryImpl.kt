package com.example.whisperandroid.data.repository

import android.content.Context
import android.provider.Settings
import com.example.whisperandroid.data.remote.api.CommonApi
import com.example.whisperandroid.data.remote.api.LoginRequest

class LoginRepositoryImpl(
    private val commonApi: CommonApi,
    private val context: Context
) {
    fun getAndroidId(): String {
        return Settings.Secure.getString(context.contentResolver, Settings.Secure.ANDROID_ID)
    }

    suspend fun login(): Result<Unit> {
        return try {
            val androidId = getAndroidId()
            val response = commonApi.login(LoginRequest(terminal_id = androidId))
            if (response.isSuccessful) {
                Result.success(Unit)
            } else if (response.code() == 404) {
                Result.failure(Exception("Terminal not found"))
            } else {
                Result.failure(Exception("Login failed: ${response.code()}"))
            }
        } catch (e: Exception) {
            Result.failure(Exception("Network error: ${e.message}"))
        }
    }
}