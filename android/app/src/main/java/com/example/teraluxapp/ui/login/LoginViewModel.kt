package com.example.teraluxapp.ui.login

import android.content.Context
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.teraluxapp.data.repository.AuthRepository
import com.example.teraluxapp.data.repository.TeraluxRepository
import com.example.teraluxapp.utils.DeviceInfoUtils
import dagger.hilt.android.lifecycle.HiltViewModel
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class LoginViewModel @Inject constructor(
    private val authRepository: AuthRepository,
    private val teraluxRepository: TeraluxRepository,
    @ApplicationContext private val context: Context
) : ViewModel() {
    
    private val _uiState = MutableStateFlow<LoginUiState>(LoginUiState.Idle)
    val uiState: StateFlow<LoginUiState> = _uiState.asStateFlow()
    
    fun login() {
        viewModelScope.launch {
            _uiState.value = LoginUiState.Loading
            
            // Step 1: Authenticate
            authRepository.authenticate()
                .onSuccess { authResponse ->
                    // Step 2: Fetch teralux by MAC address
                    val macAddress = DeviceInfoUtils.getMacAddress(context)
                    teraluxRepository.checkDeviceRegistration(macAddress)
                        .onSuccess { teralux ->
                            if (teralux != null) {
                                // Save Teralux ID for later use
                                com.example.teraluxapp.utils.PreferencesManager.saveTeraluxId(context, teralux.id)
                                _uiState.value = LoginUiState.Success(
                                    token = authResponse.accessToken,
                                    uid = authResponse.uid
                                )
                            } else {
                                _uiState.value = LoginUiState.Error(
                                    "Device not registered. Please register first."
                                )
                            }
                        }
                        .onFailure { error ->
                            _uiState.value = LoginUiState.Error(
                                error.message ?: "Failed to fetch device info."
                            )
                        }
                }
                .onFailure { error ->
                    _uiState.value = LoginUiState.Error(
                        error.message ?: "Login gagal. Silakan coba lagi."
                    )
                }
        }
    }
}
