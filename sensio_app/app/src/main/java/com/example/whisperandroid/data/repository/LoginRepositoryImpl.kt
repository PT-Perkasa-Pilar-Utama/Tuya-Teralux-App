package com.example.whisperandroid.data.repository

import android.content.Context
import android.provider.Settings
import com.example.whisperandroid.data.local.TokenManager
import com.example.whisperandroid.data.remote.api.CommonApi
import com.example.whisperandroid.data.remote.api.LoginRequest

class LoginRepositoryImpl(
    private val commonApi: CommonApi,
    private val context: Context,
    private val tokenManager: TokenManager
) {
    sealed class AuthState {
        object Authenticated : AuthState()
        object NotRegistered : AuthState()
        object Unauthorized : AuthState()
        data class Error(val message: String) : AuthState()
    }

    fun getAndroidId(): String {
        return Settings.Secure.getString(context.contentResolver, Settings.Secure.ANDROID_ID)
    }

    private fun androidIdToUuid(androidId: String): String {
        val padded = androidId.padStart(16, '0')
        return padded.substring(0, 8) + "-" +
               padded.substring(8, 12) + "-" +
               padded.substring(12, 16) + "-" +
               "0000-000000000000"
    }

    suspend fun login(): Result<AuthState> {
        return try {
            val androidId = getAndroidId()
            val terminalId = androidIdToUuid(androidId)
            val response = commonApi.login(LoginRequest(terminal_id = terminalId))
            if (response.isSuccessful) {
                response.body()?.let { loginResponse ->
                    // Check if response indicates token is still valid (data is null or no access_token)
                    // or if it contains a new token (data has access_token)
                    when {
                        // Status is "valid" (data is null or no access_token) - keep existing token
                        loginResponse.data?.access_token.isNullOrEmpty() -> {
                            // Don't overwrite existing token - token is still valid
                            tokenManager.saveTerminalId(terminalId)
                        }
                        // Status is "renewed" - save new token
                        else -> {
                            loginResponse.data?.access_token?.let { token ->
                                tokenManager.saveAccessToken(token)
                            }
                            tokenManager.saveTerminalId(terminalId)
                        }
                    }
                }
                Result.success(AuthState.Authenticated)
            } else {
                when (response.code()) {
                    404 -> Result.success(AuthState.NotRegistered)
                    401 -> Result.success(AuthState.Unauthorized)
                    503 -> Result.success(AuthState.Error("Authentication service unavailable"))
                    else -> Result.success(AuthState.Error("Server error: ${response.code()}"))
                }
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun autoLogin(): Result<AuthState> = login()
}