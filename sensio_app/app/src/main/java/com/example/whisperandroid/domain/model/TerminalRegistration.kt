package com.example.whisperandroid.domain.model

data class TerminalRegistration(
    val id: String,
    val name: String,
    val roomId: String,
    val macAddress: String,
    val deviceTypeId: String? = null,
    val aiProvider: String? = null,
    val mqttUsername: String? = null,
    val mqttPassword: String? = null
)
