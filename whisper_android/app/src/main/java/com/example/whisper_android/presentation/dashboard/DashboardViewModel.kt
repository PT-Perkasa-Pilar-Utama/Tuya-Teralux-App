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
    private val authenticateUseCase: AuthenticateUseCase
) : ViewModel() {
    private val _uiState = MutableStateFlow(DashboardUiState(isLoading = true))
    val uiState: StateFlow<DashboardUiState> = _uiState.asStateFlow()

    init {
        authenticate()
    }

    fun authenticate() {
        viewModelScope.launch {
            _uiState.value = DashboardUiState(isLoading = true)

            // Call API to authenticate and get/refresh token
            val result: Result<String> = authenticateUseCase()

            result
                .onSuccess {
                    _uiState.value = DashboardUiState(isAuthenticated = true)
                }.onFailure { e ->
                    // If authentication fails (e.g. 401, network error),
                    // we treat it as unauthenticated and should redirect to login/register.
                    _uiState.value =
                        DashboardUiState(
                            isAuthenticated = false,
                            error = e.message ?: "Authentication failed"
                        )
                }
        }
    }
}
