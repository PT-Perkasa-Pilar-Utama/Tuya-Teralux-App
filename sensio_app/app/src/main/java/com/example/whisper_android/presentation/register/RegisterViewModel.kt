package com.example.whisper_android.presentation.register

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.example.whisper_android.domain.usecase.AuthenticateUseCase
import com.example.whisper_android.domain.usecase.RegisterTerminalUseCase
import com.example.whisper_android.domain.usecase.GetTerminalByMacUseCase
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class RegisterUiState(
    val isLoading: Boolean = false,
    val error: String? = null,
    val message: String? = null, // Added message field
    val isSuccess: Boolean = false
)

class RegisterViewModel(
    application: Application,
    private val registerTerminalUseCase:
        com.example.whisper_android.domain.usecase.RegisterTerminalUseCase,
    private val getTerminalByMacUseCase:
        com.example.whisper_android.domain.usecase.GetTerminalByMacUseCase,
    private val authenticateUseCase: AuthenticateUseCase
) : AndroidViewModel(application) {
    private val _uiState = MutableStateFlow(RegisterUiState(isLoading = true)) // Start with loading
    val uiState: StateFlow<RegisterUiState> = _uiState.asStateFlow()

    init {
        checkRegistration()
    }

    fun checkRegistration() {
        viewModelScope.launch {
            val deviceId =
                com.example.whisper_android.util.DeviceUtils
                    .getDeviceId(getApplication())
            _uiState.update { it.copy(isLoading = true, error = null) }
            val result = getTerminalByMacUseCase(deviceId)

            result
                .onSuccess { registration ->
                    if (registration != null) {
                        // Registration exists, now try to authenticate
                        val authResult = authenticateUseCase()
                        authResult
                            .onSuccess {
                                _uiState.update {
                                    it.copy(
                                        isLoading = false,
                                        isSuccess = true,
                                        message = "Logged in successfully. Redirecting..."
                                    )
                                }
                            }.onFailure { e ->
                                // Registration exists but auth failed.
                                // DO NOT set isSuccess = true to avoid infinite loop.
                                _uiState.update {
                                    it.copy(
                                        isLoading = false,
                                        isSuccess = false,
                                        error =
                                        "Device registered but authentication failed: " +
                                            "${e.message}. Please check your connection or try " +
                                            "again."
                                    )
                                }
                            }
                    } else {
                        // Not registered, show form
                        _uiState.update { it.copy(isLoading = false, isSuccess = false) }
                    }
                }.onFailure { e ->
                    _uiState.update {
                        it.copy(
                            isLoading = false,
                            error = "Failed to check registration: ${e.message}"
                        )
                    }
                }
        }
    }

    fun register(
        name: String,
        roomId: String
    ) {
        if (name.isBlank() || roomId.isBlank()) {
            _uiState.value = _uiState.value.copy(error = "Name and Room ID are required")
            return
        }

        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true, error = null)
            val context = getApplication<Application>()
            val deviceId =
                com.example.whisper_android.util.DeviceUtils
                    .getDeviceId(context)
            val deviceTypeId =
                com.example.whisper_android.util.DeviceUtils
                    .getDeviceTypeId(context)

            val result = registerTerminalUseCase(name, roomId, deviceId, deviceTypeId)

            result
                .onSuccess {
                    // Register success, now authenticate to get token
                    val authResult = authenticateUseCase()

                    authResult
                        .onSuccess {
                            _uiState.value =
                                RegisterUiState(
                                    isSuccess = true,
                                    message = "Registration & Login successful! Welcome."
                                )
                        }.onFailure { e ->
                            _uiState.value =
                                RegisterUiState(
                                    isSuccess = true, // Redirect anyway, dashboard will retry auth
                                    message = "Registration successful. Login failed: ${e.message}"
                                )
                        }
                }.onFailure { e ->
                    _uiState.value = RegisterUiState(
                        error = e.message ?: "Unknown error",
                        isLoading = false
                    )
                }
        }
    }

    fun clearError() {
        _uiState.value = _uiState.value.copy(error = null)
    }

    fun clearMessage() {
        _uiState.value = _uiState.value.copy(message = null)
    }
}
