package com.example.teraluxapp.data.repository

import com.example.teraluxapp.data.model.AuthResponse
import com.example.teraluxapp.data.network.ApiService
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class AuthRepositoryImpl @Inject constructor(
    private val apiService: ApiService
) : AuthRepository {
    
    override suspend fun authenticate(): Result<AuthResponse> {
        return try {
            val response = apiService.authenticate()
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
