package com.example.whisperandroid.presentation.splash

sealed class SplashUiState {
    object Loading : SplashUiState()
    object Authenticated : SplashUiState()
    object NotRegistered : SplashUiState()
    object Unauthorized : SplashUiState()
    data class Error(val message: String) : SplashUiState()
}