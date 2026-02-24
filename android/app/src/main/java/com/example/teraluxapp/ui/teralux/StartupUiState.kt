package com.example.teraluxapp.ui.teralux

sealed class StartupUiState {
    data object Loading : StartupUiState()
    data object DeviceRegistered : StartupUiState()
    data class DeviceNotRegistered(val macAddress: String) : StartupUiState()
    data class Error(val message: String) : StartupUiState()
}
