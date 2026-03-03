package com.example.whisper_android.presentation.summary

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.example.whisper_android.domain.repository.Resource
import com.example.whisper_android.presentation.components.UiState
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.collectLatest
import kotlinx.coroutines.launch

data class SummariesData(
    val idSummary: String = "",
    val enSummary: String = ""
)

class SummaryViewModel(application: Application) : AndroidViewModel(application) {
    private val _summaries = MutableStateFlow(SummariesData())
    val summaries: StateFlow<SummariesData> = _summaries

    private val _selectedLanguage = MutableStateFlow("id")
    val selectedLanguage: StateFlow<String> = _selectedLanguage

    private val mqttHelper = com.example.whisper_android.data.di.NetworkModule.mqttHelper

    private val _mqttStatus = MutableStateFlow(com.example.whisper_android.util.MqttHelper.MqttConnectionStatus.DISCONNECTED)
    val mqttStatus: StateFlow<com.example.whisper_android.util.MqttHelper.MqttConnectionStatus> = _mqttStatus

    init {
        loadSummaries()
        
        mqttHelper.onConnectionStatusChanged = { status ->
            viewModelScope.launch {
                _mqttStatus.value = status
            }
        }
        reconnectMqtt()
    }

    fun reconnectMqtt() {
        viewModelScope.launch {
            val username = com.example.whisper_android.util.DeviceUtils.getDeviceId(getApplication())
            val pwdResult = com.example.whisper_android.data.di.NetworkModule.repository.fetchMqttPassword(username)
            if (pwdResult.isSuccess) {
                mqttHelper.connect(pwdResult.getOrNull()!!)
            }
        }
    }

    private fun loadSummaries() {
        // Mock data - in real app this would load from assets/summaries.md
        val idSummary =
            """
# Ringkasan Pertemuan
**Topik Utama:** Evaluasi Q3 dan Perencanaan Q4.

### Poin Penting:
- Pertumbuhan pasar mencapai 15% di kuartal ini.
- Alokasi anggaran baru sudah disetujui.
- Perlu fokus pada kemitraan strategis bulan depan.

### Action Items:
1. Kirim dokumen anggaran ke tim finance.
2. Jadwalkan meeting dengan partner eksternal.
            """.trimIndent()

        val enSummary =
            """
# Meeting Summary
**Main Topic:** Q3 Performance Review & Q4 Planning.

### Key Highlights:
- Market share growth reached 15% this quarter.
- New budget allocation has been approved.
- Strategic partnerships need focus next month.

### Action Items:
1. Send budget documents to the finance team.
2. Schedule meeting with external partners.
            """.trimIndent()

        _summaries.value = SummariesData(idSummary = idSummary, enSummary = enSummary)
    }

    fun selectLanguage(lang: String) {
        _selectedLanguage.value = lang
    }

    private val _emailState = MutableStateFlow<UiState<Boolean>>(UiState.Idle)
    val emailState: StateFlow<UiState<Boolean>> = _emailState

    fun sendEmail(
        email: String,
        subject: String
    ) {
        val token =
            com.example.whisper_android.data.di.NetworkModule.tokenManager
                .getAccessToken() ?: ""

        if (token.isEmpty()) {
            _emailState.value = UiState.Error("Authentication token not found. Please login again.")
            return
        }

        val recipients = email.split(",").map { it.trim() }.filter { it.isNotEmpty() }
        if (recipients.isEmpty()) {
            _emailState.value = UiState.Error("At least one valid email address is required.")
            return
        }

        viewModelScope.launch {
            val sendEmailUseCase =
                com.example.whisper_android.data.di.NetworkModule.sendEmailUseCase

            sendEmailUseCase(
                to = recipients,
                subject = subject,
                template = "summary",
                token = token,
                attachmentPath = null // SummaryViewModel currently doesn't track PDF URL
            ).collectLatest { resource ->
                when (resource) {
                    is Resource.Loading -> {
                        _emailState.value = UiState.Loading
                    }

                    is Resource.Success -> {
                        _emailState.value = UiState.Success(true)
                    }

                    is Resource.Error -> {
                        _emailState.value = UiState.Error(
                            resource.message ?: "Failed to send email"
                        )
                    }
                }
            }
        }
    }

    fun sendEmailByMac(
        macAddress: String,
        subject: String
    ) {
        val token =
            com.example.whisper_android.data.di.NetworkModule.tokenManager
                .getAccessToken() ?: ""

        if (token.isEmpty()) {
            _emailState.value = UiState.Error("Authentication token not found. Please login again.")
            return
        }

        val overrideEmails = if (macAddress.contains("@")) {
            macAddress.split(",").map { it.trim() }.filter { it.isNotEmpty() }
        } else {
            null
        }

        viewModelScope.launch {
            val sendEmailByMacUseCase =
                com.example.whisper_android.data.di.NetworkModule.sendEmailByMacUseCase

            sendEmailByMacUseCase(
                // If it's email, MAC lookup is irrelevant but usecase needs a value or override
                macAddress = if (overrideEmails != null) "" else macAddress,
                subject = subject,
                template = "summary",
                token = token,
                attachmentPath = null,
                overrideEmails = overrideEmails
            ).collectLatest { resource ->
                when (resource) {
                    is Resource.Loading -> {
                        _emailState.value = UiState.Loading
                    }

                    is Resource.Success -> {
                        _emailState.value = UiState.Success(true)
                    }

                    is Resource.Error -> {
                        _emailState.value = UiState.Error(
                            resource.message ?: "Failed to send email by MAC"
                        )
                    }
                }
            }
        }
    }

    fun resetEmailState() {
        _emailState.value = UiState.Idle
    }
}
