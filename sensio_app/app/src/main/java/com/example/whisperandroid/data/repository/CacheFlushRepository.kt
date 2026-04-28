package com.example.whisperandroid.data.repository

interface CacheFlushRepository {
    suspend fun flushCache(): Result<Boolean>
}