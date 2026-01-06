package com.example.teraluxapp.data.repository

import com.example.teraluxapp.data.model.AuthResponse
import com.example.teraluxapp.data.network.ApiService
import com.example.teraluxapp.utils.getErrorMessage
import retrofit2.HttpException
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
        } catch (e: HttpException) {
            // Extract error message from HTTP error response
            val errorMessage = e.response()?.getErrorMessage() ?: e.message()
            Result.failure(Exception(errorMessage))
        } catch (e: Exception) {
            Result.failure(Exception(e.message ?: "Terjadi kesalahan"))
        }
    }
}
