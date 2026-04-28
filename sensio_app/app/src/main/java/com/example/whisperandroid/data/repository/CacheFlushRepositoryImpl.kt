package com.example.whisperandroid.data.repository

import com.example.whisperandroid.data.local.TokenManager
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody

class CacheFlushRepositoryImpl(
    private val tokenManager: TokenManager,
    private val baseUrl: String,
    private val okhttpClient: okhttp3.OkHttpClient
) : CacheFlushRepository {

    override suspend fun flushCache(): Result<Boolean> = withContext(Dispatchers.IO) {
        try {
            val token = tokenManager.getAccessToken()
                ?: return@withContext Result.failure(Exception("No access token found"))

            val request = Request.Builder()
                .url("${baseUrl.trimEnd('/')}/api/cache/flush")
                .delete("".toRequestBody("application/json".toMediaType()))
                .addHeader("Authorization", "Bearer $token")
                .build()

            okhttpClient.newCall(request).execute().use { response ->
                if (response.isSuccessful) {
                    Result.success(true)
                } else {
                    Result.failure(Exception("HTTP ${response.code}: ${response.message}"))
                }
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}