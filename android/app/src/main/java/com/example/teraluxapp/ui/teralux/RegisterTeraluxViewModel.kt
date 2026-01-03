package com.example.teraluxapp.ui.teralux

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.teraluxapp.data.repository.TeraluxRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class RegisterTeraluxViewModel @Inject constructor(
    private val teraluxRepository: TeraluxRepository
) : ViewModel() {
    
    private val _uiState = MutableStateFlow<RegisterUiState>(RegisterUiState.Idle)
    val uiState: StateFlow<RegisterUiState> = _uiState.asStateFlow()
    
    fun registerDevice(macAddress: String, roomId: String, deviceName: String) {
        viewModelScope.launch {
            _uiState.value = RegisterUiState.Loading
            
            teraluxRepository.registerDevice(macAddress, roomId, deviceName)
                .onSuccess {
                    _uiState.value = RegisterUiState.Success
                }
                .onFailure { error ->
                    _uiState.value = RegisterUiState.Error(
                        error.message ?: "Registrasi gagal. Silakan coba lagi."
                    )
                }
        }
    }
}
