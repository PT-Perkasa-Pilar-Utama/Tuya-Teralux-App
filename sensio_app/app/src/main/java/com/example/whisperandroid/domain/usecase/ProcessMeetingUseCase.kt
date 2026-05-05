package com.example.whisperandroid.domain.usecase

import android.content.SharedPreferences
import android.util.Log
import com.example.whisperandroid.domain.repository.Resource
import java.io.File
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.channelFlow
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import org.json.JSONObject

sealed class MeetingProcessState {
    object Idle : MeetingProcessState()

    object Recording : MeetingProcessState()

    data class Uploading(val progress: Int) : MeetingProcessState()

    data class Transcribing(val taskId: String) : MeetingProcessState()
    data class Translating(val taskId: String) : MeetingProcessState()
    data class Summarizing(val taskId: String) : MeetingProcessState()

    data class Success(
        val summary: String,
        val pdfUrl: String?
    ) : MeetingProcessState()

    data class Error(
        val message: String
    ) : MeetingProcessState()

    object Cancelled : MeetingProcessState()
}

/**
 * Persisted submission state for resumable uploads with idempotency protection.
 * Stored per audio file path to survive service/app restarts during the same logical submission.
 */
data class SubmissionState(
    val sessionId: String,
    val idempotencyKey: String,
    val createdAt: Long = System.currentTimeMillis()
)

