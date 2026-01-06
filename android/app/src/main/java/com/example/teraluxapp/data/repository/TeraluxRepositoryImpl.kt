package com.example.teraluxapp.data.repository

import com.example.teraluxapp.data.model.CreateTeraluxRequest
import com.example.teraluxapp.data.model.CreateTeraluxResponse
import com.example.teraluxapp.data.model.Teralux
import com.example.teraluxapp.data.model.TeraluxResponseDTO
import com.example.teraluxapp.data.network.ApiService
import com.example.teraluxapp.utils.getErrorMessage
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class TeraluxRepositoryImpl @Inject constructor(
    private val apiService: ApiService
) : TeraluxRepository {
    
    override suspend fun checkDeviceRegistration(macAddress: String): Result<Teralux?> {
        return try {
            val response = apiService.getTeraluxByMAC(macAddress)
            if (response.status && response.data != null) {
                Result.success(response.data.teralux)
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
            if (response.isSuccessful && response.body() != null) {
                val body = response.body()!!
                if (body.status && body.data != null) {
                    Result.success(body.data)
                } else {
                    Result.failure(Exception(body.message))
                }
            } else {
                // Extract error message from error response body
                val errorMessage = response.getErrorMessage()
                Result.failure(Exception(errorMessage))
            }
        } catch (e: Exception) {
            Result.failure(Exception(e.message ?: "Terjadi kesalahan"))
        }
    }
}
