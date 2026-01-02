package com.example.teraluxapp.data.repository

import com.example.teraluxapp.data.model.AuthResponse

interface AuthRepository {
    suspend fun authenticate(): Result<AuthResponse>
}
