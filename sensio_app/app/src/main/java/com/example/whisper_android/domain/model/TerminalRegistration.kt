package com.example.whisper_android.domain.model

data class TerminalRegistration(
    val id: String,
    val name: String,
    val roomId: String,
    val macAddress: String,
    val mqttUsername: String? = null,
    val mqttPassword: String? = null
)
