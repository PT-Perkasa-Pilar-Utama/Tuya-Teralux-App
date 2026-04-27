package com.example.whisperandroid.data.auth

import android.content.Context
import com.example.whisperandroid.data.local.TokenManager
import com.example.whisperandroid.data.remote.api.CommonApi
import com.example.whisperandroid.data.repository.LoginRepositoryImpl
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

object AuthStateManager {

    enum class AuthState(val message: String? = null) {
        Authenticated,
        NotRegistered,
        Unauthorized,
        Checking,
        Authenticating,
        Error
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

    suspend fun validateAuthWithBackend(): Result<AuthState> {
        val api = commonApi
        val ctx = context

        if (api == null || ctx == null) {
            return Result.failure(IllegalStateException("AuthStateManager not initialized. Call init() first."))
        }

        _currentAuthState.value = AuthState.Authenticating

        val tokenManager = TokenManager(ctx)
        val repository = LoginRepositoryImpl(api, ctx, tokenManager)

        val result = repository.login()

        return result.map { repoState ->
            when (repoState) {
                is LoginRepositoryImpl.AuthState.Authenticated -> AuthStateManager.AuthState.Authenticated
                is LoginRepositoryImpl.AuthState.NotRegistered -> AuthStateManager.AuthState.NotRegistered
                is LoginRepositoryImpl.AuthState.Unauthorized -> AuthStateManager.AuthState.Unauthorized
                is LoginRepositoryImpl.AuthState.Error -> AuthStateManager.AuthState.Error
            }
        }
    }

    suspend fun renewTokenIfNeeded(): Boolean {
        val ctx = context ?: return false
        val api = commonApi ?: return false

        val tokenManager = TokenManager(ctx)

        if (!tokenManager.isTokenExpired()) {
            return false
        }

        val repository = LoginRepositoryImpl(api, ctx, tokenManager)
        val result = repository.login()

        return result.isSuccess && result.getOrNull() == LoginRepositoryImpl.AuthState.Authenticated
    }
}