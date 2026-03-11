package com.example.whisperandroid.data.repository

import android.util.Log
import com.example.whisperandroid.data.remote.api.WhisperApi
import com.example.whisperandroid.data.remote.dto.CreateUploadSessionRequestDto
import com.example.whisperandroid.data.remote.dto.UploadSessionResponseDto
import com.example.whisperandroid.domain.repository.Resource
import com.example.whisperandroid.domain.repository.UploadRepository
import com.example.whisperandroid.domain.repository.UploadState
import java.io.File
import java.io.RandomAccessFile
import kotlinx.coroutines.async
import kotlinx.coroutines.awaitAll
import kotlinx.coroutines.coroutineScope
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import okhttp3.MediaType.Companion.toMediaTypeOrNull
import okhttp3.RequestBody.Companion.toRequestBody

class UploadRepositoryImpl(
    private val whisperApi: WhisperApi
) : UploadRepository {

    override fun uploadFile(
        file: File,
        token: String,
        chunkSizeMb: Int,
        sessionId: String?
    ): Flow<UploadState> = flow {
        emit(UploadState.Loading("Initializing upload..."))

        val totalSize = file.length()
        val chunkSize = if (chunkSizeMb > 0) {
            chunkSizeMb * 1024 * 1024L
        } else {
            8 * 1024 * 1024L
        } // Default 8MB

        var currentSessionId: String? = sessionId
        var chunksToUpload: List<Int>? = null
        var uploadedBytes = 0L

        // 1. Try to Resume if sessionId provided
        if (currentSessionId != null) {
            try {
                val status = whisperApi.getUploadSessionStatus(currentSessionId, "Bearer $token")
                if (status.status && status.data != null) {
                    val data = status.data
                    chunksToUpload = if (data.missingRanges.isNullOrEmpty()) {
                        emptyList()
                    } else {
                        parseMissingRanges(data.missingRanges, data.totalChunks)
                    }
                    uploadedBytes = data.receivedBytes ?: 0L
                    Log.d(
                        "UploadRepo",
                        "Resuming session $currentSessionId. Missing chunks: ${chunksToUpload.size}"
                    )
                } else {
                    Log.w(
                        "UploadRepo",
                        "Resume session $currentSessionId failed: ${status.message}. " +
                            "Starting new session."
                    )
                    currentSessionId = null
                }
            } catch (e: Exception) {
                Log.w(
                    "UploadRepo",
                    "Error checking resume session: ${e.message}. Starting new session."
                )
                currentSessionId = null
            }
        }

        // 2. Create New Session if not resuming
        if (currentSessionId == null) {
            val sessionResponse = try {
                whisperApi.createUploadSession(
                    CreateUploadSessionRequestDto(
                        fileName = file.name,
                        totalSizeBytes = totalSize,
                        chunkSizeByes = chunkSize.toInt()
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

            currentSessionId = sessionResponse.data.sessionId
            val totalChunks = sessionResponse.data.totalChunks
            chunksToUpload = (0 until totalChunks).toList()
            uploadedBytes = 0L
        }

        emit(UploadState.SessionStarted(currentSessionId!!))

        if (chunksToUpload!!.isEmpty()) {
            emit(UploadState.Success(currentSessionId!!))
            return@flow
        }

        // 3. Upload Chunks concurrently (max 3 at a time)
        emit(
            UploadState.Progress(
                uploadedBytes,
                totalSize,
                (uploadedBytes.toFloat() / totalSize) * 100
            )
        )

        val raf = RandomAccessFile(file, "r")
        val mutex = Mutex()
        val maxConcurrency = 3

        try {
            coroutineScope {
                val chunkGroups = chunksToUpload.chunked(maxConcurrency)

                for (group in chunkGroups) {
                    val deferreds = group.map { i ->
                        async {
                            val offset = i.toLong() * chunkSize
                            val remaining = totalSize - offset
                            val currentChunkSize = if (remaining < chunkSize) {
                                remaining
                            } else {
                                chunkSize
                            }

                            val buffer = ByteArray(currentChunkSize.toInt())

                            mutex.withLock {
                                raf.seek(offset)
                                raf.readFully(buffer)
                            }

                            val requestBody = buffer.toRequestBody(
                                "application/octet-stream".toMediaTypeOrNull()
                            )

                            var success = false
                            var lastError = ""
                            var retryDelay = 1000L

                            for (retry in 0..3) {
                                try {
                                    val ackResponse = whisperApi.uploadChunk(
                                        sessionId = currentSessionId!!,
                                        chunkIndex = i,
                                        chunk = requestBody,
                                        token = "Bearer $token"
                                    )
                                    if (ackResponse.status) {
                                        success = true
                                        break
                                    } else {
                                        lastError = ackResponse.message ?: "Unknown API Error"
                                    }
                                } catch (e: Exception) {
                                    lastError = e.message ?: "Network error"
                                }
                                if (!success && retry < 3) {
                                    delay(retryDelay)
                                    retryDelay *= 2
                                }
                            }

                            if (!success) {
                                throw Exception(
                                    "Failed to upload chunk $i after 3 retries: $lastError"
                                )
                            }

                            mutex.withLock {
                                uploadedBytes += currentChunkSize
                            }
                        }
                    }

                    // Await group completion
                    deferreds.awaitAll()

                    val progress = if (uploadedBytes > totalSize) totalSize else uploadedBytes
                    emit(
                        UploadState.Progress(
                            progress,
                            totalSize,
                            (progress.toFloat() / totalSize) * 100
                        )
                    )
                }
            }

            emit(UploadState.Success(currentSessionId!!))
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
            val response = whisperApi.getUploadSessionStatus(sessionId, "Bearer $token")
            if (response.status && response.data != null) {
                Resource.Success(response.data)
            } else {
                Resource.Error(response.message)
            }
        } catch (e: Exception) {
            Resource.Error(e.message ?: "Unknown error")
        }
    }

    @androidx.annotation.VisibleForTesting
    internal fun parseMissingRanges(ranges: List<String>, totalChunks: Int): List<Int> {
        val missingIndices = mutableSetOf<Int>()
        ranges.forEach { range ->
            try {
                if (range.contains("-")) {
                    val parts = range.split("-")
                    if (parts.size == 2) {
                        val start = parts[0].toInt()
                        val end = parts[1].toInt()
                        for (i in start..end) {
                            if (i < totalChunks) missingIndices.add(i)
                        }
                    }
                } else {
                    val idx = range.toInt()
                    if (idx < totalChunks) missingIndices.add(idx)
                }
            } catch (e: Exception) {
                Log.w("UploadRepo", "Failed to parse range: $range")
            }
        }
        return missingIndices.toList().sorted()
    }
}
