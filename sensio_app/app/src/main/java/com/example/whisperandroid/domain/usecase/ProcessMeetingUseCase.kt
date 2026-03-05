package com.example.whisperandroid.domain.usecase

import android.content.SharedPreferences
import android.util.Log
import com.example.whisperandroid.domain.repository.Resource
import java.io.File
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.channelFlow
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
}

class ProcessMeetingUseCase(
    private val pipelineRepository: com.example.whisperandroid.domain.repository.PipelineRepository,
    private val uploadRepository: com.example.whisperandroid.domain.repository.UploadRepository,
    private val mqttHelper: com.example.whisperandroid.util.MqttHelper,
    private val prefs: SharedPreferences
) {
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
        val savedSessionId = prefs.getString(fileKey, null)
        var finalSessionId: String? = null

        uploadRepository.uploadFile(audioFile, token, sessionId = savedSessionId).collect { uploadState ->
            when (uploadState) {
                is com.example.whisperandroid.domain.repository.UploadState.SessionStarted -> {
                    // Persist the session ID immediately when known
                    prefs.edit().putString(fileKey, uploadState.sessionId).apply()
                }
                is com.example.whisperandroid.domain.repository.UploadState.Success -> {
                    finalSessionId = uploadState.sessionId
                }
                is com.example.whisperandroid.domain.repository.UploadState.Error -> {
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

        if (finalSessionId == null) return@channelFlow

        var pipelineTaskId: String? = null
        pipelineRepository.executePipelineByUpload(
            sessionId = finalSessionId!!,
            language = "id", 
            targetLanguage = targetLangCode,
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
            idempotencyKey = idempotencyKey
        ).collect { result ->
            when (result) {
                is Resource.Success -> {
                    pipelineTaskId = result.data
                }
                is Resource.Error -> {
                    send(MeetingProcessState.Error("Pipeline initiation failed: ${result.message}"))
                }
                else -> {}
            }
        }

        if (pipelineTaskId == null) return@channelFlow

        // The session is consumed by the pipeline successfully, remove it from local cache
        prefs.edit().remove(fileKey).apply()

        var isCompleted = false
        var lastEventTime = System.currentTimeMillis()

        val mqttJob = launch {
            mqttHelper.messages.collect { (topic, message) ->
                if (topic.endsWith("/task")) {
                    try {
                        val json = JSONObject(message)
                        val eventTaskId = json.optString("task_id", "")
                        if (eventTaskId == pipelineTaskId) {
                            lastEventTime = System.currentTimeMillis()
                            val event = json.optString("event", "")
                            val overallStatus = json.optString("overall_status", "")
                            val stage = json.optString("stage", "")
                            val stageStatus = json.optString("stage_status", "")
                            val error = json.optString("error", "")

                            if (overallStatus == "failed") {
                                send(MeetingProcessState.Error("Pipeline failed at $stage: $error"))
                                isCompleted = true
                            } else if (overallStatus == "completed" || event == "completed") {
                                // Signal overall completion to trigger final poll
                                lastEventTime = 0 
                            } else if (!isCompleted) {
                                when (stage) {
                                    "transcription" -> send(MeetingProcessState.Transcribing(pipelineTaskId!!))
                                    "translation" -> send(MeetingProcessState.Translating(pipelineTaskId!!))
                                    "summary" -> send(MeetingProcessState.Summarizing(pipelineTaskId!!))
                                }
                            }
                        }
                    } catch (e: Exception) {
                        Log.e("ProcessMeeting", "Failed to parse MQTT task event: ${e.message}")
                    }
                }
            }
        }

        // Multiplexing Loop: Wait for MQTT, fallback to Polling if dead
        while (!isCompleted) {
            val isMqttConnected = mqttHelper.connectionStatus.value == com.example.whisperandroid.util.MqttHelper.MqttConnectionStatus.CONNECTED
            val timeSinceLastEvent = System.currentTimeMillis() - lastEventTime

            // Fallback rule: MQTT disconnected OR no event for 10 seconds
            if (!isMqttConnected || timeSinceLastEvent > 10000) {
                var shouldDelay = false
                pipelineRepository.pollPipelineStatus(pipelineTaskId!!, token).collect { result ->
                    when (result) {
                        is Resource.Success -> {
                            val statusDto = result.data!!
                            val stages = statusDto.stages ?: emptyMap()

                            val transcribeStatus = stages["transcription"]?.status
                            val translateStatus = stages["translation"]?.status
                            val summarizeStatus = stages["summary"]?.status

                            if (!isCompleted) {
                                when {
                                    summarizeStatus == "processing" -> send(MeetingProcessState.Summarizing(pipelineTaskId!!))
                                    translateStatus == "processing" -> send(MeetingProcessState.Translating(pipelineTaskId!!))
                                    transcribeStatus == "processing" || transcribeStatus == "pending" -> send(MeetingProcessState.Transcribing(pipelineTaskId!!))
                                }
                            }

                            if (statusDto.overallStatus == "completed") {
                                isCompleted = true
                                val summaryStage = stages["summary"]
                                if (summaryStage?.status == "completed") {
                                    val resMap = summaryStage.result as? Map<*, *>
                                    val summary = resMap?.get("summary") as? String ?: "Meeting summary is ready"
                                    val pdfUrl = resMap?.get("pdf_url") as? String
                                    send(MeetingProcessState.Success(summary, pdfUrl))
                                } else {
                                    // Handle cases where overall is completed but summary stage is missing/skipped
                                    send(MeetingProcessState.Success("Processing complete", null))
                                }
                            } else if (statusDto.overallStatus == "failed") {
                                isCompleted = true
                                send(MeetingProcessState.Error("Pipeline execution failed"))
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
                if (shouldDelay && !isCompleted) {
                    kotlinx.coroutines.delay(5000) // Fallback interval 5s
                }
            } else {
                // MQTT healthy, just wait a bit before checking again
                kotlinx.coroutines.delay(1000)
            }
        }

        mqttJob.cancel()
    }
}
