package com.example.whisper_android.domain.repository

import com.example.whisper_android.domain.model.TerminalRegistration

interface TerminalRepository {
    suspend fun registerTerminal(
        name: String,
        roomId: String,
        macAddress: String,
        deviceTypeId: String
    ): Result<TerminalRegistration>

    suspend fun getTerminalByMac(macAddress: String): Result<TerminalRegistration?>
}
