package com.example.whisperandroid.presentation.auth

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.example.whisperandroid.data.di.NetworkModule
import com.example.whisperandroid.domain.repository.TerminalRepository
import com.example.whisperandroid.util.AppLog
import com.example.whisperandroid.util.DeviceUtils
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class AuthUiState(
    val isLoading: Boolean = true,
    val error: String? = null,
    val isMacRegistered: Boolean = false
)

class AuthViewModel(
    application: Application,
    private val terminalRepository: TerminalRepository
) : AndroidViewModel(application) {

    private val _uiState = MutableStateFlow(AuthUiState())
    val uiState: StateFlow<AuthUiState> = _uiState.asStateFlow()

    private val _navigationEvent = MutableSharedFlow<Boolean>()
    val navigationEvent: SharedFlow<Boolean> = _navigationEvent.asSharedFlow()

    private val TAG = "AuthViewModel"

    init {
        checkMacRegistration()
    }

    fun checkMacRegistration() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, error = null) }

            val deviceId = DeviceUtils.getDeviceId(getApplication())
            AppLog.i(TAG, "Checking MAC registration for device: $deviceId")

            val result = terminalRepository.getTerminalByMac(deviceId)

            result
                .onSuccess { registration ->
                    if (registration == null) {
                        AppLog.i(TAG, "MAC not registered for device: $deviceId")
                        _uiState.update {
                            it.copy(
                                isLoading = false,
                                isMacRegistered = false,
                                error = null
                            )
                        }
                        _navigationEvent.emit(false)
                    } else {
                        AppLog.i(TAG, "MAC registered for device: $deviceId, terminalId: ${registration.id}")
                        _uiState.update {
                            it.copy(
                                isLoading = false,
                                isMacRegistered = true,
                                error = null
                            )
                        }
                        _navigationEvent.emit(true)
                    }
                }
                .onFailure { e ->
                    AppLog.e(TAG, "Failed to check MAC registration", e)
                    _uiState.update {
                        it.copy(
                            isLoading = false,
                            isMacRegistered = false,
                            error = "Failed to check device registration. Please check your connection and try again."
                        )
                    }
                }
        }
    }

    fun clearError() {
        _uiState.update { it.copy(error = null) }
    }
}