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
            
            try {
                val authResponse = authRepository.authenticate().getOrThrow()
                val macAddress = DeviceInfoUtils.getMacAddress(context)
                val teralux = teraluxRepository.checkDeviceRegistration(macAddress).getOrThrow()

                if (teralux != null) {
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
            } catch (e: Exception) {
                _uiState.value = LoginUiState.Error(e.message ?: "An unexpected error occurred.")
            }
        }
    }
}
