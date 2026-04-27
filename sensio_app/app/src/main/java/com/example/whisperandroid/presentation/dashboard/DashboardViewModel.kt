package com.example.whisperandroid.presentation.dashboard

import android.content.Context
import android.provider.Settings
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.whisperandroid.data.di.NetworkModule
import com.example.whisperandroid.data.repository.LoginRepositoryImpl
import com.example.whisperandroid.domain.repository.TerminalRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class DashboardUiState(
    val isBackgroundModeEnabled: Boolean = false,
    val isOverlayPermissionGranted: Boolean = false,
    val isTuyaSyncReady: Boolean = false,
    val terminalName: String = "",
    val terminalId: String = "",
    val macAddress: String = "",
    val isRegistered: Boolean = false,
    val isLoadingAuth: Boolean = true,
    val isSavingAuth: Boolean = false,
    val isAuthModalVisible: Boolean = false,
    val aiEngineProfile: String? = null,
    val isSavingAiEngineProfile: Boolean = false,
    val isMicrophoneActive: Boolean = false,
    val error: String? = null,
    val legacyMigrationWarning: String? = null
)

class DashboardViewModel(
    private val getTuyaDevicesUseCase:
        com.example.whisperandroid.domain.usecase.GetTuyaDevicesUseCase,
    private val backgroundAssistantModeStore:
        com.example.whisperandroid.data.local.BackgroundAssistantModeStore,
    private val terminalRepository: TerminalRepository,
    private val tokenManager: com.example.whisperandroid.data.local.TokenManager,
    private val tuyaSyncReadyFlow: kotlinx.coroutines.flow.StateFlow<Boolean> = NetworkModule.isTuyaSyncReady,
    private val onNavigateToAuth: () -> Unit = {}
) : ViewModel() {
    private var renewalAttemptedThisSession = false

    private val _uiState = MutableStateFlow(
        DashboardUiState(
            isBackgroundModeEnabled = backgroundAssistantModeStore.isEnabled.value
        )
    )
    val uiState: StateFlow<DashboardUiState> = _uiState.asStateFlow()

    private val loginRepository: LoginRepositoryImpl by lazy {
        LoginRepositoryImpl(
            NetworkModule.commonApi,
            NetworkModule.appContext,
            NetworkModule.tokenManager
        )
    }

    private var isAuthCheckInProgress = false

    init {
        observeBackgroundMode()
        observeTuyaSyncReady()
        loadCurrentAiEngineProfile()
        checkLoginStatus()
    }

    fun checkLoginStatus() {
        if (isAuthCheckInProgress) return
        isAuthCheckInProgress = true

        viewModelScope.launch {
            val result = loginRepository.login()
            result.onSuccess { state ->
                when (state) {
                    is LoginRepositoryImpl.AuthState.Authenticated -> {
                        _uiState.value = _uiState.value.copy(
                            isLoadingAuth = false,
                            error = null
                        )
                    }
                    is LoginRepositoryImpl.AuthState.NotRegistered,
                    is LoginRepositoryImpl.AuthState.Unauthorized -> {
                        _uiState.value = _uiState.value.copy(
                            isLoadingAuth = false,
                            error = "Authentication required. Please register your device."
                        )
                    }
                    is LoginRepositoryImpl.AuthState.Error -> {
                        _uiState.value = _uiState.value.copy(
                            isLoadingAuth = false,
                            error = state.message
                        )
                    }
                }
                isAuthCheckInProgress = false
            }.onFailure { e ->
                _uiState.value = _uiState.value.copy(
                    isLoadingAuth = false,
                    error = e.message ?: "Auth check failed"
                )
                isAuthCheckInProgress = false
            }
        }
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

    fun loadCurrentAiEngineProfile() {
        viewModelScope.launch {
            val macAddress = tokenManager.getMacAddress()
            if (macAddress.isNullOrEmpty()) {
                return@launch
            }

            val result = terminalRepository.getAiEngineProfileByMac(macAddress)
            result.onSuccess { state ->
                val warning = if (state?.source == "legacy_provider") {
                    "This terminal is still using a legacy AI provider setting. Choose Premium or Standard to migrate."
                } else {
                    null
                }

                _uiState.update {
                    it.copy(
                        aiEngineProfile = state?.profile,
                        legacyMigrationWarning = warning
                    )
                }
            }
        }
    }

    fun updateAiEngineProfile(profile: String?) {
        viewModelScope.launch {
            _uiState.update {
                it.copy(
                    isSavingAiEngineProfile = true,
                    error = null
                )
            }

            val terminalId = tokenManager.getTerminalId()
            if (terminalId.isNullOrEmpty()) {
                _uiState.update {
                    it.copy(
                        isSavingAiEngineProfile = false,
                        error = "Terminal ID not found. Please register your device."
                    )
                }
                return@launch
            }

            val result = terminalRepository.updateAiEngineProfile(terminalId, profile)
            result.onSuccess {
                _uiState.update {
                    it.copy(
                        isSavingAiEngineProfile = false,
                        aiEngineProfile = profile,
                        legacyMigrationWarning = null
                    )
                }
            }.onFailure { e ->
                val errorMsg = e.message ?: "Failed to update AI profile"
                val isNotFound = errorMsg.contains("404") || errorMsg.contains("not found", ignoreCase = true) ||
                    errorMsg.contains("Terminal not found", ignoreCase = true)

                _uiState.value = _uiState.value.copy(
                    isSavingAiEngineProfile = false,
                    error = if (isNotFound) "Terminal not found. Please register your device." else errorMsg
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
                val errorMsg = e.message ?: ""
                val is401 = errorMsg.contains("401") || errorMsg.contains("Unauthorized", ignoreCase = true)

                if (is401 && !renewalAttemptedThisSession) {
                    renewalAttemptedThisSession = true
                    android.util.Log.w("DashboardViewModel", "Received 401, attempting token renewal")

                    viewModelScope.launch {
                        val renewed = com.example.whisperandroid.data.auth.AuthStateManager.renewTokenIfNeeded()
                        if (renewed) {
                            android.util.Log.w("DashboardViewModel", "Token renewed, retrying fetchDevices")
                            fetchDevices(force = true)
                        } else {
                            android.util.Log.e("DashboardViewModel", "Token renewal failed, redirecting to auth")
                            onNavigateToAuth()
                        }
                    }
                } else if (is401) {
                    // Already attempted renewal this session, just redirect
                    android.util.Log.w("DashboardViewModel", "401 after renewal attempt, redirecting to auth")
                    onNavigateToAuth()
                } else {
                    // Keep non-blocking behavior: UI can continue even when sync fails.
                    NetworkModule.setTuyaSyncReady(true)
                    _uiState.value = _uiState.value.copy(error = e.message ?: "Failed to sync devices")
                    android.util.Log.e("DashboardViewModel", "Failed to sync devices", e)
                }
            }
        }
    }
}