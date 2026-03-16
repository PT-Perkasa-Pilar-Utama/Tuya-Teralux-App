package com.example.whisperandroid.presentation.dashboard

import android.content.Context
import android.provider.Settings
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.whisperandroid.data.di.NetworkModule
import com.example.whisperandroid.domain.repository.TerminalRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

data class DashboardUiState(
    val isBackgroundModeEnabled: Boolean = false,
    val isOverlayPermissionGranted: Boolean = false,
    val isTuyaSyncReady: Boolean = false,
    val aiProvider: String? = null,
    val isSavingAiProvider: Boolean = false,
    val error: String? = null,
    val shouldRedirectToRegister: Boolean = false
)

class DashboardViewModel(
    private val authenticateUseCase: com.example.whisperandroid.domain.usecase.AuthenticateUseCase,
    private val getTuyaDevicesUseCase:
        com.example.whisperandroid.domain.usecase.GetTuyaDevicesUseCase,
    private val backgroundAssistantModeStore:
        com.example.whisperandroid.data.local.BackgroundAssistantModeStore,
    private val terminalRepository: TerminalRepository,
    private val tokenManager: com.example.whisperandroid.data.local.TokenManager,
    private val tuyaSyncReadyFlow: kotlinx.coroutines.flow.StateFlow<Boolean> = NetworkModule.isTuyaSyncReady
) : ViewModel() {
    private val _uiState = MutableStateFlow(
        DashboardUiState(
            isBackgroundModeEnabled = backgroundAssistantModeStore.isEnabled.value
        )
    )
    val uiState: StateFlow<DashboardUiState> = _uiState.asStateFlow()

    init {
        observeBackgroundMode()
        observeTuyaSyncReady()
        loadCurrentAiProvider()
    }

    private fun observeTuyaSyncReady() {
        viewModelScope.launch {
            tuyaSyncReadyFlow.collect { ready ->
                _uiState.value = _uiState.value.copy(
                    isTuyaSyncReady = ready
                )
            }
        }
    }

    private fun observeBackgroundMode() {
        viewModelScope.launch {
            backgroundAssistantModeStore.isEnabled.collect { enabled ->
                _uiState.value = _uiState.value.copy(isBackgroundModeEnabled = enabled)
            }
        }
    }

    fun loadCurrentAiProvider() {
        viewModelScope.launch {
            // Get MAC address from token manager (source of truth for terminal lookup)
            val macAddress = tokenManager.getMacAddress()
            if (macAddress.isNullOrEmpty()) {
                return@launch
            }

            val result = terminalRepository.getTerminalByMac(macAddress)
            result.onSuccess { registration ->
                _uiState.value = _uiState.value.copy(
                    aiProvider = registration?.aiProvider
                )
            }
        }
    }

    fun updateAiProvider(provider: String?) {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(
                isSavingAiProvider = true,
                error = null
            )

            val terminalId = tokenManager.getTerminalId()
            if (terminalId.isNullOrEmpty()) {
                _uiState.value = _uiState.value.copy(
                    isSavingAiProvider = false,
                    error = "Terminal ID not found",
                    shouldRedirectToRegister = true
                )
                return@launch
            }

            val result = terminalRepository.updateTerminal(terminalId, provider)
            result.onSuccess {
                _uiState.value = _uiState.value.copy(
                    isSavingAiProvider = false,
                    aiProvider = provider
                )
            }.onFailure { e ->
                val errorMsg = e.message ?: "Failed to update AI provider"
                // Check if error is 404 - Terminal not found
                val isNotFound = errorMsg.contains("404") || errorMsg.contains("not found", ignoreCase = true) ||
                    errorMsg.contains("Terminal not found", ignoreCase = true)

                _uiState.value = _uiState.value.copy(
                    isSavingAiProvider = false,
                    error = if (isNotFound) "Terminal not found. Please register your device." else errorMsg,
                    shouldRedirectToRegister = isNotFound
                )
            }
        }
    }

    fun setBackgroundMode(context: Context, enabled: Boolean) {
        backgroundAssistantModeStore.setEnabled(enabled)
        val intent = android.content.`Intent`(context, com.example.whisperandroid.service.BackgroundAssistantService::class.java).apply {
            action = if (enabled) {
                com.example.whisperandroid.service.BackgroundAssistantService.ACTION_START_ASSISTANT
            } else {
                com.example.whisperandroid.service.BackgroundAssistantService.ACTION_STOP_ASSISTANT
            }
        }

        if (enabled) {
            if (android.os.Build.VERSION.SDK_INT >= android.os.Build.VERSION_CODES.O) {
                context.startForegroundService(intent)
            } else {
                context.startService(intent)
            }
        } else {
            // Even if stopping, we send the action first for clean shutdown, then stopService
            context.startService(intent)
            context.stopService(intent)
        }
    }

    fun checkOverlayPermission(context: Context) {
        _uiState.value = _uiState.value.copy(
            isOverlayPermissionGranted = Settings.canDrawOverlays(context)
        )
    }

    private var lastFetchAtMs = 0L
    private val FETCH_THROTTLE_MS = 5000L

    fun fetchDevices(force: Boolean = false) {
        val currentTime = System.currentTimeMillis()
        if (!force && currentTime - lastFetchAtMs < FETCH_THROTTLE_MS) {
            android.util.Log.d("DashboardViewModel", "Fetch throttled (last fetch was ${currentTime - lastFetchAtMs}ms ago)")
            return
        }

        viewModelScope.launch {
            lastFetchAtMs = currentTime
            val result = getTuyaDevicesUseCase()
            result.onSuccess { response ->
                NetworkModule.setTuyaSyncReady(true)
                _uiState.value = _uiState.value.copy(
                    error = null
                )
                android.util.Log.d(
                    "DashboardViewModel",
                    "Devices synced with backend (Request complete)"
                )
            }.onFailure { e ->
                // Keep non-blocking behavior: UI can continue even when sync fails.
                NetworkModule.setTuyaSyncReady(true)
                _uiState.value = _uiState.value.copy(error = e.message ?: "Failed to sync devices")
                android.util.Log.e("DashboardViewModel", "Failed to sync devices", e)
            }
        }
    }
}
