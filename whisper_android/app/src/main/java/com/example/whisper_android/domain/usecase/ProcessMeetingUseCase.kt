package com.example.whisper_android.domain.usecase

import com.example.whisper_android.data.remote.dto.SpeechResponseDto
import com.example.whisper_android.data.remote.dto.TranscriptionResultText
import com.example.whisper_android.domain.repository.RagRepository
import com.example.whisper_android.domain.repository.Resource
import com.example.whisper_android.domain.repository.SpeechRepository
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow
import java.io.File

sealed class MeetingProcessState {
    object Idle : MeetingProcessState()
    object Recording : MeetingProcessState()
    object Uploading : MeetingProcessState()
    object Transcribing : MeetingProcessState()
    object Translating : MeetingProcessState()
    object Summarizing : MeetingProcessState()
    data class Success(val summary: String, val pdfUrl: String?) : MeetingProcessState()
    data class Error(val message: String) : MeetingProcessState()
}

class ProcessMeetingUseCase(
    private val transcribeAudioUseCase: TranscribeAudioUseCase,
    private val translateTextUseCase: TranslateTextUseCase,
    private val summarizeTextUseCase: SummarizeTextUseCase
) {
    suspend operator fun invoke(audioFile: File, token: String, targetLang: String = "English"): Flow<MeetingProcessState> = flow {
        emit(MeetingProcessState.Uploading)
        
        // 1. Transcribe
        var transcriptionText: String? = null
        transcribeAudioUseCase(audioFile, token, "id").collect { result ->
            when (result) {
                is Resource.Loading -> emit(MeetingProcessState.Transcribing)
                is Resource.Success -> transcriptionText = result.data
                is Resource.Error -> {
                    emit(MeetingProcessState.Error("Transcription failed: ${result.message}"))
                    return@collect
                }
            }
        }
        
        if (transcriptionText == null) return@flow

        // 2. Translate
        val translateTarget = when(targetLang.lowercase()) {
            "id", "indonesia" -> "Indonesia"
            else -> "English"
        }

        emit(MeetingProcessState.Translating)
        var translatedText: String? = null
        translateTextUseCase(transcriptionText!!, translateTarget, token).collect { result ->
            when (result) {
                is Resource.Loading -> emit(MeetingProcessState.Translating)
                is Resource.Success -> translatedText = result.data
                is Resource.Error -> {
                    emit(MeetingProcessState.Error("Translation failed: ${result.message}"))
                    return@collect
                }
            }
        }

        if (translatedText == null) return@flow

        // 3. Summarize
        emit(MeetingProcessState.Summarizing)
        summarizeTextUseCase(translatedText!!, targetLang.lowercase(), "meeting_minutes", token).collect { result ->
            when (result) {
                is Resource.Loading -> emit(MeetingProcessState.Summarizing)
                is Resource.Success -> {
                    val summaryData = result.data
                    if (summaryData != null) {
                        emit(MeetingProcessState.Success(summaryData.summary, summaryData.pdfUrl))
                    }
                }
                is Resource.Error -> emit(MeetingProcessState.Error("Summary failed: ${result.message}"))
            }
        }
    }
}
