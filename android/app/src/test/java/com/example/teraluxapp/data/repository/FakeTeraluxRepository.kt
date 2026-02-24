package com.example.teraluxapp.data.repository

import com.example.teraluxapp.data.model.CreateTeraluxResponse
import com.example.teraluxapp.data.model.Teralux

// Shared fake implementation for testing
class FakeTeraluxRepository : TeraluxRepository {
    private var registeredDevices = mutableMapOf<String, Teralux>()
    private var shouldFail = false
    private var errorMessage: String? = null
    private var shouldReturnError = false
    
    fun setRegisteredDevice(macAddress: String) {
        registeredDevices[macAddress] = Teralux(
            id = "test-id",
            macAddress = macAddress,
            name = "Test Device",
            roomId = "101",
            createdAt = "2024-01-01",
            updatedAt = "2024-01-01"
        )
    }
    
    fun setError(message: String) {
        errorMessage = message
        shouldFail = true
    }
    
    fun setShouldFail(fail: Boolean) {
        shouldFail = fail
    }

    // Added for new registerDevice overload
    fun setShouldReturnError(error: Boolean) {
        shouldReturnError = error
    }
    
    override suspend fun checkDeviceRegistration(macAddress: String): Result<Teralux?> {
        if (shouldFail) {
            return Result.failure(Exception(errorMessage ?: "Unknown error"))
        }
        return Result.success(registeredDevices[macAddress])
    }
    
    override suspend fun registerDevice(macAddress: String, roomId: String, name: String): Result<CreateTeraluxResponse> {
        if (shouldReturnError || shouldFail) {
             return Result.failure(Exception(errorMessage ?: "Registration failed"))
        }
        registeredDevices[macAddress] = Teralux(
            id = "new-id",
            macAddress = macAddress,
            name = name,
            roomId = roomId,
            createdAt = "2024-01-01",
            updatedAt = "2024-01-01"
        )
        return Result.success(CreateTeraluxResponse("new-id"))
    }
}
