package com.example.whisper_android.presentation.dashboard

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.whisper_android.domain.usecase.AuthenticateUseCase
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

data class DashboardUiState(
    val isLoading: Boolean = false,
    val isAuthenticated: Boolean = false,
    val error: String? = null
)

class DashboardViewModel(
    private val authenticateUseCase: AuthenticateUseCase,
    private val getTuyaDevicesUseCase: com.example.whisper_android.domain.usecase.GetTuyaDevicesUseCase
) : ViewModel() {
    private val _uiState = MutableStateFlow(DashboardUiState(isLoading = true))
    val uiState: StateFlow<DashboardUiState> = _uiState.asStateFlow()

    init {
        authenticate()
    }

    fun authenticate() {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true, error = null)

            // Call API to authenticate and get/refresh token
            val result: Result<String> = authenticateUseCase()

            result
                .onSuccess {
                    _uiState.value = _uiState.value.copy(isAuthenticated = true, isLoading = false)
                }.onFailure { e ->
                    _uiState.value =
                        _uiState.value.copy(
                            isLoading = false,
                            isAuthenticated = false,
                            error = e.message ?: "Authentication failed"
                        )
                }
        }
    }

    fun fetchDevices() {
        viewModelScope.launch {
            // Fetch devices but don't store in state as requested
            val result = getTuyaDevicesUseCase()
            result.onSuccess { response ->
                android.util.Log.d(
                    "DashboardViewModel",
                    "Devices synced with backend (Found ${response.devices.size})"
                )
            }.onFailure { e ->
                android.util.Log.e("DashboardViewModel", "Failed to sync devices", e)
            }
        }
    }
}
