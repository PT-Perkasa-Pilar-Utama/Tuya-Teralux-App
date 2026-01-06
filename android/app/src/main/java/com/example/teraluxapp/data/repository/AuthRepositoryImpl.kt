package com.example.teraluxapp.data.repository

import com.example.teraluxapp.data.model.AuthResponse
import com.example.teraluxapp.data.network.ApiService
import com.example.teraluxapp.utils.getErrorMessage
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class AuthRepositoryImpl @Inject constructor(
    private val apiService: ApiService
) : AuthRepository {
    
    override suspend fun authenticate(): Result<AuthResponse> {
        return try {
            val response = apiService.authenticate()
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
            Result.failure(Exception(e.message ?: "An error occurred"))
        }
    }
}
