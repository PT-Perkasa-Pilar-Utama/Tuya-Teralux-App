package com.example.whisperandroid.domain.usecase

import com.example.whisperandroid.domain.model.TranscriptionPollingOutcome
import com.example.whisperandroid.domain.repository.Resource
import com.example.whisperandroid.domain.repository.WhisperRepository
import java.io.File
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow

class TranscribeAudioUseCase(
    private val whisperRepository: WhisperRepository
) {
    suspend fun initiate(
        audioFile: File,
        token: String,
        language: String,
        macAddress: String? = null,
        idempotencyKey: String? = null
    ): Flow<Resource<String>> =
        flow {
            emit(Resource.Loading())
            whisperRepository.transcribeAudio(
                file = audioFile,
                token = token,
                language = language,
                macAddress = macAddress,
                idempotencyKey = idempotencyKey
            ).collect { result ->
                emit(result)
            }
        }

    suspend fun getResult(
        taskId: String,
        token: String
    ): Flow<TranscriptionPollingOutcome> =
        flow {
            emit(TranscriptionPollingOutcome.Pending)
            whisperRepository.pollTranscription(taskId, token).collect { result ->
                when (result) {
                    is Resource.Success -> {
                        val statusDto = result.data
                        val statusResult = statusDto?.result
                        val status = statusDto?.status?.lowercase()

                        when (status) {
                            "completed" -> {
                                // Prefer refined_text over raw transcription
                                val text = statusResult?.refinedText
                                    ?.takeIf { it.isNotBlank() }
                                    ?: statusResult?.transcription

                                if (!text.isNullOrBlank()) {
                                    emit(TranscriptionPollingOutcome.Completed(text))
                                } else {
                                    // Check ASR quality gate metadata for rejection
                                    val isRejected = statusResult?.transcriptValid == false
                                    val rejectionReason = statusResult?.transcriptRejectionReason
                                    val audioClass = statusResult?.audioClass
                                    val providerSkipped = statusResult?.providerSkipped

                                    if (isRejected && !rejectionReason.isNullOrBlank()) {
                                        emit(
                                            TranscriptionPollingOutcome.Rejected(
                                                reason = rejectionReason,
                                                audioClass = audioClass,
                                                providerSkipped = providerSkipped
                                            )
                                        )
                                    } else {
                                        // Completed but no usable text and no rejection metadata
                                        emit(
                                            TranscriptionPollingOutcome.Failed(
                                                "Completed but no usable transcript"
                                            )
                                        )
                                    }
                                }
                            }

                            "failed" -> {
                                emit(
                                    TranscriptionPollingOutcome.Failed(
                                        statusDto.error ?: "Transcription task failed"
                                    )
                                )
                            }

                            else -> {
                                // Pending or Processing
                                emit(TranscriptionPollingOutcome.Pending)
                            }
                        }
                    }

                    is Resource.Error -> {
                        emit(TranscriptionPollingOutcome.Failed(result.message ?: "Transcription fetch failed"))
                    }

                    is Resource.Loading -> {
                        emit(TranscriptionPollingOutcome.Pending)
                    }
                }
            }
        }
}
