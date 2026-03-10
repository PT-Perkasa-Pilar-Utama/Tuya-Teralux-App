package com.example.whisperandroid.presentation.bootstrap

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.whisperandroid.data.di.NetworkModule
import com.example.whisperandroid.domain.usecase.AuthenticateUseCase
import com.example.whisperandroid.util.AppLog
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

data class BootstrapUiState(
    val isBootstrapped: Boolean = false,
    val isSyncing: Boolean = false,
    val error: String? = null
)

class AppBootstrapViewModel(
    private val authenticateUseCase: AuthenticateUseCase
) : ViewModel() {

    private val _uiState = MutableStateFlow(BootstrapUiState())
    val uiState: StateFlow<BootstrapUiState> = _uiState.asStateFlow()

    private val TAG = "AppBootstrap"
    private var hasStartedBootstrap = false

    fun bootstrap(forceRetry: Boolean = false) {
        if (forceRetry) {
            hasStartedBootstrap = false
        }

        if (hasStartedBootstrap || (_uiState.value.isBootstrapped && !forceRetry)) return
        hasStartedBootstrap = true

        viewModelScope.launch {
            AppLog.i(TAG, "Starting app bootstrap auth (forceRetry=$forceRetry)")
            _uiState.value = _uiState.value.copy(isSyncing = true, error = null, isBootstrapped = false)
            if (!forceRetry) {
                NetworkModule.setTuyaSyncReady(false)
            }

            // Authenticate only; device fetch is handled by Dashboard on first mount.
            val authResult = authenticateUseCase()
            authResult.onSuccess {
                AppLog.i(TAG, "Auth success")
                _uiState.value = _uiState.value.copy(
                    isBootstrapped = true,
                    isSyncing = false
                )
            }.onFailure { e ->
                AppLog.e(TAG, "Auth failed during bootstrap", e)
                _uiState.value = _uiState.value.copy(
                    isSyncing = false,
                    error = "Authentication failed: ${e.message}"
                )
            }
        }
    }
}