class ProcessMeetingUseCase(
    private val pipelineRepository: com.example.whisperandroid.domain.repository.PipelineRepository,
    private val uploadRepository: com.example.whisperandroid.domain.repository.UploadRepository,
    private val prefs: SharedPreferences
) {
    companion object {
        private const val SUBMISSION_STATE_PREFIX = "submission_"
    }

    /**
     * Load persisted submission state for a file.
     */
    private fun loadSubmissionState(audioFilePath: String): SubmissionState? {
        val key = SUBMISSION_STATE_PREFIX + audioFilePath
        val json = prefs.getString(key, null) ?: return null
        return try {
            val obj = JSONObject(json)
            SubmissionState(
                sessionId = obj.getString("sessionId"),
                idempotencyKey = obj.getString("idempotencyKey"),
                createdAt = obj.optLong("createdAt", System.currentTimeMillis())
            )
        } catch (e: Exception) {
            Log.w("ProcessMeeting", "Failed to parse submission state for $audioFilePath: ${e.message}")
            null
        }
    }

    /**
     * Persist submission state for a file.
     */
    private fun saveSubmissionState(audioFilePath: String, state: SubmissionState) {
        val key = SUBMISSION_STATE_PREFIX + audioFilePath
        val json = JSONObject()
            .put("sessionId", state.sessionId)
            .put("idempotencyKey", state.idempotencyKey)
            .put("createdAt", state.createdAt)
            .toString()
        prefs.edit().putString(key, json).apply()
    }

    /**
     * Clear persisted submission state for a file.
     */
    private fun clearSubmissionState(audioFilePath: String) {
        val key = SUBMISSION_STATE_PREFIX + audioFilePath
        prefs.edit().remove(key).apply()
        Log.d("ProcessMeeting", "Cleared submission state for file: $audioFilePath")
    }
    suspend operator fun invoke(
        audioFile: File,
        token: String,
        targetLang: String = "English",
        macAddress: String? = null,
        idempotencyKey: String? = null,
        sessionId: String? = null
    ): Flow<MeetingProcessState> = channelFlow {
        send(MeetingProcessState.Uploading(0))

        val targetLangCode = when (targetLang.lowercase()) {
            "id", "indonesia" -> "id"
            else -> "en"
        }

        val fileKey = audioFile.absolutePath

        // Load or create submission state
        var submissionState = loadSubmissionState(fileKey)

        // Use provided idempotency key only if no saved state exists (fresh submission)
        var finalIdempotencyKey = idempotencyKey
        var finalSessionId: String? = null

        if (submissionState != null) {
            // Resume existing submission: reuse saved idempotency key and session ID
            Log.d("ProcessMeeting", "Resuming submission for file: $fileKey with saved session ${submissionState.sessionId}")
            finalIdempotencyKey = submissionState.idempotencyKey
            finalSessionId = submissionState.sessionId
        } else if (idempotencyKey != null) {
            // Fresh submission with provided idempotency key
            // Session will be created during upload
            Log.d("ProcessMeeting", "Starting fresh submission for file: $fileKey with idempotency key: $idempotencyKey")
        }

        try {
            uploadRepository.uploadFile(audioFile, token, sessionId = finalSessionId)
                .collect { uploadState ->
                    // Check for cancellation
                    if (!isActive) {
                        // DO NOT clear state here - this allows resume after service restart/recreation
                        // State is only cleared on:
                        // 1. User-initiated cancellation (via clearSessionState)
                        // 2. Terminal states (success/error)
                        // 3. Session invalidation (409 conflict)
                        Log.d("ProcessMeeting", "Upload cancelled by coroutine cancellation - preserving state for resume")
                        return@collect
                    }

                    when (uploadState) {
                        is com.example.whisperandroid.domain.repository.UploadState.SessionStarted -> {
                            // Persist the session ID immediately when known
                            // Also persist idempotency key if we have one
                            val currentState = submissionState
                                ?: SubmissionState(
                                    sessionId = uploadState.sessionId,
                                    idempotencyKey = finalIdempotencyKey ?: "meeting_${audioFile.lastModified()}_${System.currentTimeMillis()}"
                                )

                            if (submissionState == null) {
                                // First time seeing session ID - save it
                                submissionState = currentState.copy(sessionId = uploadState.sessionId)
                                saveSubmissionState(fileKey, submissionState!!)
                            } else {
                                // Update session ID in existing state
                                submissionState = currentState.copy(sessionId = uploadState.sessionId)
                                saveSubmissionState(fileKey, submissionState!!)
                            }
                        }
                        is com.example.whisperandroid.domain.repository.UploadState.Success -> {
                            finalSessionId = uploadState.sessionId
                        }
                        is com.example.whisperandroid.domain.repository.UploadState.Error -> {
                            // Check if error indicates session invalidation (corrupt session)
                            val isSessionInvalidated = uploadState.message.contains("invalidated") ||
                                uploadState.message.contains("409") ||
                                uploadState.message.contains("conflict")

                            if (isSessionInvalidated) {
                                // Clear the corrupted submission state
                                clearSubmissionState(fileKey)
                                Log.w("ProcessMeeting", "Cleared corrupted submission state from preferences for file: $fileKey")
                            }

                            send(MeetingProcessState.Error("Upload failed: ${uploadState.message}"))
                        }
                        is com.example.whisperandroid.domain.repository.UploadState.Progress -> {
                            val progressPercent = uploadState.percent.toInt().coerceIn(0, 100)
                            send(MeetingProcessState.Uploading(progressPercent))
                            Log.d("ProcessMeeting", "Upload progress: $progressPercent%")
                        }
                        else -> {}
                    }
                }
        } catch (e: kotlinx.coroutines.CancellationException) {
            Log.d("ProcessMeeting", "Upload cancelled: ${e.message}")
            throw e
        }

        // Check if cancelled before proceeding to pipeline
        if (!isActive || finalSessionId == null) {
            Log.d("ProcessMeeting", "Process cancelled or no session ID, aborting pipeline")
            return@channelFlow
        }

        var pipelineTaskId: String? = null

        try {
            pipelineRepository.executePipelineByUpload(
                sessionId = finalSessionId!!,
                language = "id", targetLanguage = targetLangCode,
                summarize = true,
                refine = true,
                diarize = false,
                context = null,
                style = null,
                date = null,
                location = null,
                participants = null,
                macAddress = macAddress,
                token = token,
                idempotencyKey = finalIdempotencyKey
            ).collect { result ->
                // Check for cancellation
                if (!isActive) {
                    // DO NOT clear state here - allows resume after service restart/recreation
                    Log.d("ProcessMeeting", "Pipeline initiation cancelled by coroutine cancellation - preserving state for resume")
                    return@collect
                }

                when (result) {
                    is Resource.Success -> {
                        pipelineTaskId = result.data
                        // Store task ID in MeetingProcessManager for cancellation
                        pipelineTaskId?.let { id ->
                            com.example.whisperandroid.data.manager.MeetingProcessManager.setPipelineTaskId(id)
                            Log.d("ProcessMeeting", "Stored pipeline task ID for cancellation: $id")
                        }
                    }
                    is Resource.Error -> {
                        send(MeetingProcessState.Error("Pipeline initiation failed: ${result.message}"))
                    }
                    else -> {}
                }
            }
        } catch (e: kotlinx.coroutines.CancellationException) {
            Log.d("ProcessMeeting", "Pipeline initiation cancelled: ${e.message}")
            throw e
        }

        // Check if cancelled before proceeding to polling
        if (!isActive || pipelineTaskId == null) {
            Log.d("ProcessMeeting", "Process cancelled or no pipeline task ID, aborting")
            // DO NOT clear state here - allows resume after service restart/recreation
            // State will be cleared only on terminal outcomes or user-initiated cancellation
            return@channelFlow
        }

        // DO NOT clear submission state here - it must persist during polling phase
        // to allow restart safety. State will be cleared only on terminal outcomes.

        var isCompleted = false

        // Polling Loop: Check pipeline status via HTTP
        while (!isCompleted) {
            // Check for cancellation at the start of each iteration
            if (!isActive) {
                Log.d("ProcessMeeting", "Polling loop cancelled by coroutine cancellation - preserving state for resume")
                // DO NOT clear state here - allows resume after service restart/recreation
                break
            }

            var shouldDelay = false
            try {
                pipelineRepository.pollPipelineStatus(pipelineTaskId!!, token).collect { result ->
                    // Check for cancellation during polling
                    if (!isActive) {
                        // DO NOT clear state here - allows resume after service restart/recreation
                        Log.d("ProcessMeeting", "Polling cancelled by coroutine cancellation - preserving state for resume")
                        return@collect
                    }

                    when (result) {
                        is Resource.Success -> {
                            val statusDto = result.data!!
                            val stages = statusDto.stages ?: emptyMap()

                            val transcribeStatus = stages["transcription"]?.status
                            val translateStatus = stages["translation"]?.status
                            val summarizeStatus = stages["summary"]?.status

                            if (!isCompleted) {
                                when {
                                    summarizeStatus == "processing" -> send(
                                        MeetingProcessState.Summarizing(pipelineTaskId!!)
                                    )
                                    translateStatus == "processing" -> send(
                                        MeetingProcessState.Translating(pipelineTaskId!!)
                                    )
                                    transcribeStatus == "processing" ||
                                        transcribeStatus == "pending" -> send(
                                        MeetingProcessState.Transcribing(pipelineTaskId!!)
                                    )
                                }
                            }

                            if (statusDto.overallStatus == "completed") {
                                isCompleted = true
                                val summaryStage = stages["summary"]
                                if (summaryStage?.status == "completed") {
                                    val resMap = summaryStage.result as? Map<*, *>
                                    val summary = resMap?.get("summary") as? String
                                        ?: "Meeting summary is ready"
                                    val pdfUrl = resMap?.get("pdf_url") as? String
                                    send(MeetingProcessState.Success(summary, pdfUrl))
                                } else {
                                    // Handle cases where overall is completed but summary stage is missing/skipped
                                    send(MeetingProcessState.Success("Processing complete", null))
                                }
                                // Clear state on terminal success
                                clearSubmissionState(fileKey)
                            } else if (statusDto.overallStatus == "failed") {
                                isCompleted = true
                                send(MeetingProcessState.Error("Pipeline execution failed"))
                                // Clear state on terminal failure
                                clearSubmissionState(fileKey)
                            } else if (statusDto.overallStatus == "cancelled") {
                                isCompleted = true
                                send(MeetingProcessState.Cancelled)
                                // Clear state on user-initiated cancellation
                                clearSubmissionState(fileKey)
                            } else {
                                shouldDelay = true
                            }
                        }
                        is Resource.Error -> {
                            shouldDelay = true
                        }
                        else -> {}
                    }
                }
            } catch (e: kotlinx.coroutines.CancellationException) {
                Log.d("ProcessMeeting", "Polling cancelled: ${e.message}")
                break
            }

            if (shouldDelay && !isCompleted) {
                kotlinx.coroutines.delay(5000) // Polling interval 5s
            }
        }
    }

    /**
     * Clear the persisted submission state for a given audio file.
     * This should be called when the user cancels the processing
     * to prevent resuming an abandoned upload session.
     */
    fun clearSessionState(audioFilePath: String) {
        clearSubmissionState(audioFilePath)
    }
}
