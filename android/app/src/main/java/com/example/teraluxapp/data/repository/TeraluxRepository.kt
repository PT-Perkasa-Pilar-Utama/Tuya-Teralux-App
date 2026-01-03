package com.example.teraluxapp.data.repository

import com.example.teraluxapp.data.model.BaseResponse
import com.example.teraluxapp.data.model.CreateTeraluxRequest
import com.example.teraluxapp.data.model.CreateTeraluxResponse
import com.example.teraluxapp.data.model.TeraluxResponseDTO

interface TeraluxRepository {
    suspend fun checkDeviceRegistration(macAddress: String): Result<TeraluxResponseDTO?>
    suspend fun registerDevice(macAddress: String, roomId: String, name: String): Result<CreateTeraluxResponse>
}
