package com.example.whisper_android.domain.repository

import com.example.whisper_android.domain.model.TeraluxRegistration

interface TeraluxRepository {
    suspend fun registerTeralux(
        name: String,
        roomId: String,
        macAddress: String,
    ): Result<TeraluxRegistration>

    suspend fun getTeraluxByMac(macAddress: String): Result<TeraluxRegistration?>
}
