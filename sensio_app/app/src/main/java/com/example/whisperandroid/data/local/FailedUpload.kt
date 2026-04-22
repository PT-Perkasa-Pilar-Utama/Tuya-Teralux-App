package com.example.whisperandroid.data.local

data class FailedUpload(
    val localFilePath: String,
    val bookingId: String,
    val createdAt: Long = System.currentTimeMillis(),
    val retryCount: Int = 0,
    val lastError: String = ""
) {
    fun withIncrementedRetry(): FailedUpload = copy(retryCount = retryCount + 1)

    fun withError(error: String): FailedUpload = copy(lastError = error)
}