package com.example.whisperandroid.domain.repository

import com.example.whisperandroid.domain.model.TerminalRegistration

interface TerminalRepository {
    suspend fun registerTerminal(
        name: String,
        roomId: String,
        macAddress: String,
        deviceTypeId: String
    ): Result<TerminalRegistration>

    suspend fun getTerminalByMac(macAddress: String): Result<TerminalRegistration?>

    suspend fun fetchMqttPassword(username: String): Result<String>
}
