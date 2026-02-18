package com.example.whisper_android.data.repository

import com.example.whisper_android.common.util.getErrorMessage
import com.example.whisper_android.data.remote.api.TeraluxApi
import com.example.whisper_android.data.remote.dto.TeraluxRequestDto
import com.example.whisper_android.domain.model.TeraluxRegistration
import com.example.whisper_android.domain.repository.TeraluxRepository

class TeraluxRepositoryImpl(
    private val api: TeraluxApi,
    private val apiKey: String
) : TeraluxRepository {
    override suspend fun registerTeralux(name: String, roomId: String, macAddress: String): Result<TeraluxRegistration> {
        return try {
            val request = TeraluxRequestDto(name = name, roomId = roomId, macAddress = macAddress)
            val response = api.registerTeralux(apiKey, request)
            
            if (response.status && response.data?.teraluxId != null) {
                Result.success(
                    TeraluxRegistration(
                        id = response.data.teraluxId,
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
    }

    override suspend fun getTeraluxByMac(macAddress: String): Result<TeraluxRegistration?> {
        return try {
            val response = api.getTeraluxByMac(apiKey, macAddress)
            
            if (response.status && response.data != null) {
                // Success - Found
                Result.success(
                    TeraluxRegistration(
                        id = response.data.id ?: "",
                        name = response.data.name ?: "",
                        roomId = response.data.roomId ?: "",
                        macAddress = response.data.macAddress ?: macAddress
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
}
