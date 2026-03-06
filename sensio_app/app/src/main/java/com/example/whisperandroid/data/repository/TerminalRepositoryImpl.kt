package com.example.whisperandroid.data.repository

import com.example.whisperandroid.common.util.getErrorMessage
import com.example.whisperandroid.data.remote.api.TerminalApi
import com.example.whisperandroid.data.remote.dto.TerminalRequestDto
import com.example.whisperandroid.domain.model.TerminalRegistration
import com.example.whisperandroid.domain.repository.TerminalRepository

class TerminalRepositoryImpl(
    private val api: TerminalApi,
    private val apiKey: String,
    private val tokenManager: com.example.whisperandroid.data.local.TokenManager
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
                val tId = response.data.terminalId
                tokenManager.saveTerminalId(tId)

                Result.success(
                    TerminalRegistration(
                        id = tId,
                        name = name,
                        roomId = roomId,
                        macAddress = macAddress,
                        mqttUsername = response.data.mqttUsername,
                        mqttPassword = response.data.mqttPassword
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
                val tId = terminalItem?.id ?: response.data.id ?: ""
                if (tId.isNotEmpty()) {
                    tokenManager.saveTerminalId(tId)
                }

                // Save MQTT credentials if present
                val mUsername = terminalItem?.mqttUsername ?: response.data.mqttUsername
                val mPassword = terminalItem?.mqttPassword ?: response.data.mqttPassword

                Result.success(
                    TerminalRegistration(
                        id = tId,
                        name = terminalItem?.name ?: response.data.name ?: "",
                        roomId = terminalItem?.roomId ?: response.data.roomId ?: "",
                        macAddress = terminalItem?.macAddress ?: response.data.macAddress
                            ?: macAddress,
                        mqttUsername = mUsername,
                        mqttPassword = mPassword
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

    override suspend fun fetchMqttPassword(username: String): Result<String> =
        try {
            val response = api.getMqttCredentials(
                "Bearer ${tokenManager.getAccessToken()}",
                username
            )
            if (response.status && response.data != null) {
                Result.success(response.data.password)
            } else {
                Result.failure(Exception(response.message))
            }
        } catch (e: retrofit2.HttpException) {
            Result.failure(Exception(e.getErrorMessage()))
        } catch (e: Exception) {
            Result.failure(e)
        }
}
