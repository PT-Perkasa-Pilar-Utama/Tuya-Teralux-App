package com.example.whisperandroid.presentation.bootstrap

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.whisperandroid.data.di.NetworkModule
import com.example.whisperandroid.domain.repository.TerminalRepository
import com.example.whisperandroid.domain.usecase.AuthenticateUseCase
import com.example.whisperandroid.util.AppLog
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

data class BootstrapUiState(
    val isBootstrapped: Boolean = false,
    val isSyncing: Boolean = false,
    val error: String? = null,
    val shouldRedirectToRegister: Boolean = false
)

class AppBootstrapViewModel(
    private val authenticateUseCase: AuthenticateUseCase,
    private val terminalRepository: TerminalRepository,
    private val tokenManager: com.example.whisperandroid.data.local.TokenManager
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
            _uiState.value = _uiState.value.copy(isSyncing = true, error = null, isBootstrapped = false, shouldRedirectToRegister = false)
            if (!forceRetry) {
                NetworkModule.setTuyaSyncReady(false)
            }

            // Authenticate first
            val authResult = authenticateUseCase()
            authResult.onSuccess {
                AppLog.i(TAG, "Auth success, checking terminal registration")
                // After auth success, check terminal registration
                val macAddress = tokenManager.getMacAddress()
                if (macAddress.isNullOrEmpty()) {
                    AppLog.e(TAG, "MAC address not found in token manager")
                    _uiState.value = _uiState.value.copy(
                        isSyncing = false,
                        error = "Device MAC address not found",
                        shouldRedirectToRegister = true
                    )
                    return@launch
                }

                val terminalResult = terminalRepository.getTerminalByMac(macAddress)
                terminalResult.onSuccess { registration ->
                    if (registration == null) {
                        AppLog.i(TAG, "Terminal not registered for MAC: $macAddress")
                        _uiState.value = _uiState.value.copy(
                            isBootstrapped = true,
                            isSyncing = false,
                            shouldRedirectToRegister = true
                        )
                    } else {
                        AppLog.i(TAG, "Terminal registered: ${registration.id}")
                        _uiState.value = _uiState.value.copy(
                            isBootstrapped = true,
                            isSyncing = false,
                            shouldRedirectToRegister = false
                        )
                    }
                }.onFailure { e ->
                    AppLog.e(TAG, "Failed to check device registration", e)
                    _uiState.value = _uiState.value.copy(
                        isSyncing = false,
                        error = "Failed to check device registration. Please try again.",
                        shouldRedirectToRegister = false
                    )
                }
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
