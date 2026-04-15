package com.example.whisperandroid.domain.repository

import com.example.whisperandroid.domain.model.TerminalRegistration

data class AiEngineProfileState(
    val profile: String?,
    val source: String,
    val effectiveProvider: String?,
    val effectiveMode: String
)

interface TerminalRepository {
    suspend fun registerTerminal(
        name: String,
        roomId: String,
        macAddress: String,
        deviceTypeId: String
    ): Result<TerminalRegistration>

    suspend fun getTerminalByMac(macAddress: String): Result<TerminalRegistration?>

    suspend fun getAiEngineProfileByMac(macAddress: String): Result<AiEngineProfileState?>

    suspend fun fetchMqttPassword(username: String): Result<String>

    suspend fun updateTerminal(
        terminalId: String,
        aiProvider: String?
    ): Result<Unit>

    suspend fun updateAiEngineProfile(
        terminalId: String,
        profile: String?
    ): Result<Unit>
}
