package com.example.whisperandroid.domain.usecase

import com.example.whisperandroid.domain.repository.Resource
import java.io.File
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow

sealed class MeetingProcessState {
    object Idle : MeetingProcessState()

    object Recording : MeetingProcessState()

    object Uploading : MeetingProcessState()

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
    private val pipelineRepository: com.example.whisperandroid.domain.repository.PipelineRepository
) {
    suspend operator fun invoke(
        audioFile: File,
        token: String,
        targetLang: String = "English",
        macAddress: String? = null,
        idempotencyKey: String? = null
    ): Flow<MeetingProcessState> =
        flow {
            emit(MeetingProcessState.Uploading)

            val targetLangCode = when (targetLang.lowercase()) {
                "id", "indonesia" -> "id"
                else -> "en"
            }

            var pipelineTaskId: String? = null
            pipelineRepository.executePipeline(
                audioFile = audioFile,
                language = "id", // assuming source is id
                targetLanguage = targetLangCode,
                summarize = true,
                refine = true,
                macAddress = macAddress,
                token = token,
                idempotencyKey = idempotencyKey
            ).collect { result ->
                when (result) {
                    is Resource.Success -> {
                        pipelineTaskId = result.data
                    }
                    is Resource.Error -> {
                        emit(MeetingProcessState.Error("Pipeline initiation failed: ${result.message}"))
                    }
                    else -> {}
                }
            }

            if (pipelineTaskId == null) return@flow

            // Polling
            var isCompleted = false
            while (!isCompleted) {
                pipelineRepository.pollPipelineStatus(pipelineTaskId!!, token).collect { result ->
                    when (result) {
                        is Resource.Success -> {
                            val statusDto = result.data!!
                            val stages = statusDto.stages ?: emptyMap()
                            
                            // Determine current visual state based on stages
                            // Stage keys from backend: transcription, refinement, translation, summary
                            val transcribeStatus = stages["transcription"]?.status
                            val translateStatus = stages["translation"]?.status
                            val summarizeStatus = stages["summary"]?.status

                            when {
                                summarizeStatus == "processing" -> emit(MeetingProcessState.Summarizing(pipelineTaskId!!))
                                translateStatus == "processing" -> emit(MeetingProcessState.Translating(pipelineTaskId!!))
                                transcribeStatus == "processing" || transcribeStatus == "pending" -> emit(MeetingProcessState.Transcribing(pipelineTaskId!!))
                            }

                            if (statusDto.overallStatus == "completed") {
                                isCompleted = true
                                val summaryStage = stages["summary"]
                                if (summaryStage?.status == "completed") {
                                    val resMap = summaryStage.result as? Map<*, *>
                                    val summary = resMap?.get("summary") as? String ?: "Meeting summary is ready"
                                    val pdfUrl = resMap?.get("pdf_url") as? String
                                    emit(MeetingProcessState.Success(summary, pdfUrl))
                                } else {
                                    emit(MeetingProcessState.Success("Processing complete", null))
                                }
                            } else if (statusDto.overallStatus == "failed") {
                                isCompleted = true
                                emit(MeetingProcessState.Error("Pipeline execution failed"))
                            }
                        }
                        is Resource.Error -> {
                            isCompleted = true
                            emit(MeetingProcessState.Error("Polling failed: ${result.message}"))
                        }
                        else -> {}
                    }
                }
                if (!isCompleted) {
                    kotlinx.coroutines.delay(2000)
                }
            }
        }
}
