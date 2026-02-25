package com.example.whisper_android.data.repository

import com.example.whisper_android.common.util.getErrorMessage
import com.example.whisper_android.data.remote.api.TerminalApi
import com.example.whisper_android.data.remote.dto.TerminalRequestDto
import com.example.whisper_android.domain.model.TerminalRegistration
import com.example.whisper_android.domain.repository.TerminalRepository

class TerminalRepositoryImpl(
    private val api: TerminalApi,
    private val apiKey: String
) : TerminalRepository {
    override suspend fun registerTerminal(
        name: String,
        roomId: String,
        macAddress: String,
        deviceTypeId: String
    ): Result<TerminalRegistration> =
        try {
            val request = TerminalRequestDto(
                name = name,
                roomId = roomId,
                macAddress = macAddress,
                deviceTypeId = deviceTypeId
            )
            val response = api.registerTerminal(apiKey, request)

            if (response.status && response.data?.terminalId != null) {
                Result.success(
                    TerminalRegistration(
                        id = response.data.terminalId,
                        name = name,
                        roomId = roomId,
                        macAddress = macAddress
                    )
                )
            } else {
                Result.failure(Exception(response.message))
            }
        } catch (e: retrofit2.HttpException) {
            Result.failure(Exception(e.getErrorMessage()))
        } catch (e: Exception) {
            Result.failure(e)
        }

    override suspend fun getTerminalByMac(macAddress: String): Result<TerminalRegistration?> =
        try {
            val response = api.getTerminalByMac(apiKey, macAddress)

            if (response.status && response.data != null) {
                // Success - Found
                val terminalItem = response.data.terminal
                Result.success(
                    TerminalRegistration(
                        id = terminalItem?.id ?: response.data.id ?: "",
                        name = terminalItem?.name ?: response.data.name ?: "",
                        roomId = terminalItem?.roomId ?: response.data.roomId ?: "",
                        macAddress = terminalItem?.macAddress ?: response.data.macAddress ?: macAddress
                    )
                )
            } else {
                // 404 or other error - Not Found (or API specific error)
                // If API returns false status with "not found", we treat as null (not registered)
                if (response.message.contains("not found", ignoreCase = true)) {
                    Result.success(null)
                } else {
                    Result.failure(Exception(response.message))
                }
            }
        } catch (e: retrofit2.HttpException) {
            if (e.code() == 404) {
                Result.success(null) // Not found means not registered
            } else {
                Result.failure(Exception(e.getErrorMessage()))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
}
