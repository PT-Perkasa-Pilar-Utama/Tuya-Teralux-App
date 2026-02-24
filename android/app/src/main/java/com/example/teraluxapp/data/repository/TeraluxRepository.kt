package com.example.teraluxapp.data.repository

import com.example.teraluxapp.data.model.CreateTeraluxResponse
import com.example.teraluxapp.data.model.Teralux

interface TeraluxRepository {
    suspend fun checkDeviceRegistration(macAddress: String): Result<Teralux?>
    suspend fun registerDevice(macAddress: String, roomId: String, name: String): Result<CreateTeraluxResponse>
}
