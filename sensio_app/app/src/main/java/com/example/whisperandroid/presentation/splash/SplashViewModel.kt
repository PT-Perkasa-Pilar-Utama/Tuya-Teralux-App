package com.example.whisperandroid.presentation.splash

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.whisperandroid.data.repository.LoginRepositoryImpl
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

class SplashViewModel : ViewModel() {

    private val loginRepository: LoginRepositoryImpl by lazy {
        LoginRepositoryImpl(
            com.example.whisperandroid.data.di.NetworkModule.commonApi,
            com.example.whisperandroid.data.di.NetworkModule.appContext
        )
    }

    private val _uiState = MutableStateFlow<SplashUiState>(SplashUiState.Loading)
    val uiState: StateFlow<SplashUiState> = _uiState.asStateFlow()

    init {
        checkLoginStatus()
    }

    fun checkLoginStatus() {
        viewModelScope.launch {
            delay(1500)

            val result = loginRepository.login()
            result.onSuccess { state ->
                _uiState.value = state
            }.onFailure { e ->
                _uiState.value = SplashUiState.Error(e.message ?: "Unknown error")
            }
        }
    }

    fun retry() {
        checkLoginStatus()
    }
}