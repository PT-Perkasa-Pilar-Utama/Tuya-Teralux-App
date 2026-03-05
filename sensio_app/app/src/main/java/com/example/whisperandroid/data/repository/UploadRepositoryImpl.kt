package com.example.whisperandroid.data.repository

import android.util.Log
import com.example.whisperandroid.data.remote.api.SpeechApi
import com.example.whisperandroid.data.remote.dto.CreateUploadSessionRequestDto
import com.example.whisperandroid.data.remote.dto.UploadSessionResponseDto
import com.example.whisperandroid.domain.repository.Resource
import com.example.whisperandroid.domain.repository.UploadRepository
import com.example.whisperandroid.domain.repository.UploadState
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow
import okhttp3.MediaType.Companion.toMediaTypeOrNull
import okhttp3.RequestBody.Companion.toRequestBody
import java.io.File
import java.io.RandomAccessFile

class UploadRepositoryImpl(
    private val speechApi: SpeechApi
) : UploadRepository {

    override fun uploadFile(
        file: File,
        token: String,
        chunkSizeMb: Int
    ): Flow<UploadState> = flow {
        emit(UploadState.Loading("Creating upload session..."))
        
        val totalSize = file.length()
        val chunkSize = chunkSizeMb * 1024 * 1024
        
        // 1. Create Session
        val sessionResponse = try {
            speechApi.createUploadSession(
                CreateUploadSessionRequestDto(
                    fileName = file.name,
                    totalSizeBytes = totalSize,
                    chunkSizeByes = chunkSize
                ),
                "Bearer $token"
            )
        } catch (e: Exception) {
            emit(UploadState.Error("Failed to create session: ${e.message}"))
            return@flow
        }
        
        if (!sessionResponse.status || sessionResponse.data == null) {
            emit(UploadState.Error("Failed to create session: ${sessionResponse.message}"))
            return@flow
        }
        
        val sessionId = sessionResponse.data.sessionId
        val totalChunks = sessionResponse.data.totalChunks
        
        // 2. Upload Chunks (sequentially for simplicity and stability)
        emit(UploadState.Progress(0, totalSize, 0f))
        
        val raf = RandomAccessFile(file, "r")
        try {
            for (i in 0 until totalChunks) {
                // Check if chunk is missing (can be optimized with missingRanges from response)
                // For now, simple sequential
                
                val offset = i.toLong() * chunkSize
                val remaining = totalSize - offset
                val currentChunkSize = if (remaining < chunkSize) remaining else chunkSize.toLong()
                
                val buffer = ByteArray(currentChunkSize.toInt())
                raf.seek(offset)
                raf.readFully(buffer)
                
                val requestBody = buffer.toRequestBody("application/octet-stream".toMediaTypeOrNull())
                
                val ackResponse = speechApi.uploadChunk(
                    sessionId = sessionId,
                    chunkIndex = i,
                    chunk = requestBody,
                    token = "Bearer $token"
                )
                
                if (!ackResponse.status) {
                    emit(UploadState.Error("Failed to upload chunk $i: ${ackResponse.message}"))
                    return@flow
                }
                
                val uploaded = (i + 1).toLong() * chunkSize
                val progress = if (uploaded > totalSize) totalSize else uploaded
                emit(UploadState.Progress(progress, totalSize, (progress.toFloat() / totalSize) * 100))
            }
            
            emit(UploadState.Success(sessionId))
        } catch (e: Exception) {
            emit(UploadState.Error("Upload failed: ${e.message}"))
        } finally {
            raf.close()
        }
    }

    override suspend fun getSessionStatus(
        sessionId: String,
        token: String
    ): Resource<UploadSessionResponseDto> {
        return try {
            val response = speechApi.getUploadSessionStatus(sessionId, "Bearer $token")
            if (response.status && response.data != null) {
                Resource.Success(response.data)
            } else {
                Resource.Error(response.message)
            }
        } catch (e: Exception) {
            Resource.Error(e.message ?: "Unknown error")
        }
    }
}
