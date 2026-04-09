package com.example.whisperandroid.presentation.dashboard

import android.content.Context
import android.content.Intent
import android.provider.Settings
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.whisperandroid.BuildConfig
import com.example.whisperandroid.data.di.NetworkModule
import com.example.whisperandroid.data.local.BackgroundAssistantModeStore
import com.example.whisperandroid.data.local.TokenManager
import com.example.whisperandroid.domain.repository.TerminalRepository
import com.example.whisperandroid.domain.usecase.AuthenticateUseCase
import com.example.whisperandroid.domain.usecase.GetTuyaDevicesUseCase
import com.example.whisperandroid.service.BackgroundAssistantService
import com.example.whisperandroid.util.AppLog
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
    private val authenticateUseCase: AuthenticateUseCase,
    private val getTuyaDevicesUseCase: GetTuyaDevicesUseCase,
    private val backgroundAssistantModeStore: BackgroundAssistantModeStore,
    private val terminalRepository: TerminalRepository,
    private val tokenManager: TokenManager,
    private val tuyaSyncReadyFlow: StateFlow<Boolean> = NetworkModule.isTuyaSyncReady,
    private val isAiEngineSelectorVisible: Boolean = BuildConfig.AI_ENGINE_SELECTOR_VISIBLE
) : ViewModel() {
    private val _uiState = MutableStateFlow(
        DashboardUiState(
            isBackgroundModeEnabled = backgroundAssistantModeStore.isEnabled.value
        )
    )
    val uiState: StateFlow<DashboardUiState> = _uiState.asStateFlow()

    private var hasResetAiProvider = false

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
                // After loading the current provider, check if we need to reset it
                resetAiProviderIfNeeded()
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

    /**
     * Resets the AI provider preference if the selector is hidden via BuildConfig flag.
     * This is a best-effort, one-shot operation per ViewModel lifecycle.
     * Errors are logged internally but not exposed to the UI state.
     */
    private fun resetAiProviderIfNeeded() {
        // Only reset if the selector is hidden
        if (isAiEngineSelectorVisible) {
            return
        }

        // Only reset once per ViewModel lifecycle
        if (hasResetAiProvider) {
            return
        }

        // Only reset if there's a provider to clear
        val currentProvider = _uiState.value.aiProvider
        if (currentProvider.isNullOrEmpty()) {
            return
        }

        // Perform the reset
        viewModelScope.launch {
            val terminalId = tokenManager.getTerminalId()
            if (terminalId.isNullOrEmpty()) {
                AppLog.w("DashboardViewModel", "Cannot reset AI provider: terminal ID is null/empty")
                return@launch
            }

            kotlin.runCatching {
                terminalRepository.updateTerminal(terminalId, "")
            }.onSuccess { result ->
                result.onSuccess {
                    hasResetAiProvider = true
                    _uiState.value = _uiState.value.copy(
                        aiProvider = ""
                    )
                    AppLog.d("DashboardViewModel", "AI provider reset successfully")
                }.onFailure { e ->
                    AppLog.e("DashboardViewModel", "Failed to reset AI provider: ${e.message}")
                }
            }.onFailure { e ->
                AppLog.e("DashboardViewModel", "Unexpected error during AI provider reset: ${e.message}")
            }
        }
    }

    fun setBackgroundMode(context: Context, enabled: Boolean) {
        backgroundAssistantModeStore.setEnabled(enabled)
        val intent = Intent(context, BackgroundAssistantService::class.java).apply {
            action = if (enabled) {
                BackgroundAssistantService.ACTION_START_ASSISTANT
            } else {
                BackgroundAssistantService.ACTION_STOP_ASSISTANT
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
