package com.example.whisper_android.domain.usecase

import com.example.whisper_android.domain.repository.Resource
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
    private val speechRepository: com.example.whisper_android.domain.repository.SpeechRepository,
    private val ragRepository: com.example.whisper_android.domain.repository.RagRepository,
    private val transcribeAudioUseCase: TranscribeAudioUseCase,
    private val translateTextUseCase: TranslateTextUseCase,
    private val summarizeTextUseCase: SummarizeTextUseCase
) {
    suspend operator fun invoke(
        audioFile: File,
        token: String,
        targetLang: String = "English",
        macAddress: String? = null,
        waitSignal: suspend (String) -> Unit // message to wait for
    ): Flow<MeetingProcessState> =
        flow {
            emit(MeetingProcessState.Uploading)

            // 1. Transcribe
            var transcribeTaskId: String? = null
            transcribeAudioUseCase.initiate(audioFile, token, "id", macAddress).collect { result ->
                when (result) {
                    is Resource.Success -> {
                        transcribeTaskId = result.data
                    }
                    is Resource.Error -> {
                        emit(
                            MeetingProcessState.Error(
                                "Transcription initiation failed: ${result.message}"
                            )
                        )
                    }
                    else -> {}
                }
            }

            if (transcribeTaskId == null) return@flow

            emit(MeetingProcessState.Transcribing(transcribeTaskId!!))
            waitSignal("Transcribe")

            var transcriptionText: String? = null
            transcribeAudioUseCase.getResult(transcribeTaskId!!, token).collect { result ->
                when (result) {
                    is Resource.Success -> {
                        transcriptionText = result.data
                    }
                    is Resource.Error -> {
                        emit(
                            MeetingProcessState.Error(
                                "Transcription fetch failed: ${result.message}"
                            )
                        )
                    }
                    else -> {}
                }
            }

            if (transcriptionText == null) return@flow

            // 2. Translate
            val translateTarget =
                when (targetLang.lowercase()) {
                    "id", "indonesia" -> "Indonesia"
                    else -> "English"
                }

            var translateTaskId: String? = null
            translateTextUseCase.initiate(transcriptionText!!, translateTarget, macAddress, token).collect { result ->
                when (result) {
                    is Resource.Success -> {
                        translateTaskId = result.data
                    }
                    is Resource.Error -> {
                        emit(
                            MeetingProcessState.Error(
                                "Translation initiation failed: ${result.message}"
                            )
                        )
                    }
                    else -> {}
                }
            }

            if (translateTaskId == null) return@flow

            emit(MeetingProcessState.Translating(translateTaskId!!))
            waitSignal("RAG")

            var translatedText: String? = null
            translateTextUseCase.getResult(translateTaskId!!, token).collect { result ->
                when (result) {
                    is Resource.Success -> {
                        translatedText = result.data
                    }
                    is Resource.Error -> {
                        emit(
                            MeetingProcessState.Error("Translation fetch failed: ${result.message}")
                        )
                    }
                    else -> {}
                }
            }

            if (translatedText == null) return@flow

            // 3. Summarize
            var summarizeTaskId: String? = null
            summarizeTextUseCase.initiate(
                translatedText!!,
                targetLang.lowercase(),
                "meeting_minutes",
                macAddress,
                token
            ).collect { result ->
                when (result) {
                    is Resource.Success -> {
                        summarizeTaskId = result.data
                    }
                    is Resource.Error -> {
                        emit(
                            MeetingProcessState.Error(
                                "Summary initiation failed: ${result.message}"
                            )
                        )
                    }
                    else -> {}
                }
            }

            if (summarizeTaskId == null) return@flow

            emit(MeetingProcessState.Summarizing(summarizeTaskId!!))
            waitSignal("RAG")

            summarizeTextUseCase.getResult(summarizeTaskId!!, token).collect { result ->
                when (result) {
                    is Resource.Success -> {
                        val summaryData = result.data
                        if (summaryData != null) {
                            emit(
                                MeetingProcessState.Success(summaryData.summary, summaryData.pdfUrl)
                            )
                        }
                    }
                    is Resource.Error -> {
                        emit(MeetingProcessState.Error("Summary fetch failed: ${result.message}"))
                    }
                    else -> {}
                }
            }
        }
}
