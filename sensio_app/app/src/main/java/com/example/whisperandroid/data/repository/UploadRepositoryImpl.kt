package com.example.whisperandroid.data.repository

import android.content.Context
import android.util.Log
import com.example.whisperandroid.data.remote.api.WhisperApi
import com.example.whisperandroid.utils.ZipEncryptor
import com.example.whisperandroid.data.remote.dto.CreateUploadIntentRequestDto
import com.example.whisperandroid.data.remote.dto.CreateUploadSessionRequestDto
import com.example.whisperandroid.data.remote.dto.SaveRecordingRequestDto
import com.example.whisperandroid.data.remote.dto.UploadIntentResponseDto
import com.example.whisperandroid.data.remote.dto.UploadSessionResponseDto
import com.example.whisperandroid.domain.repository.Resource
import com.example.whisperandroid.domain.repository.UploadRepository
import com.example.whisperandroid.domain.repository.UploadState
import java.io.File
import java.io.IOException
import java.io.RandomAccessFile
import java.net.SocketTimeoutException
import java.util.UUID
import kotlinx.coroutines.async
import kotlinx.coroutines.awaitAll
import kotlinx.coroutines.coroutineScope
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.FlowCollector
import kotlinx.coroutines.flow.flow
import kotlinx.coroutines.sync.Mutex
import kotlinx.coroutines.sync.withLock
import okhttp3.MediaType.Companion.toMediaTypeOrNull
import okhttp3.RequestBody.Companion.toRequestBody
import org.json.JSONObject
import org.bouncycastle.crypto.generators.Argon2BytesGenerator
import org.bouncycastle.crypto.params.Argon2Parameters
import retrofit2.HttpException

