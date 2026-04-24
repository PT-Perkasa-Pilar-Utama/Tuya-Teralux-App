package com.example.whisperandroid.data.auth

import android.content.Context
import android.provider.Settings
import com.example.whisperandroid.data.remote.api.CommonApi
import com.example.whisperandroid.data.remote.api.LoginRequest
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

object AuthStateManager {

    private const val PREFS_NAME = "sensio_prefs"

    enum class AuthState {
        Authenticated,
        NotRegistered,
        Unauthorized,
        Checking
    }

    private val _currentAuthState = MutableStateFlow(AuthState.Checking)
    val currentAuthState: StateFlow<AuthState> = _currentAuthState.asStateFlow()

    private var commonApi: CommonApi? = null
    private var context: Context? = null

    fun isInitialized(): Boolean = commonApi != null && context != null

    fun init(api: CommonApi, ctx: Context) {
        commonApi = api
        context = ctx.applicationContext
    }

    fun checkAuthOnStart(): AuthState {
        val ctx = context ?: return AuthState.Checking

        val prefs = ctx.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)
        val accessToken = prefs.getString("access_token", null)

        if (accessToken.isNullOrEmpty()) {
            _currentAuthState.value = AuthState.Unauthorized
            return AuthState.Unauthorized
        }

        _currentAuthState.value = AuthState.Authenticated
        return AuthState.Authenticated
    }

    @Deprecated("Do not use. TokenManager is the source of truth for token storage.")
    suspend fun autoLogin(): Result<AuthState> {
        val api = commonApi
        val ctx = context

        if (api == null || ctx == null) {
            return Result.failure(IllegalStateException("AuthStateManager not initialized. Call init() first."))
        }

        _currentAuthState.value = AuthState.Checking

        return try {
            val androidId = getAndroidId()
            val terminalId = androidIdToUuid(androidId)
            val response = api.login(LoginRequest(terminal_id = terminalId))

            val newState = if (response.isSuccessful) {
                val prefs = ctx.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)
                response.body()?.let { loginResponse ->
                    // Check if response has new token (renewed) or not (valid)
                    if (!loginResponse.data?.access_token.isNullOrEmpty()) {
                        // TokenManager handles storage
                    }
                    // If data is null/empty, token is still valid - don't overwrite
                }
                AuthState.Authenticated
            } else {
                when (response.code()) {
                    404 -> AuthState.NotRegistered
                    401 -> AuthState.Unauthorized
                    503 -> AuthState.Unauthorized  // Tuya service unavailable
                    else -> AuthState.Unauthorized
                }
            }

            _currentAuthState.value = newState
            Result.success(newState)
        } catch (e: Exception) {
            _currentAuthState.value = AuthState.Unauthorized
            Result.failure(e)
        }
    }

    fun clearAuth() {
        val ctx = context ?: return
        val prefs = ctx.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)
        prefs.edit()
            .remove("access_token")
            .remove("terminal_id")
            .remove("tuya_uid")
            .remove("mac_address")
            .apply()
        _currentAuthState.value = AuthState.Unauthorized
    }

    private fun getAndroidId(): String {
        val ctx = context ?: throw IllegalStateException("Context not initialized")
        return Settings.Secure.getString(ctx.contentResolver, Settings.Secure.ANDROID_ID)
    }

    private fun androidIdToUuid(androidId: String): String {
        val padded = androidId.padStart(16, '0')
        return padded.substring(0, 8) + "-" +
               padded.substring(8, 12) + "-" +
               padded.substring(12, 16) + "-" +
               "0000-000000000000"
    }
}