package com.example.whisper_android.data.repository

import com.example.whisper_android.data.remote.api.SpeechApi
import com.example.whisper_android.data.remote.dto.SpeechResponseDto
import com.example.whisper_android.data.remote.dto.TranscriptionStatusData
import com.example.whisper_android.data.remote.dto.TranscriptionSubmissionData
import okhttp3.MediaType.Companion.toMediaTypeOrNull
import okhttp3.MultipartBody
import okhttp3.RequestBody
import okio.BufferedSink
import okio.source
import java.io.File

import com.example.whisper_android.data.remote.api.RAGApi
import com.example.whisper_android.data.remote.dto.*

/**
 * Repository for handling speech transcription and RAG operations.
 */
class SpeechRepository(
    private val api: com.example.whisper_android.data.remote.api.SpeechApi,
    private val ragApi: RAGApi
) {

    /**
     * Uploads an audio file using chunked (streaming) multipart request.
     */
    suspend fun transcribeAudio(file: File, token: String): Result<TranscriptionSubmissionData> {
        return try {
            val authToken = if (token.startsWith("Bearer ")) token else "Bearer $token"
            val requestFile = StreamRequestBody(file)
            val body = MultipartBody.Part.createFormData("audio", file.name, requestFile)
            
            val response = api.transcribeAudio(body, authToken)
            if (response.status && response.data != null) {
                Result.success(response.data)
            } else {
                Result.failure(Exception(response.message))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    /**
     * Checks the status of a transcription task.
     */
    suspend fun getStatus(taskId: String, token: String): Result<TranscriptionStatusData> {
        return try {
            val authToken = if (token.startsWith("Bearer ")) token else "Bearer $token"
            val response = api.getTranscriptionStatus(taskId, authToken)
            if (response.status && response.data != null) {
                Result.success(response.data)
            } else {
                Result.failure(Exception(response.message))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    /**
     * Translates text using RAG API.
     */
    suspend fun translate(text: String, targetLang: String, token: String): Result<String> {
        return try {
            val authToken = if (token.startsWith("Bearer ")) token else "Bearer $token"
            val response = ragApi.translate(RAGRequestDto(text, targetLang), authToken)
            if (response.status && response.data != null) {
                Result.success(response.data)
            } else {
                Result.failure(Exception(response.message))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    /**
     * Translates text using RAG API (Asynchronous).
     */
    suspend fun translateAsync(text: String, targetLang: String, token: String): Result<TranscriptionSubmissionData> {
        return try {
            val authToken = if (token.startsWith("Bearer ")) token else "Bearer $token"
            val response = ragApi.translateAsync(RAGRequestDto(text, targetLang), authToken)
            if (response.status && response.data != null) {
                Result.success(response.data)
            } else {
                Result.failure(Exception(response.message))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    /**
     * Generates a summary using RAG API (Asynchronous).
     */
    suspend fun summaryAsync(request: RAGSummaryRequestDto, token: String): Result<TranscriptionSubmissionData> {
        return try {
            val authToken = if (token.startsWith("Bearer ")) token else "Bearer $token"
            val response = ragApi.summaryAsync(request, authToken)
            if (response.status && response.data != null) {
                Result.success(response.data)
            } else {
                Result.failure(Exception(response.message))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    /**
     * Checks the status of a RAG task.
     */
    suspend fun getRagStatus(taskId: String, token: String): Result<RAGStatusDto> {
        return try {
            val authToken = if (token.startsWith("Bearer ")) token else "Bearer $token"
            val response = ragApi.getStatus(taskId, authToken)
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

/**
 * Custom RequestBody to stream file content in chunks to avoid loading entire file into memory.
 */
class StreamRequestBody(private val file: File) : RequestBody() {
    override fun contentType() = "audio/mpeg".toMediaTypeOrNull()
    override fun contentLength() = file.length()

    override fun writeTo(sink: BufferedSink) {
        file.source().use { source ->
            sink.writeAll(source)
        }
    }
}