class UploadRepositoryImpl(
    private val whisperApi: WhisperApi,
    private val signedUploadEnabled: Boolean = false
) : UploadRepository {

    private val zipEncryptor = ZipEncryptor()

    companion object {
        private const val TAG = "UploadRepositoryImpl"

        // Chunk size ladder for probe-first strategy
        private val CHUNK_SIZE_LADDER = listOf(
            1 * 1024 * 1024L, // 1 MB
            512 * 1024L, // 512 KB
            256 * 1024L // 256 KB (minimum)
        )
        private const val MAX_RETRIES_PER_CHUNK = 3

        // Retryable HTTP status codes
        private val RETRYABLE_HTTP_CODES = setOf(408, 429, 500, 502, 503, 504)

        // S3 Multipart Upload constants
        private const val MULTIPART_MIN_PART_SIZE = 5 * 1024 * 1024L // 5MB minimum per S3 requirement
        private const val MAX_MULTIPART_RETRIES = 3
        private val MULTIPART_BACKOFF_DELAYS = listOf(1000L, 2000L, 4000L) // 1s, 2s, 4s exponential backoff
    }

    override suspend fun uploadSignedUrl(
        context: Context,
        audioFile: File,
        bookingId: String,
        token: String
    ): Flow<UploadState> = flow {
        emit(UploadState.Loading("Requesting signed upload URL..."))

        val password = UUID.randomUUID().toString()
        val zipFile: File
        try {
            zipFile = zipEncryptor.createEncryptedZip(context, audioFile.absolutePath, password)
        } catch (e: Exception) {
            emit(UploadState.Error("Failed to create encrypted ZIP: ${e.message}"))
            return@flow
        }

        try {
            val intentRequest = CreateUploadIntentRequestDto(
                filename = zipFile.name,
                size = zipFile.length(),
                contentType = "application/zip",
                bookingId = bookingId
            )

            val intentResponse = whisperApi.createUploadIntent(intentRequest, "Bearer $token")

            if (!intentResponse.status || intentResponse.data == null) {
                emit(UploadState.Error("Failed to get signed URL: ${intentResponse.message}"))
                return@flow
            }

            val intent = intentResponse.data
            emit(UploadState.Loading("Uploading to signed URL..."))

            val uploadResult = uploadToSignedUrl(
                file = zipFile,
                signedUrl = intent.presignedUrl,
                contentType = "application/zip",
                totalSize = zipFile.length()
            ) { progress, total ->
                val progressPercent = (progress * 100 / total).toFloat().coerceIn(0f, 100f)
                emit(UploadState.Progress(progress, total, progressPercent))
            }

            if (uploadResult) {
                val passwordHash = hashPasswordWithArgon2(password)
                saveRecordingMetadata(
                    objectKey = intent.objectKey,
                    bookingId = bookingId,
                    passwordHash = passwordHash,
                    token = token
                )
                emit(UploadState.Success(intent.objectKey))
            } else {
                emit(UploadState.Error("Signed URL upload failed"))
            }
        } finally {
            if (zipFile.exists()) {
                zipFile.delete()
            }
        }
    }

    override fun uploadFile(
        file: File,
        token: String,
        chunkSizeMb: Int,
        sessionId: String?
    ): Flow<UploadState> = flow {
        emit(UploadState.Loading("Initializing upload..."))

        // Signed URL upload path (alternative to chunk upload)
        if (signedUploadEnabled) {
            try {
                emit(UploadState.Loading("Requesting signed upload URL..."))

                val intentResponse = whisperApi.createUploadIntent(
                    CreateUploadIntentRequestDto("audio/wav"),
                    token
                )

                if (!intentResponse.status || intentResponse.data == null) {
                    emit(UploadState.Error("Failed to get signed URL: ${intentResponse.message}"))
                    return@flow
                }

                val intent = intentResponse.data
                emit(UploadState.Loading("Uploading to signed URL..."))

                val uploadResult = uploadToSignedUrl(
                    file = file,
                    signedUrl = intent.presignedUrl,
                    contentType = intent.contentType,
                    totalSize = file.length()
                ) { progress, total ->
                    val progressPercent = (progress * 100 / total).toFloat().coerceIn(0f, 100f)
                    emit(UploadState.Progress(progress, total, progressPercent))
                }

                if (uploadResult) {
                    emit(UploadState.Success(intent.objectKey))
                } else {
                    emit(UploadState.Error("Signed URL upload failed"))
                }
                return@flow
            } catch (e: Exception) {
                Log.w(TAG, "Signed upload failed, falling back to chunk upload: ${e.message}")
                // Fall through to chunk upload path
            }
        }

        // Original chunk upload path
        val totalSize = file.length()
        val requestedChunkSize = if (chunkSizeMb > 0) {
            chunkSizeMb * 1024 * 1024L
        } else {
            1 * 1024 * 1024L // Default 1MB initial request
        }

        // Phase 1: Resolve session plan (resume or probe-first)
        val sessionPlan = resolveSessionPlan(
            file = file,
            token = token,
            savedSessionId = sessionId,
            requestedChunkSize = requestedChunkSize
        )

        when (sessionPlan) {
            is SessionPlan.ResumeValidSession -> {
                // Resume existing valid session
                uploadOnExistingSession(
                    file = file,
                    sessionId = sessionPlan.sessionId,
                    serverChunkSize = sessionPlan.serverChunkSize,
                    chunksToUpload = sessionPlan.chunksToUpload,
                    uploadedBytes = sessionPlan.uploadedBytes,
                    totalSize = totalSize,
                    token = token
                )
            }
            is SessionPlan.ProbeFreshSession -> {
                // Probe-first: find optimal chunk size for current network conditions
                probeAndUpload(
                    file = file,
                    token = token,
                    totalSize = totalSize,
                    initialChunkSize = sessionPlan.initialChunkSize
                )
            }
            is SessionPlan.TerminalFailure -> {
                // Non-recoverable error (auth/ownership violation) - emit error immediately
                emit(UploadState.Error("Upload failed: ${sessionPlan.message}"))
            }
        }
    }

    /**
     * Session plan decision result
     *
     * Session states and resumability:
     * - [ResumeValidSession]: Backend state is "uploading" or "ready" with missing chunks
     * - [ProbeFreshSession]: Session is invalid, consumed, aborted, expired, or not found
     * - [TerminalFailure]: Non-recoverable error (401/403 auth or ownership violation)
     */
    private sealed class SessionPlan {
        data class ResumeValidSession(
            val sessionId: String,
            val serverChunkSize: Long,
            val chunksToUpload: List<Int>,
            val uploadedBytes: Long
        ) : SessionPlan()

        data class ProbeFreshSession(val initialChunkSize: Long) : SessionPlan()

        data class TerminalFailure(val message: String) : SessionPlan()
    }

    /**
     * Check if a backend session state is resumable.
     *
     * Resumable states:
     * - "uploading": Session is active and accepting chunks
     * - "ready": All chunks received, may need finalization
     *
     * Non-resumable terminal states:
     * - "consumed": Session finalized/completed
     * - "aborted": Session was cancelled
     * - "expired": Session timed out
     *
     * Unknown states are treated as non-resumable for safety.
     */
    private fun isResumableSessionState(state: String): Boolean {
        return state == "uploading" || state == "ready"
    }

    /**
     * Classify resume status HTTP failure into a resume decision.
     *
     * Recoverable resume-invalid cases (start fresh probe):
     * - 404 Not Found: Session does not exist
     * - 409 Conflict: Session was invalidated or state conflict
     * - 408, 429, 500, 502, 503, 504: Temporary/server errors
     * - IOException, SocketTimeoutException: Network issues
     *
     * Non-recoverable contract/auth cases (terminal failure):
     * - 401 Unauthorized: Authentication failure
     * - 403 Forbidden: Ownership violation - DO NOT start fresh upload
     * - Other 4xx: Contract errors that should be surfaced
     */
    private fun classifyResumeStatusFailure(e: HttpException): ResumeDecision {
        val code = e.code()
        return when {
            // Session-invalid signals: discard stale session and start fresh
            code == 404 || code == 409 -> ResumeDecision.StartFreshProbe("Session invalid: $code")

            // Auth/ownership violations: terminal error, do not silently start new upload
            code == 401 -> ResumeDecision.Fail("Authentication failed: $code")
            code == 403 -> ResumeDecision.Fail("Forbidden: session ownership violation ($code)")

            // Temporary failures: start fresh probe
            code == 408 || code == 429 || code >= 500 -> ResumeDecision.StartFreshProbe("Temporary failure: $code")

            // Other 4xx: contract errors, surface as failure
            code >= 400 && code < 500 -> ResumeDecision.Fail("Client error: $code")

            // Fallback: treat as start fresh for unknown codes
            else -> ResumeDecision.StartFreshProbe("Unknown HTTP error: $code")
        }
    }

    /**
     * Classify non-HTTP exceptions during resume status check.
     * Network errors are generally retryable, so start fresh probe.
     */
    private fun classifyResumeStatusNonHttpFailure(e: Exception): ResumeDecision {
        return when (e) {
            is SocketTimeoutException, is IOException -> ResumeDecision.StartFreshProbe("Network error: ${e.message}")
            else -> ResumeDecision.StartFreshProbe("Error: ${e.message}")
        }
    }

    /**
     * Internal sealed class for resume decision classification
     */
    private sealed class ResumeDecision {
        data class ResumeExisting(val reason: String) : ResumeDecision()
        data class StartFreshProbe(val reason: String) : ResumeDecision()
        data class Fail(val message: String) : ResumeDecision()
    }

    /**
     * Phase 1: Decide whether to resume a valid session or start fresh with probing
     *
     * Session state handling:
     * - "uploading": Resume missing chunks
     * - "ready" with no missing chunks: Treat as completed, emit success
     * - "ready" with missing chunks: Contract anomaly, start fresh probe
     * - "consumed", "aborted", "expired": Discard saved session, start fresh probe
     * - Unknown state: Start fresh probe with warning
     *
     * Resume status error handling:
     * - 404, 409: Start fresh probe (session invalid)
     * - 401, 403: Terminal failure (auth/ownership violation)
     * - 408, 429, 5xx: Start fresh probe (temporary failure)
     * - Network errors: Start fresh probe
     */
    private suspend fun resolveSessionPlan(
        file: File,
        token: String,
        savedSessionId: String?,
        requestedChunkSize: Long
    ): SessionPlan {
        // If no saved session, start fresh with probe-first
        if (savedSessionId == null) {
            Log.d(TAG, "No saved session, starting probe-first with $requestedChunkSize")
            return SessionPlan.ProbeFreshSession(requestedChunkSize)
        }

        // Try to resume existing session
        return try {
            val status = whisperApi.getUploadSessionStatus(savedSessionId, "Bearer $token")

            // Check if backend returned error response
            if (!status.status || status.data == null) {
                Log.w(TAG, "Resume session $savedSessionId failed: ${status.message}. Starting fresh probe.")
                return SessionPlan.ProbeFreshSession(requestedChunkSize)
            }

            val data = status.data
            val serverChunkSize = data.chunkSizeByes.toLong()
            val sessionState = data.state

            // Check if session state is resumable
            if (!isResumableSessionState(sessionState)) {
                Log.w(
                    TAG,
                    "Session $savedSessionId in non-resumable state '$sessionState'. Discarding and starting fresh probe."
                )
                return SessionPlan.ProbeFreshSession(requestedChunkSize)
            }

            val missingRanges = data.missingRanges
            val chunksToUpload = if (missingRanges.isNullOrEmpty()) {
                emptyList()
            } else {
                parseMissingRanges(missingRanges, data.totalChunks)
            }
            val uploadedBytes = data.receivedBytes ?: 0L

            // Handle "ready" state specially
            if (sessionState == "ready") {
                if (chunksToUpload.isEmpty()) {
                    // All chunks received, session is ready for finalization
                    // Return ResumeValidSession with no chunks - will emit success immediately
                    Log.d(TAG, "Session $savedSessionId is 'ready' with all chunks present. Completing without reupload.")
                    return SessionPlan.ResumeValidSession(savedSessionId, serverChunkSize, chunksToUpload, uploadedBytes)
                } else {
                    // Contract anomaly: "ready" state should not have missing chunks
                    Log.e(
                        TAG,
                        "Session $savedSessionId is 'ready' but has ${chunksToUpload.size} missing chunks. Backend contract anomaly. Starting fresh probe."
                    )
                    return SessionPlan.ProbeFreshSession(requestedChunkSize)
                }
            }

            // "uploading" state: normal resume path
            Log.d(
                TAG,
                "Resuming session $savedSessionId. Server chunk size: $serverChunkSize, Missing chunks: ${chunksToUpload.size}"
            )
            SessionPlan.ResumeValidSession(savedSessionId, serverChunkSize, chunksToUpload, uploadedBytes)
        } catch (e: HttpException) {
            // Classify HTTP failure into resume decision
            val decision = classifyResumeStatusFailure(e)
            when (decision) {
                is ResumeDecision.ResumeExisting -> {
                    // This case shouldn't happen in catch block, but handle for completeness
                    Log.d(TAG, "Resume decision: ${decision.reason}")
                    SessionPlan.ProbeFreshSession(requestedChunkSize)
                }
                is ResumeDecision.StartFreshProbe -> {
                    Log.w(TAG, "Resume status check failed: ${decision.reason}. Starting fresh probe.")
                    SessionPlan.ProbeFreshSession(requestedChunkSize)
                }
                is ResumeDecision.Fail -> {
                    Log.e(TAG, "Resume status check failed with terminal error: ${decision.message}")
                    SessionPlan.TerminalFailure(decision.message)
                }
            }
        } catch (e: Exception) {
            // Network error during status check - classify and decide
            val decision = classifyResumeStatusNonHttpFailure(e)
            when (decision) {
                is ResumeDecision.StartFreshProbe -> {
                    Log.w(TAG, "Resume status check failed: ${decision.reason}. Starting fresh probe.")
                    SessionPlan.ProbeFreshSession(requestedChunkSize)
                }
                is ResumeDecision.Fail -> {
                    Log.e(TAG, "Resume status check failed with terminal error: ${decision.message}")
                    SessionPlan.TerminalFailure(decision.message)
                }
                else -> SessionPlan.ProbeFreshSession(requestedChunkSize)
            }
        }
    }

    /**
     * Upload on an existing valid session (resume path)
     */
    private suspend fun FlowCollector<UploadState>.uploadOnExistingSession(
        file: File,
        sessionId: String,
        serverChunkSize: Long,
        chunksToUpload: List<Int>,
        uploadedBytes: Long,
        totalSize: Long,
        token: String
    ) {
        emit(UploadState.SessionStarted(sessionId))

        if (chunksToUpload.isEmpty()) {
            emit(UploadState.Success(sessionId))
            return
        }

        val raf = RandomAccessFile(file, "r")
        try {
            val result = uploadChunks(
                file = file,
                raf = raf,
                sessionId = sessionId,
                chunksToUpload = chunksToUpload,
                serverChunkSize = serverChunkSize,
                totalSize = totalSize,
                uploadedBytesRef = uploadedBytes,
                token = token,
                mutex = Mutex(),
                emitProgress = { progress, total ->
                    val progressPercent = (progress * 100 / total).toFloat().coerceIn(0f, 100f)
                    emit(UploadState.Progress(progress, total, progressPercent))
                }
            )

            when (result) {
                is UploadResult.Success -> emit(UploadState.Success(sessionId))
                is UploadResult.Failed -> emit(UploadState.Error(result.message))
            }
        } finally {
            raf.close()
        }
    }

    /**
     * Phase 2: Probe-first session selection
     * Try one chunk with candidate size, fallback to smaller sizes only on retryable failures
     */
    private suspend fun FlowCollector<UploadState>.probeAndUpload(
        file: File,
        token: String,
        totalSize: Long,
        initialChunkSize: Long
    ) {
        var currentChunkSize = initialChunkSize
        var probeAttempts = 0

        while (probeAttempts < CHUNK_SIZE_LADDER.size) {
            val probeResult = probeSession(
                file = file,
                token = token,
                totalSize = totalSize,
                chunkSize = currentChunkSize
            )

            when (probeResult) {
                is ProbeResult.ProbeSuccess -> {
                    // Phase 3: Upload remaining chunks on proven session
                    uploadRemainingChunks(
                        file = file,
                        sessionId = probeResult.sessionId,
                        serverChunkSize = probeResult.serverChunkSize,
                        totalSize = totalSize,
                        token = token,
                        firstChunkAlreadyUploaded = true
                    )
                    return
                }
                is ProbeResult.ProbeFailure -> {
                    // Non-retryable errors (auth, bad request, etc.) should NOT trigger chunk shrinking
                    if (!probeResult.isRetryable) {
                        emit(
                            UploadState.Error(
                                "Upload failed: ${probeResult.reason}. Please check your connection and try again."
                            )
                        )
                        return
                    }

                    // Retryable error (timeout, network) - try smaller chunk size
                    probeAttempts++
                    if (probeAttempts >= CHUNK_SIZE_LADDER.size) {
                        emit(
                            UploadState.Error(
                                "Upload failed: connection too slow or unstable for server timeout. " +
                                    "Please retry on a stronger network."
                            )
                        )
                        return
                    }

                    // Try next smaller chunk size
                    val nextChunkSize = getNextSmallerChunkSize(currentChunkSize)
                    if (nextChunkSize == null) {
                        emit(
                            UploadState.Error(
                                "Upload failed: reached minimum chunk size but still failing. " +
                                    "Please retry on a stronger network."
                            )
                        )
                        return
                    }

                    Log.d(TAG, "Probe failed at ${currentChunkSize / 1024} KB, trying ${nextChunkSize / 1024} KB")
                    currentChunkSize = nextChunkSize
                    emit(UploadState.Loading("Probing with ${currentChunkSize / 1024} KB chunks..."))
                }
            }
        }
    }

    /**
     * Probe result
     */
    private sealed class ProbeResult {
        data class ProbeSuccess(
            val sessionId: String,
            val serverChunkSize: Long
        ) : ProbeResult()

        data class ProbeFailure(
            val reason: String,
            val isRetryable: Boolean
        ) : ProbeResult()
    }

    /**
     * Phase 2: Probe one chunk on a fresh session
     * Returns success if probe chunk uploaded, failure with isRetryable flag indicating whether to try smaller size
     */
    private suspend fun probeSession(
        file: File,
        token: String,
        totalSize: Long,
        chunkSize: Long
    ): ProbeResult {
        Log.d(TAG, "Creating probe session with chunk size ${chunkSize / 1024} KB")

        // Create fresh session
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
            Log.e(TAG, "Failed to create probe session: ${e.message}")
            val isRetryable = isRetryableError(e)
            return ProbeResult.ProbeFailure("Failed to create session: ${classifyError(e)}", isRetryable)
        }

        if (!sessionResponse.status || sessionResponse.data == null) {
            Log.e(TAG, "Failed to create probe session: ${sessionResponse.message}")
            // Check if error message indicates retryable failure
            val isRetryable = sessionResponse.message?.let { isRetryableMessage(it) } ?: false
            return ProbeResult.ProbeFailure("Failed to create session: ${sessionResponse.message}", isRetryable)
        }

        val sessionId = sessionResponse.data.sessionId
        val serverChunkSize = sessionResponse.data.chunkSizeByes.toLong()

        Log.d(TAG, "Probe session $sessionId created with server chunk size: $serverChunkSize")

        // Upload first chunk (chunk 0) as probe
        val raf = RandomAccessFile(file, "r")
        try {
            val offset = 0L
            val remaining = totalSize - offset
            val actualChunkSize = if (remaining < serverChunkSize) remaining else serverChunkSize
            val buffer = ByteArray(actualChunkSize.toInt())

            raf.seek(offset)
            raf.readFully(buffer)

            val requestBody = buffer.toRequestBody("application/octet-stream".toMediaTypeOrNull())

            val uploadResult = uploadChunkWithRetry(
                sessionId = sessionId,
                chunkIndex = 0,
                requestBody = requestBody,
                token = token,
                chunkSize = actualChunkSize
            )

            return when (uploadResult) {
                is ChunkUploadResult.Success -> {
                    Log.d(TAG, "Probe chunk uploaded successfully on session $sessionId")
                    ProbeResult.ProbeSuccess(sessionId, serverChunkSize)
                }
                is ChunkUploadResult.Failed -> {
                    Log.e(TAG, "Probe chunk upload failed: ${uploadResult.message}")
                    ProbeResult.ProbeFailure(uploadResult.message, uploadResult.isRetryable)
                }
            }
        } catch (e: Exception) {
            Log.e(TAG, "Probe chunk exception: ${e.message}")
            val isRetryable = isRetryableError(e)
            return ProbeResult.ProbeFailure("Probe chunk exception: ${classifyError(e)}", isRetryable)
        } finally {
            raf.close()
        }
    }

    /**
     * Phase 3: Upload remaining chunks after successful probe
     */
    private suspend fun FlowCollector<UploadState>.uploadRemainingChunks(
        file: File,
        sessionId: String,
        serverChunkSize: Long,
        totalSize: Long,
        token: String,
        firstChunkAlreadyUploaded: Boolean
    ) {
        emit(UploadState.SessionStarted(sessionId))

        val totalChunks = ((totalSize + serverChunkSize - 1) / serverChunkSize).toInt()
        val chunksToUpload = if (firstChunkAlreadyUploaded) {
            1 until totalChunks
        } else {
            0 until totalChunks
        }

        if (chunksToUpload.isEmpty()) {
            emit(UploadState.Success(sessionId))
            return
        }

        val raf = RandomAccessFile(file, "r")
        val uploadedBytesRef = if (firstChunkAlreadyUploaded) serverChunkSize else 0L

        try {
            val result = uploadChunks(
                file = file,
                raf = raf,
                sessionId = sessionId,
                chunksToUpload = chunksToUpload.toList(),
                serverChunkSize = serverChunkSize,
                totalSize = totalSize,
                uploadedBytesRef = uploadedBytesRef,
                token = token,
                mutex = Mutex(),
                emitProgress = { progress, total ->
                    val progressPercent = (progress * 100 / total).toFloat().coerceIn(0f, 100f)
                    emit(UploadState.Progress(progress, total, progressPercent))
                }
            )

            when (result) {
                is UploadResult.Success -> emit(UploadState.Success(sessionId))
                is UploadResult.Failed -> emit(UploadState.Error(result.message))
            }
        } finally {
            raf.close()
        }
    }

    /**
     * Result of single chunk upload attempt
     */
    private sealed class ChunkUploadResult {
        object Success : ChunkUploadResult()
        data class Failed(val message: String, val isRetryable: Boolean) : ChunkUploadResult()
    }

    /**
     * Result of upload attempt
     */
    private sealed class UploadResult {
        object Success : UploadResult()
        data class Failed(val message: String) : UploadResult()
    }

    /**
     * Result of a single chunk upload in batch processing
     */
    private data class ChunkUploadTaskResult(
        val chunkIndex: Int,
        val chunkSize: Long,
        val uploadResult: ChunkUploadResult
    )

    /**
     * Upload chunks with status-aware retry logic and bounded concurrency
     * Uses:
     * - Concurrency 1 for chunk sizes > 512 KB
     * - Concurrency 2 for chunk sizes <= 512 KB
     * Progress updates are emitted from the parent coroutine only, serialized per batch.
     */
    private suspend fun uploadChunks(
        file: File,
        raf: RandomAccessFile,
        sessionId: String,
        chunksToUpload: List<Int>,
        serverChunkSize: Long,
        totalSize: Long,
        uploadedBytesRef: Long,
        token: String,
        mutex: Mutex,
        emitProgress: suspend (Long, Long) -> Unit
    ): UploadResult {
        var uploadedBytes = uploadedBytesRef

        // Determine max concurrency based on chunk size
        val maxConcurrency = if (serverChunkSize > 512 * 1024) {
            1
        } else {
            2
        }

        try {
            // Process chunks in batches for bounded concurrency
            val batches = chunksToUpload.chunked(maxConcurrency)

            for (batch in batches) {
                val batchResults = coroutineScope {
                    val deferreds = batch.map { i ->
                        async {
                            val offset = i.toLong() * serverChunkSize
                            val remaining = totalSize - offset
                            val currentChunkSize = if (remaining < serverChunkSize) {
                                remaining
                            } else {
                                serverChunkSize
                            }

                            val buffer = ByteArray(currentChunkSize.toInt())

                            mutex.withLock {
                                raf.seek(offset)
                                raf.readFully(buffer)
                            }

                            val requestBody = buffer.toRequestBody(
                                "application/octet-stream".toMediaTypeOrNull()
                            )

                            val result = uploadChunkWithRetry(
                                sessionId = sessionId,
                                chunkIndex = i,
                                requestBody = requestBody,
                                token = token,
                                chunkSize = currentChunkSize
                            )

                            ChunkUploadTaskResult(i, currentChunkSize, result)
                        }
                    }

                    // Await all results in this batch
                    deferreds.awaitAll()
                }

                // Sort results by chunk index for stable progress ordering
                val sortedResults = batchResults.sortedBy { it.chunkIndex }

                // Process results and emit progress from parent coroutine only
                for (taskResult in sortedResults) {
                    when (val result = taskResult.uploadResult) {
                        is ChunkUploadResult.Success -> {
                            uploadedBytes += taskResult.chunkSize
                            // Emit progress in parent coroutine context (serialized)
                            emitProgress(uploadedBytes, totalSize)
                        }
                        is ChunkUploadResult.Failed -> {
                            // Stop upload on first failure
                            val errorMessage = if (result.isRetryable) {
                                "Upload failed: ${result.message}"
                            } else {
                                // Non-retryable error - propagate real message directly
                                result.message
                            }
                            return UploadResult.Failed(errorMessage)
                        }
                    }
                }
            }

            return UploadResult.Success
        } catch (e: Exception) {
            val msg = e.message ?: "Unknown error"
            return UploadResult.Failed("Upload failed: $msg")
        }
    }

    /**
     * Upload a single chunk with status-aware retry logic
     * Responsible for:
     * - Bounded retries
     * - Retryable error classification
     * - Session status check after retryable failure
     * - Response-loss recovery (treating "chunk no longer missing" as success)
     */
    private suspend fun uploadChunkWithRetry(
        sessionId: String,
        chunkIndex: Int,
        requestBody: okhttp3.RequestBody,
        token: String,
        chunkSize: Long
    ): ChunkUploadResult {
        var lastError = ""
        var retryDelay = 1000L // 1 second initial delay

        for (retry in 0 until MAX_RETRIES_PER_CHUNK) {
            try {
                val ackResponse = whisperApi.uploadChunk(
                    sessionId = sessionId,
                    chunkIndex = chunkIndex,
                    chunk = requestBody,
                    token = "Bearer $token"
                )

                if (ackResponse.status) {
                    Log.d(TAG, "Chunk $chunkIndex uploaded successfully")
                    return ChunkUploadResult.Success
                } else {
                    lastError = ackResponse.message ?: "Unknown API Error"
                    // Check if error is retryable
                    if (!isRetryableMessage(lastError)) {
                        Log.e(TAG, "Chunk $chunkIndex failed with non-retryable error: $lastError")
                        return ChunkUploadResult.Failed(lastError, isRetryable = false)
                    }
                }
            } catch (e: Exception) {
                lastError = classifyError(e)

                // On retryable failure, check session status before retrying
                if (retry < MAX_RETRIES_PER_CHUNK - 1 && isRetryableError(e)) {
                    try {
                        // Check session status after retryable error (server may have received chunk but response was lost)
                        val status = whisperApi.getUploadSessionStatus(sessionId, "Bearer $token")
                        if (status.status && status.data != null) {
                            val data = status.data
                            // Check if this chunk is no longer in missing ranges (already received)
                            val missingRanges = data.missingRanges
                            val chunkAlreadyReceived = if (missingRanges.isNullOrEmpty()) {
                                // No missing ranges means all chunks received
                                true
                            } else {
                                // Check if current chunk index is NOT in the missing ranges
                                !isChunkMissing(chunkIndex, missingRanges)
                            }

                            if (chunkAlreadyReceived) {
                                Log.d(TAG, "Chunk $chunkIndex failed but session status shows it was already received (response loss). Treating as success.")
                                return ChunkUploadResult.Success
                            }
                            Log.d(TAG, "Chunk $chunkIndex failed and is still missing according to session status")
                        }
                    } catch (statusEx: Exception) {
                        Log.w(TAG, "Failed to check session status after chunk failure: ${statusEx.message}")
                    }
                }
            }

            if (retry < MAX_RETRIES_PER_CHUNK - 1) {
                Log.d(TAG, "Chunk $chunkIndex retry #${retry + 1} after ${retryDelay}ms: $lastError")
                delay(retryDelay)
                retryDelay *= 2 // Exponential backoff: 1s, 2s, 4s
            }
        }

        Log.e(TAG, "Chunk $chunkIndex exhausted all retries: $lastError")
        // After exhausting retries, return failure with retryable flag based on error type
        val isRetryable = isRetryableMessage(lastError) || lastError.contains("timeout", ignoreCase = true)
        return ChunkUploadResult.Failed(lastError, isRetryable)
    }

    /**
     * Check if a chunk index is in the missing ranges list
     * @VisibleForTesting - exposed for unit testing
     */
    @androidx.annotation.VisibleForTesting
    internal fun isChunkMissing(chunkIndex: Int, missingRanges: List<String>): Boolean {
        return missingRanges.any { range ->
            try {
                if (range.contains("-")) {
                    val parts = range.split("-")
                    if (parts.size == 2) {
                        val start = parts[0].toInt()
                        val end = parts[1].toInt()
                        chunkIndex in start..end
                    } else {
                        false
                    }
                } else {
                    range.toInt() == chunkIndex
                }
            } catch (e: Exception) {
                Log.w(TAG, "Failed to parse missing range: $range")
                false // Assume not missing on parse error (safer to retry)
            }
        }
    }

    /**
     * Classify exception into user-friendly message
     * @VisibleForTesting - exposed for unit testing
     */
    @androidx.annotation.VisibleForTesting
    internal fun classifyError(e: Exception): String {
        return when (e) {
            is SocketTimeoutException -> "Connection timeout"
            is HttpException -> {
                val code = e.code()
                when {
                    code == 408 -> "Request timeout (server)"
                    code == 429 -> "Rate limited"
                    code >= 500 -> "Server error ($code)"
                    else -> "HTTP error ($code)"
                }
            }
            is IOException -> "Network error: ${e.message}"
            else -> e.message ?: "Unknown error"
        }
    }

    /**
     * Check if exception is retryable
     * @VisibleForTesting - exposed for unit testing
     */
    @androidx.annotation.VisibleForTesting
    internal fun isRetryableError(e: Exception): Boolean {
        return when (e) {
            is SocketTimeoutException -> true
            is HttpException -> RETRYABLE_HTTP_CODES.contains(e.code())
            is IOException -> true // Network issues are generally retryable
            else -> false
        }
    }

    /**
     * Check if error message indicates retryable failure
     * @VisibleForTesting - exposed for unit testing
     */
    @androidx.annotation.VisibleForTesting
    internal fun isRetryableMessage(msg: String): Boolean {
        val lowerMsg = msg.lowercase()
        return lowerMsg.contains("timeout") ||
            lowerMsg.contains("retry") ||
            lowerMsg.contains("temporary") ||
            RETRYABLE_HTTP_CODES.any { code ->
                lowerMsg.contains(code.toString())
            }
    }

    /**
     * Get the next smaller chunk size from the ladder.
     * Given a chunk size that failed, returns the next smaller ladder value to try.
     * Returns null if already at minimum chunk size.
     * @VisibleForTesting - exposed for unit testing
     */
    @androidx.annotation.VisibleForTesting
    internal fun getNextSmallerChunkSize(currentChunkSize: Long): Long? {
        // Find the largest ladder value strictly less than currentChunkSize
        for (i in CHUNK_SIZE_LADDER.indices) {
            val ladderValue = CHUNK_SIZE_LADDER[i]
            if (ladderValue < currentChunkSize) {
                return ladderValue
            }
        }

        // Current size is at or below minimum, no smaller size available
        return null
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
                Log.w(TAG, "Failed to parse range: $range")
            }
        }
        return missingIndices.toList().sorted()
    }

    /**
     * Upload file directly to signed S3 URL
     */
    private suspend fun uploadToSignedUrl(
        file: File,
        signedUrl: String,
        contentType: String,
        totalSize: Long,
        onProgress: suspend (Long, Long) -> Unit
    ): Boolean {
        if (totalSize < MULTIPART_MIN_PART_SIZE) {
            return uploadSinglePut(file, signedUrl, contentType, totalSize, onProgress)
        }
        return uploadMultipart(file, signedUrl, contentType, totalSize, onProgress)
    }

    private suspend fun uploadSinglePut(
        file: File,
        signedUrl: String,
        contentType: String,
        totalSize: Long,
        onProgress: suspend (Long, Long) -> Unit
    ): Boolean {
        return try {
            val client = okhttp3.OkHttpClient()
                .newBuilder()
                .connectTimeout(30, java.util.concurrent.TimeUnit.SECONDS)
                .readTimeout(5, java.util.concurrent.TimeUnit.MINUTES)
                .writeTimeout(5, java.util.concurrent.TimeUnit.MINUTES)
                .build()

            val requestBody = okhttp3.RequestBody.create(
                contentType.toMediaTypeOrNull(),
                file
            )

            val request = okhttp3.Request.Builder()
                .url(signedUrl)
                .put(requestBody)
                .build()

            val response = client.newCall(request).execute()

            if (response.isSuccessful) {
                onProgress(totalSize, totalSize)
                true
            } else {
                Log.e(TAG, "Signed URL upload failed: ${response.code} ${response.message}")
                false
            }
        } catch (e: Exception) {
            Log.e(TAG, "Signed URL upload exception: ${e.message}")
            false
        }
    }

    private suspend fun uploadMultipart(
        file: File,
        signedUrl: String,
        contentType: String,
        totalSize: Long,
        onProgress: suspend (Long, Long) -> Unit
    ): Boolean {
        val client = okhttp3.OkHttpClient()
            .newBuilder()
            .connectTimeout(30, java.util.concurrent.TimeUnit.SECONDS)
            .readTimeout(5, java.util.concurrent.TimeUnit.MINUTES)
            .writeTimeout(5, java.util.concurrent.TimeUnit.MINUTES)
            .build()

        var uploadId: String? = null
        val partETags = mutableListOf<String>()
        val raf = RandomAccessFile(file, "r")

        try {
            val createRequest = okhttp3.Request.Builder()
                .url(signedUrl)
                .post(okhttp3.RequestBody.create(contentType.toMediaTypeOrNull(), ""))
                .header("Content-Type", contentType)
                .header("x-amz-backmodes", "CreateMultipartUpload")
                .build()

            val createResponse = client.newCall(createRequest).execute()
            if (!createResponse.isSuccessful) {
                Log.e(TAG, "CreateMultipartUpload failed: ${createResponse.code}")
                return false
            }
            val createBody = createResponse.body?.string() ?: ""
            uploadId = extractUploadIdFromResponse(createBody)
            if (uploadId == null) {
                Log.e(TAG, "Failed to extract uploadId from CreateMultipartUpload response")
                return false
            }
            createResponse.close()

            val partSize = calculatePartSize(totalSize)
            val numParts = ((totalSize + partSize - 1) / partSize).toInt()
            var uploadedBytes = 0L

            for (partNumber in 1..numParts) {
                val offset = (partNumber - 1).toLong() * partSize
                val remaining = totalSize - offset
                val currentPartSize = minOf(partSize, remaining)

                val buffer = ByteArray(currentPartSize.toInt())
                raf.seek(offset)
                raf.readFully(buffer)

                var partSuccess = false
                var lastError: Exception? = null

                for (retryCount in 0 until MAX_MULTIPART_RETRIES) {
                    try {
                        val partRequestBody = okhttp3.RequestBody.create(
                            contentType.toMediaTypeOrNull(),
                            buffer
                        )

                        val partUrl = "$signedUrl&uploadId=$uploadId&partNumber=$partNumber"
                        val partRequest = okhttp3.Request.Builder()
                            .url(partUrl)
                            .put(partRequestBody)
                            .build()

                        val partResponse = client.newCall(partRequest).execute()

                        if (partResponse.isSuccessful) {
                            val etag = partResponse.header("ETag")?.replace("\"", "") ?: ""
                            partETags.add(etag)
                            partSuccess = true
                            partResponse.close()
                            break
                        } else {
                            lastError = Exception("UploadPart failed: ${partResponse.code}")
                            partResponse.close()
                        }
                    } catch (e: Exception) {
                        lastError = e
                    }

                    if (!partSuccess && retryCount < MAX_MULTIPART_RETRIES - 1) {
                        val delayMs = MULTIPART_BACKOFF_DELAYS[retryCount]
                        delay(delayMs)
                    }
                }

                if (!partSuccess) {
                    Log.e(TAG, "UploadPart $partNumber failed after $MAX_MULTIPART_RETRIES retries")
                    abortMultipartUpload(client, signedUrl, uploadId)
                    return false
                }

                uploadedBytes += currentPartSize
                onProgress(uploadedBytes, totalSize)
            }

            val completeResult = completeMultipartUpload(client, signedUrl, uploadId, partETags)
            if (!completeResult) {
                Log.e(TAG, "CompleteMultipartUpload failed")
                abortMultipartUpload(client, signedUrl, uploadId)
                return false
            }

            onProgress(totalSize, totalSize)
            return true

        } catch (e: Exception) {
            Log.e(TAG, "Multipart upload exception: ${e.message}")
            uploadId?.let { abortMultipartUpload(client, signedUrl, it) }
            return false
        } finally {
            raf.close()
        }
    }

    private fun calculatePartSize(totalSize: Long): Long {
        val minPartSize = MULTIPART_MIN_PART_SIZE
        val numParts = ((totalSize + minPartSize - 1) / minPartSize).toInt()
        return (totalSize + numParts - 1) / numParts
    }

    private fun extractUploadIdFromResponse(responseBody: String): String? {
        val json = JSONObject(responseBody)
        return json.optString("UploadId")
    }

    private suspend fun completeMultipartUpload(
        client: okhttp3.OkHttpClient,
        signedUrl: String,
        uploadId: String,
        partETags: List<String>
    ): Boolean {
        val etagsJson = partETags.mapIndexed { index, etag ->
            """{"PartNumber":${index + 1},"ETag":"$etag"}"""
        }.joinToString(",")

        val completeBody = """{"UploadId":"$uploadId","Parts":[$etagsJson]}"""

        val request = okhttp3.Request.Builder()
            .url(signedUrl)
            .post(completeBody.toRequestBody("application/json".toMediaTypeOrNull()))
            .header("Content-Type", "application/json")
            .header("x-amz-backmodes", "CompleteMultipartUpload")
            .build()

        return try {
            val response = client.newCall(request).execute()
            val success = response.isSuccessful
            response.close()
            success
        } catch (e: Exception) {
            Log.e(TAG, "CompleteMultipartUpload exception: ${e.message}")
            false
        }
    }

    private suspend fun abortMultipartUpload(
        client: okhttp3.OkHttpClient,
        signedUrl: String,
        uploadId: String
    ) {
        try {
            val abortUrl = "$signedUrl&uploadId=$uploadId"
            val request = okhttp3.Request.Builder()
                .url(abortUrl)
                .delete()
                .header("x-amz-backmodes", "AbortMultipartUpload")
                .build()

            client.newCall(request).execute().close()
            Log.i(TAG, "AbortMultipartUpload succeeded for uploadId: $uploadId")
        } catch (e: Exception) {
            Log.w(TAG, "AbortMultipartUpload failed: ${e.message}")
        }
    }

    private fun hashPasswordWithArgon2(password: String): String {
        val salt = ByteArray(16)
        java.security.SecureRandom().nextBytes(salt)

        val factory = org.bouncycastle.crypto.generators.Argon2BytesGenerator()
        val params = org.bouncycastle.crypto.params.Argon2Parameters.Builder(
            org.bouncycastle.crypto.params.Argon2Parameters.ARGON2_id
        )
            .withSalt(salt)
            .withMemoryAsKB(65536)
            .withIterations(3)
            .withParallelism(4)
            .build()
        factory.init(params)

        val hash = ByteArray(32)
        factory.generateBytes(password.toCharArray(), hash)

        val saltHex = salt.joinToString("") { "%02x".format(it) }
        val hashHex = hash.joinToString("") { "%02x".format(it) }
        return "$saltHex:$hashHex"
    }

    private suspend fun saveRecordingMetadata(
        objectKey: String,
        bookingId: String,
        passwordHash: String,
        token: String
    ) {
        try {
            val request = SaveRecordingRequestDto(
                s3Key = objectKey,
                bookingId = bookingId,
                passwordHash = passwordHash
            )
            whisperApi.saveRecording(request, "Bearer $token")
        } catch (e: Exception) {
            Log.w(TAG, "Failed to save recording metadata: ${e.message}")
        }
    }
}
