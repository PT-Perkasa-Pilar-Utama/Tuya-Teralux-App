package com.example.whisperandroid.presentation.authenticating

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.example.whisperandroid.data.auth.AuthStateManager
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

class AuthenticatingViewModel(
    application: Application,
    private val onNavigateToDashboard: () -> Unit,
    private val onNavigateToRegister: () -> Unit,
    private val onAuthError: (String) -> Unit
) : AndroidViewModel(application) {

    data class UiState(
        val isLoading: Boolean = true,
        val errorMessage: String? = null,
        val isRetryEnabled: Boolean = false,
        val retryCount: Int = 0
    )

    private val _uiState = MutableStateFlow(UiState())
    val uiState: StateFlow<UiState> = _uiState.asStateFlow()

    init {
        checkAuth()
    }

    fun checkAuth() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, errorMessage = null, isRetryEnabled = false) }

            val result = AuthStateManager.validateAuthWithBackend()

            result
                .onSuccess { authState ->
                    handleAuthState(authState)
                }
                .onFailure { exception ->
                    val canRetry = _uiState.value.retryCount < 3
                    _uiState.update {
                        it.copy(
                            isLoading = false,
                            errorMessage = exception.message ?: "Authentication failed",
                            isRetryEnabled = canRetry
                        )
                    }
                    onAuthError(exception.message ?: "Authentication failed")
                }
        }
    }

    fun retry() {
        val currentCount = _uiState.value.retryCount
        if (currentCount >= 3) {
            _uiState.update {
                it.copy(
                    isLoading = false,
                    errorMessage = "Connection failed. Please restart the app.",
                    isRetryEnabled = false,
                    retryCount = currentCount + 1
                )
            }
            return
        }
        _uiState.update { it.copy(retryCount = currentCount + 1) }
        checkAuth()
    }

    private fun handleAuthState(authState: AuthStateManager.AuthState) {
        when (authState) {
            AuthStateManager.AuthState.Authenticated -> {
                _uiState.update { it.copy(isLoading = false, errorMessage = null, isRetryEnabled = false) }
                onNavigateToDashboard()
            }
            AuthStateManager.AuthState.NotRegistered,
            AuthStateManager.AuthState.Unauthorized -> {
                _uiState.update { it.copy(isLoading = false, errorMessage = null, isRetryEnabled = false) }
                onNavigateToRegister()
            }
            AuthStateManager.AuthState.Error -> {
                val canRetry = _uiState.value.retryCount < 3
                _uiState.update {
                    it.copy(
                        isLoading = false,
                        errorMessage = "Authentication service unavailable. Please try again.",
                        isRetryEnabled = canRetry
                    )
                }
                onAuthError("Authentication service unavailable. Please try again.")
            }
            AuthStateManager.AuthState.Checking,
            AuthStateManager.AuthState.Authenticating -> {
                _uiState.update { it.copy(isLoading = true) }
            }
        }
    }
}