package com.example.teraluxapp.ui.teralux

import android.content.Context
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
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
class StartupViewModel @Inject constructor(
    private val teraluxRepository: TeraluxRepository,
    @ApplicationContext private val context: Context
) : ViewModel() {
    
    private val _uiState = MutableStateFlow<StartupUiState>(StartupUiState.Loading)
    val uiState: StateFlow<StartupUiState> = _uiState.asStateFlow()
    
    init {
        checkDeviceRegistration()
    }
    
    private fun checkDeviceRegistration() {
        viewModelScope.launch {
            val macAddress = DeviceInfoUtils.getMacAddress(context)
            
            teraluxRepository.checkDeviceRegistration(macAddress)
                .onSuccess { teralux ->
                    _uiState.value = if (teralux != null) {
                        // Save Teralux ID for later use
                        com.example.teraluxapp.utils.PreferencesManager.saveTeraluxId(context, teralux.id)
                        StartupUiState.DeviceRegistered
                    } else {
                        StartupUiState.DeviceNotRegistered(macAddress)
                    }
                }
                .onFailure {
                    // Treat failure as not registered
                    _uiState.value = StartupUiState.DeviceNotRegistered(macAddress)
                }
        }
    }
}
