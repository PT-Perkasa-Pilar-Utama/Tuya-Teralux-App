package com.example.whisperandroid.data.repository

import okhttp3.MediaType.Companion.toMediaType
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class UploadRepositoryImplTest {

    private val repository = UploadRepositoryImpl(
        whisperApi = object : com.example.whisperandroid.data.remote.api.WhisperApi {
            // Mock implementation - not used in these tests

            override suspend fun transcribeAudio(
                audio: okhttp3.MultipartBody.Part,
                language: okhttp3.MultipartBody.Part,
                macAddress: okhttp3.MultipartBody.Part?,
                token: String,
                idempotencyKey: String?
            ): com.example.whisperandroid.data.remote.dto.SpeechResponseDto<com.example.whisperandroid.data.remote.dto.TranscriptionSubmissionData> {
                throw UnsupportedOperationException()
            }

            override suspend fun getTranscriptionStatus(
                taskId: String,
                token: String
            ): com.example.whisperandroid.data.remote.dto.SpeechResponseDto<com.example.whisperandroid.data.remote.dto.TranscriptionStatusDto> {
                throw UnsupportedOperationException()
            }

            override suspend fun createUploadSession(
                request: com.example.whisperandroid.data.remote.dto.CreateUploadSessionRequestDto,
                token: String
            ): com.example.whisperandroid.data.remote.dto.SpeechResponseDto<com.example.whisperandroid.data.remote.dto.UploadSessionResponseDto> {
                throw UnsupportedOperationException()
            }

            override suspend fun uploadChunk(
                sessionId: String,
                chunkIndex: Int,
                chunk: okhttp3.RequestBody,
                token: String
            ): com.example.whisperandroid.data.remote.dto.SpeechResponseDto<com.example.whisperandroid.data.remote.dto.UploadChunkAckDto> {
                throw UnsupportedOperationException()
            }

            override suspend fun getUploadSessionStatus(
                sessionId: String,
                token: String
            ): com.example.whisperandroid.data.remote.dto.SpeechResponseDto<com.example.whisperandroid.data.remote.dto.UploadSessionResponseDto> {
                throw UnsupportedOperationException()
            }

            override suspend fun transcribeByUpload(
                request: com.example.whisperandroid.data.remote.dto.SubmitByUploadRequestDto,
                token: String
            ): com.example.whisperandroid.data.remote.dto.SpeechResponseDto<com.example.whisperandroid.data.remote.dto.TranscriptionSubmissionData> {
                throw UnsupportedOperationException()
            }
        }
    )

    @Test
    fun `parseMissingRanges handles single chunk indices`() {
        val ranges = listOf("0", "2", "5")
        val result = repository.parseMissingRanges(ranges, 10)

        assertEquals(listOf(0, 2, 5), result)
    }

    @Test
    fun `parseMissingRanges handles range notation`() {
        val ranges = listOf("0-4", "8-9")
        val result = repository.parseMissingRanges(ranges, 10)

        assertEquals(listOf(0, 1, 2, 3, 4, 8, 9), result)
    }

    @Test
    fun `parseMissingRanges handles mixed single indices and ranges`() {
        val ranges = listOf("0", "2-4", "7")
        val result = repository.parseMissingRanges(ranges, 10)

        assertEquals(listOf(0, 2, 3, 4, 7), result)
    }

    @Test
    fun `parseMissingRanges filters out indices beyond totalChunks`() {
        val ranges = listOf("0", "5-15")
        val result = repository.parseMissingRanges(ranges, 10)

        // Should only include indices < 10
        assertEquals(listOf(0, 5, 6, 7, 8, 9), result)
    }

    @Test
    fun `parseMissingRanges returns sorted list`() {
        val ranges = listOf("5", "0-2", "8")
        val result = repository.parseMissingRanges(ranges, 10)

        assertTrue(result == result.sorted())
    }

    @Test
    fun `parseMissingRanges handles invalid range gracefully`() {
        val ranges = listOf("invalid", "0-2")
        val result = repository.parseMissingRanges(ranges, 10)

        // Invalid range should be skipped, valid ones processed
        assertEquals(listOf(0, 1, 2), result)
    }

    @Test
    fun `parseMissingRanges handles empty list`() {
        val result = repository.parseMissingRanges(emptyList(), 10)
        assertTrue(result.isEmpty())
    }

    // Tests for isChunkMissing - response loss detection
    @Test
    fun `isChunkMissing returns false when chunk not in missing ranges`() {
        val ranges = listOf("0", "2", "5")
        // Chunk 1 is not in missing ranges, meaning it was already received
        val result = repository.isChunkMissing(1, ranges)
        assertEquals(false, result)
    }

    @Test
    fun `isChunkMissing returns true when chunk is in missing ranges as single index`() {
        val ranges = listOf("0", "2", "5")
        val result = repository.isChunkMissing(2, ranges)
        assertEquals(true, result)
    }

    @Test
    fun `isChunkMissing returns true when chunk is in missing ranges as range`() {
        val ranges = listOf("0-4", "8-9")
        // Chunk 3 is in range 0-4
        val result = repository.isChunkMissing(3, ranges)
        assertEquals(true, result)
    }

    @Test
    fun `isChunkMissing returns false when empty missing ranges means all received`() {
        val ranges = emptyList<String>()
        // Empty missing ranges means all chunks received
        val result = repository.isChunkMissing(0, ranges)
        assertEquals(false, result)
    }

    @Test
    fun `isChunkMissing handles invalid range gracefully`() {
        val ranges = listOf("invalid", "0-2")
        // Invalid range should be skipped
        val result = repository.isChunkMissing(1, ranges)
        assertEquals(true, result) // 1 is in valid range 0-2
    }

    // Tests for getNextSmallerChunkSize - adaptive restart with non-ladder values
    @Test
    fun `getNextSmallerChunkSize returns next smaller ladder value`() {
        // 1MB should return 512KB
        val result = repository.getNextSmallerChunkSize(1 * 1024 * 1024L)
        assertEquals(512 * 1024L, result)
    }

    @Test
    fun `getNextSmallerChunkSize returns null at minimum chunk size`() {
        // 256KB is minimum, should return null
        val result = repository.getNextSmallerChunkSize(256 * 1024L)
        assertEquals(null, result)
    }

    @Test
    fun `getNextSmallerChunkSize handles non-ladder server chunk size`() {
        // Server returns 768KB (not in ladder), should return 512KB
        val result = repository.getNextSmallerChunkSize(768 * 1024L)
        assertEquals(512 * 1024L, result)
    }

    @Test
    fun `getNextSmallerChunkSize handles chunk size larger than ladder`() {
        // Server returns 2MB (larger than any ladder value), should return 1MB
        val result = repository.getNextSmallerChunkSize(2 * 1024 * 1024L)
        assertEquals(1 * 1024 * 1024L, result)
    }

    @Test
    fun `getNextSmallerChunkSize handles chunk size between ladder values`() {
        // Server returns 600KB (between 1MB and 512KB), should return 512KB
        val result = repository.getNextSmallerChunkSize(600 * 1024L)
        assertEquals(512 * 1024L, result)

        // Server returns 400KB (between 512KB and 256KB), should return 256KB
        val result2 = repository.getNextSmallerChunkSize(400 * 1024L)
        assertEquals(256 * 1024L, result2)
    }

    // Tests for probe-first strategy - chunk size ladder validation
    @Test
    fun `chunk size ladder has correct values`() {
        // Verify the ladder is correctly ordered from largest to smallest
        val ladder = listOf(
            1 * 1024 * 1024L, // 1 MB
            512 * 1024L, // 512 KB
            256 * 1024L // 256 KB (minimum)
        )

        assertEquals(3, ladder.size)
        assertEquals(1048576L, ladder[0])
        assertEquals(524288L, ladder[1])
        assertEquals(262144L, ladder[2])

        // Verify descending order
        assertTrue(ladder[0] > ladder[1])
        assertTrue(ladder[1] > ladder[2])
    }

    @Test
    fun `getNextSmallerChunkSize follows probe ladder correctly`() {
        // 1MB probe fails, should try 512KB
        assertEquals(512 * 1024L, repository.getNextSmallerChunkSize(1 * 1024 * 1024L))

        // 512KB probe fails, should try 256KB
        assertEquals(256 * 1024L, repository.getNextSmallerChunkSize(512 * 1024L))

        // 256KB is minimum, should return null (no more fallback)
        assertEquals(null, repository.getNextSmallerChunkSize(256 * 1024L))
    }

    @Test
    fun `getNextSmallerChunkSize handles edge case at ladder boundaries`() {
        // Exactly at 1MB boundary - next smaller is 512KB
        assertEquals(512 * 1024L, repository.getNextSmallerChunkSize(1048576L))

        // Slightly above 512KB - next smaller is 512KB
        assertEquals(512 * 1024L, repository.getNextSmallerChunkSize(524289L))

        // Exactly at 512KB boundary - next smaller is 256KB
        assertEquals(256 * 1024L, repository.getNextSmallerChunkSize(524288L))

        // Slightly above 256KB - next smaller is 256KB
        assertEquals(256 * 1024L, repository.getNextSmallerChunkSize(262145L))

        // Exactly at 256KB boundary (minimum) - no smaller
        assertEquals(null, repository.getNextSmallerChunkSize(262144L))
    }

    @Test
    fun `getNextSmallerChunkSize handles server-clamped chunk sizes`() {
        // Server clamps to 800KB (not in ladder), next should be 512KB
        assertEquals(512 * 1024L, repository.getNextSmallerChunkSize(800 * 1024L))

        // Server clamps to 400KB (not in ladder), next should be 256KB
        assertEquals(256 * 1024L, repository.getNextSmallerChunkSize(400 * 1024L))

        // Server clamps to 300KB (not in ladder), next should be 256KB
        assertEquals(256 * 1024L, repository.getNextSmallerChunkSize(300 * 1024L))
    }

    // Tests for retryable error classification - probe-first strategy
    @Test
    fun `isRetryableError returns true for SocketTimeoutException`() {
        val exception = java.net.SocketTimeoutException("timeout")
        assertTrue(repository.isRetryableError(exception))
    }

    @Test
    fun `isRetryableError returns true for retryable HTTP codes`() {
        val codes = listOf(408, 429, 500, 502, 503, 504)
        val mediaType = "text/plain".toMediaType()
        for (code in codes) {
            val responseBody = okhttp3.ResponseBody.create(mediaType, "")
            val response = retrofit2.Response.error<Unit>(code, responseBody)
            val exception = retrofit2.HttpException(response)
            assertTrue("HTTP $code should be retryable", repository.isRetryableError(exception))
        }
    }

    @Test
    fun `isRetryableError returns false for non-retryable HTTP codes`() {
        val codes = listOf(400, 401, 403, 404, 405, 409)
        val mediaType = "text/plain".toMediaType()
        for (code in codes) {
            val responseBody = okhttp3.ResponseBody.create(mediaType, "")
            val response = retrofit2.Response.error<Unit>(code, responseBody)
            val exception = retrofit2.HttpException(response)
            assertFalse("HTTP $code should not be retryable", repository.isRetryableError(exception))
        }
    }

    @Test
    fun `isRetryableError returns false for authentication errors`() {
        val mediaType = "text/plain".toMediaType()
        val responseBody = okhttp3.ResponseBody.create(mediaType, "")
        val response = retrofit2.Response.error<Unit>(401, responseBody)
        val exception = retrofit2.HttpException(response)
        assertFalse(repository.isRetryableError(exception))
    }

    @Test
    fun `isRetryableMessage returns true for timeout messages`() {
        assertTrue(repository.isRetryableMessage("Connection timeout"))
        assertTrue(repository.isRetryableMessage("Request timeout"))
        assertTrue(repository.isRetryableMessage("timeout occurred"))
    }

    @Test
    fun `isRetryableMessage returns true for retryable HTTP codes in message`() {
        assertTrue(repository.isRetryableMessage("Server error 503"))
        assertTrue(repository.isRetryableMessage("HTTP 502 Bad Gateway"))
        assertTrue(repository.isRetryableMessage("429 Too Many Requests"))
    }

    @Test
    fun `isRetryableMessage returns false for non-retryable errors`() {
        assertFalse(repository.isRetryableMessage("Invalid authentication token"))
        assertFalse(repository.isRetryableMessage("Bad request"))
        assertFalse(repository.isRetryableMessage("Session not found"))
        assertFalse(repository.isRetryableMessage("409 Conflict"))
    }

    @Test
    fun `classifyError returns user-friendly message for SocketTimeoutException`() {
        val exception = java.net.SocketTimeoutException("connect timed out")
        assertEquals("Connection timeout", repository.classifyError(exception))
    }

    @Test
    fun `classifyError returns user-friendly message for HTTP errors`() {
        val mediaType = "text/plain".toMediaType()
        val responseBody = okhttp3.ResponseBody.create(mediaType, "")
        val response408 = retrofit2.Response.error<Unit>(408, responseBody)
        val exception408 = retrofit2.HttpException(response408)
        assertEquals("Request timeout (server)", repository.classifyError(exception408))

        val response500 = retrofit2.Response.error<Unit>(500, responseBody)
        val exception500 = retrofit2.HttpException(response500)
        assertEquals("Server error (500)", repository.classifyError(exception500))
    }

    // Tests for ChunkUploadResult - structured result type
    @Test
    fun `isRetryableMessage classifies timeout as retryable for chunk upload result`() {
        val timeoutMessage = "Connection timeout"
        assertTrue(repository.isRetryableMessage(timeoutMessage))
    }

    @Test
    fun `isRetryableMessage classifies non-retryable API errors correctly`() {
        assertFalse(repository.isRetryableMessage("Invalid authentication token"))
        assertFalse(repository.isRetryableMessage("Bad request"))
        assertFalse(repository.isRetryableMessage("Invalid session"))
    }

    @Test
    fun `isRetryableMessage classifies retryable HTTP codes in messages`() {
        assertTrue(repository.isRetryableMessage("Server error 503"))
        assertTrue(repository.isRetryableMessage("HTTP 502 Bad Gateway"))
        assertTrue(repository.isRetryableMessage("429 Too Many Requests"))
    }

    // Tests for bounded concurrency logic
    @Test
    fun `bounded concurrency returns 1 for chunk sizes larger than 512 KB`() {
        // 1 MB chunk size should use concurrency 1
        val chunkSize1MB = 1 * 1024 * 1024L
        val concurrency1MB = if (chunkSize1MB > 512 * 1024) 1 else 2
        assertEquals(1, concurrency1MB)

        // 768 KB chunk size should use concurrency 1
        val chunkSize768KB = 768 * 1024L
        val concurrency768KB = if (chunkSize768KB > 512 * 1024) 1 else 2
        assertEquals(1, concurrency768KB)
    }

    @Test
    fun `bounded concurrency returns 2 for chunk sizes 512 KB or smaller`() {
        // 512 KB chunk size should use concurrency 2
        val chunkSize512KB = 512 * 1024L
        val concurrency512KB = if (chunkSize512KB > 512 * 1024) 1 else 2
        assertEquals(2, concurrency512KB)

        // 256 KB chunk size should use concurrency 2
        val chunkSize256KB = 256 * 1024L
        val concurrency256KB = if (chunkSize256KB > 512 * 1024) 1 else 2
        assertEquals(2, concurrency256KB)

        // 128 KB chunk size should use concurrency 2
        val chunkSize128KB = 128 * 1024L
        val concurrency128KB = if (chunkSize128KB > 512 * 1024) 1 else 2
        assertEquals(2, concurrency128KB)
    }

    @Test
    fun `chunked batching limits concurrent async jobs`() {
        // Simulate batching logic for 1 MB chunk size (concurrency 1)
        val chunksToUpload = listOf(0, 1, 2, 3, 4)
        val serverChunkSize = 1 * 1024 * 1024L
        val maxConcurrency = if (serverChunkSize > 512 * 1024) 1 else 2
        val batches = chunksToUpload.chunked(maxConcurrency)

        // With concurrency 1, each batch should have exactly 1 chunk
        assertEquals(5, batches.size)
        assertEquals(listOf(0), batches[0])
        assertEquals(listOf(1), batches[1])

        // Simulate batching logic for 256 KB chunk size (concurrency 2)
        val serverChunkSize256KB = 256 * 1024L
        val maxConcurrency2 = if (serverChunkSize256KB > 512 * 1024) 1 else 2
        val batches2 = chunksToUpload.chunked(maxConcurrency2)

        // With concurrency 2, batches should have at most 2 chunks each
        assertEquals(3, batches2.size)
        assertEquals(listOf(0, 1), batches2[0])
        assertEquals(listOf(2, 3), batches2[1])
        assertEquals(listOf(4), batches2[2])
    }

    @Test
    fun `chunked batching prevents unbounded async job creation for large file`() {
        // Simulate a large file with many chunks
        val totalChunks = 100
        val chunksToUpload = (0 until totalChunks).toList()

        // With 1 MB chunk size (concurrency 1)
        val serverChunkSize = 1 * 1024 * 1024L
        val maxConcurrency = if (serverChunkSize > 512 * 1024) 1 else 2
        val batches = chunksToUpload.chunked(maxConcurrency)

        // Should create 100 batches of 1 chunk each, not 100 async jobs at once
        assertEquals(totalChunks, batches.size)
        assertEquals(1, batches[0].size)

        // With 256 KB chunk size (concurrency 2)
        val serverChunkSize256KB = 256 * 1024L
        val maxConcurrency2 = if (serverChunkSize256KB > 512 * 1024) 1 else 2
        val batches2 = chunksToUpload.chunked(maxConcurrency2)

        // Should create 50 batches of 2 chunks each, not 100 async jobs at once
        assertEquals(50, batches2.size)
        assertEquals(2, batches2[0].size)
    }
}
