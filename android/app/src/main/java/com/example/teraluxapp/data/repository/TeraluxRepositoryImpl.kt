package com.example.teraluxapp.data.repository

import com.example.teraluxapp.data.model.CreateTeraluxRequest
import com.example.teraluxapp.data.model.CreateTeraluxResponse
import com.example.teraluxapp.data.model.TeraluxResponseDTO
import com.example.teraluxapp.data.network.ApiService
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class TeraluxRepositoryImpl @Inject constructor(
    private val apiService: ApiService
) : TeraluxRepository {
    
    override suspend fun checkDeviceRegistration(macAddress: String): Result<TeraluxResponseDTO?> {
        return try {
            val response = apiService.getTeraluxByMAC(macAddress)
            if (response.status) {
                Result.success(response.data)
            } else {
                Result.success(null) // Not registered
            }
        } catch (e: Exception) {
            // Treat 404 or any error as not registered
            Result.success(null)
        }
    }
    
    override suspend fun registerDevice(macAddress: String, roomId: String, name: String): Result<CreateTeraluxResponse> {
        return try {
            val response = apiService.registerTeralux(CreateTeraluxRequest(macAddress, roomId, name))
            if (response.status && response.data != null) {
                Result.success(response.data)
            } else {
                Result.failure(Exception(response.message))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}
