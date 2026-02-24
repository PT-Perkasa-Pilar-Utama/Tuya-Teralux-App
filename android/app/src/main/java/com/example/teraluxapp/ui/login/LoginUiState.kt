package com.example.teraluxapp.ui.login

sealed class LoginUiState {
    data object Idle : LoginUiState()
    data object Loading : LoginUiState()
    data class Success(val token: String, val uid: String) : LoginUiState()
    data class Error(val message: String) : LoginUiState()
}
