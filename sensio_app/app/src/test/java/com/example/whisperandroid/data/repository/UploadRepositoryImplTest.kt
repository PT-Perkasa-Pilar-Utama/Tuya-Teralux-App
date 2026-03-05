package com.example.whisperandroid.data.repository

import com.example.whisperandroid.data.remote.api.SpeechApi
import org.junit.Assert.assertEquals
import org.junit.Before
import org.junit.Test
import com.example.whisperandroid.data.remote.dto.*
import okhttp3.MultipartBody
import okhttp3.RequestBody

class UploadRepositoryImplTest {

    private lateinit var uploadRepositoryImpl: UploadRepositoryImpl
    private lateinit var speechApi: SpeechApi

    @Before
    fun setUp() {
        speechApi = object : SpeechApi {
            override suspend fun transcribeAudio(audio: MultipartBody.Part, language: MultipartBody.Part, macAddress: MultipartBody.Part?, token: String, idempotencyKey: String?, apiKey: String): SpeechResponseDto<TranscriptionSubmissionData> = TODO()
            override suspend fun getTranscriptionStatus(taskId: String, token: String, apiKey: String): SpeechResponseDto<TranscriptionStatusDto> = TODO()
            override suspend fun createUploadSession(request: CreateUploadSessionRequestDto, token: String, apiKey: String): SpeechResponseDto<UploadSessionResponseDto> = TODO()
            override suspend fun uploadChunk(sessionId: String, chunkIndex: Int, chunk: RequestBody, token: String, apiKey: String): SpeechResponseDto<UploadChunkAckDto> = TODO()
            override suspend fun getUploadSessionStatus(sessionId: String, token: String, apiKey: String): SpeechResponseDto<UploadSessionResponseDto> = TODO()
            override suspend fun transcribeByUpload(request: SubmitByUploadRequestDto, token: String, apiKey: String): SpeechResponseDto<TranscriptionSubmissionData> = TODO()
        }
        uploadRepositoryImpl = UploadRepositoryImpl(speechApi)
    }

    @Test
    fun parseMissingRanges_withSingleValues() {
        val ranges = listOf("1", "3", "5")
        val result = uploadRepositoryImpl.parseMissingRanges(ranges, 10)
        assertEquals(listOf(1, 3, 5), result)
    }

    @Test
    fun parseMissingRanges_withRanges() {
        val ranges = listOf("1-3", "5-6")
        val result = uploadRepositoryImpl.parseMissingRanges(ranges, 10)
        assertEquals(listOf(1, 2, 3, 5, 6), result)
    }

    @Test
    fun parseMissingRanges_outOfBounds() {
        val ranges = listOf("1-3", "8-12") // total chunks is 10
        val result = uploadRepositoryImpl.parseMissingRanges(ranges, 10)
        assertEquals(listOf(1, 2, 3, 8, 9), result) // should cap at totalChunks-1
    }

    @Test
    fun parseMissingRanges_mixed() {
        val ranges = listOf("0", "2-4", "7")
        val result = uploadRepositoryImpl.parseMissingRanges(ranges, 10)
        assertEquals(listOf(0, 2, 3, 4, 7), result)
    }

    @Test
    fun parseMissingRanges_empty() {
        val ranges = emptyList<String>()
        val result = uploadRepositoryImpl.parseMissingRanges(ranges, 10)
        assertEquals(emptyList<Int>(), result)
    }
}
