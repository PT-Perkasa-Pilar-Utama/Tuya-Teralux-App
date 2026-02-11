package com.example.whisper_android.presentation.register

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.example.whisper_android.domain.usecase.RegisterTeraluxUseCase
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import java.util.UUID

data class RegisterUiState(
    val isLoading: Boolean = false,
    val error: String? = null,
    val message: String? = null, // Added message field
    val isSuccess: Boolean = false
)

class RegisterViewModel(
    private val registerTeraluxUseCase: RegisterTeraluxUseCase,
    private val getTeraluxByMacUseCase: com.example.whisper_android.domain.usecase.GetTeraluxByMacUseCase
) : ViewModel() {
    private val _uiState = MutableStateFlow(RegisterUiState(isLoading = true)) // Start with loading
    val uiState: StateFlow<RegisterUiState> = _uiState.asStateFlow()

    init {
        checkRegistration()
    }

    private fun checkRegistration() {
        viewModelScope.launch {
            val macAddress = getMacAddress()
            val result = getTeraluxByMacUseCase(macAddress)
            
            result.onSuccess { registration ->
                if (registration != null) {
                    // Already registered, go to dashboard
                    _uiState.value = RegisterUiState(
                        isSuccess = true, 
                        message = "Device sudah terdaftar. Mengalihkan ke Dashboard..."
                    )
                } else {
                    // Not registered, show form
                    _uiState.value = RegisterUiState(isLoading = false)
                }
            }.onFailure {
                _uiState.value = RegisterUiState(isLoading = false)
            }
        }
    }

    fun register(name: String, roomId: String) {
        if (name.isBlank() || roomId.isBlank()) {
            _uiState.value = _uiState.value.copy(error = "Name and Room ID are required")
            return
        }

        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true, error = null)
            val macAddress = getMacAddress()
            
            val result = registerTeraluxUseCase(name, roomId, macAddress)
            
            result.onSuccess {
                _uiState.value = RegisterUiState(
                    isSuccess = true,
                    message = "Registrasi berhasil! Selamat datang."
                )
            }.onFailure { e ->
                _uiState.value = RegisterUiState(error = e.message ?: "Unknown error", isLoading = false)
            }
        }
    }

    private fun getMacAddress(): String {
        // ... (keep existing implementation)
        try {
            val all = java.util.Collections.list(java.net.NetworkInterface.getNetworkInterfaces())
            for (nif in all) {
                if (!nif.name.equals("wlan0", ignoreCase = true)) continue

                val macBytes = nif.hardwareAddress ?: return generateRandomMac()

                val res1 = StringBuilder()
                for (b in macBytes) {
                    res1.append(String.format("%02X:", b))
                }

                if (res1.isNotEmpty()) {
                    res1.deleteCharAt(res1.length - 1)
                }
                return res1.toString()
            }
        } catch (ex: Exception) {
            // Ignore
        }
        return generateRandomMac()
    }

    private fun generateRandomMac(): String {
        val random = java.util.Random()
        val mac = StringBuilder()
        for (i in 0 until 6) {
            val n = random.nextInt(256)
            mac.append(String.format("%02X%s", n, if (i < 5) ":" else ""))
        }
        return mac.toString()
    }

    fun clearError() {
        _uiState.value = _uiState.value.copy(error = null)
    }

    fun clearMessage() {
        _uiState.value = _uiState.value.copy(message = null)
    }
}
