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
    private val speechRepository: SpeechRepository,
    private val ragRepository: RagRepository
) {
    suspend operator fun invoke(audioFile: File, token: String, targetLang: String = "English"): Flow<MeetingProcessState> = flow {
        emit(MeetingProcessState.Uploading)
        
        // 1. Transcribe (Get Task ID)
        var transcriptionTaskId: String? = null
        speechRepository.transcribeAudio(audioFile, token, targetLang.lowercase()).collect { result ->
            when (result) {
                is Resource.Success -> transcriptionTaskId = result.data
                is Resource.Error -> {
                    emit(MeetingProcessState.Error("Upload failed: ${result.message}"))
                    return@collect
                }
                is Resource.Loading -> emit(MeetingProcessState.Uploading)
            }
        }
        
        if (transcriptionTaskId == null) return@flow

        // 2. Poll Transcription
        emit(MeetingProcessState.Transcribing)
        var transcriptionText: String? = null
        speechRepository.pollTranscription(transcriptionTaskId!!, token).collect { result ->
            when (result) {
                is Resource.Success -> transcriptionText = result.data?.transcription // Or refinedText?
                is Resource.Error -> {
                    emit(MeetingProcessState.Error("Transcription failed: ${result.message}"))
                    return@collect
                }
                is Resource.Loading -> emit(MeetingProcessState.Transcribing)
            }
        }

        if (transcriptionText == null) return@flow

        // 3. Translate
        // Map "en"/"id" to "English"/"Indonesia"
        val translateTarget = when(targetLang.lowercase()) {
            "id", "indonesia" -> "Indonesia"
            "en", "english" -> "English"
            else -> "English"
        }

        emit(MeetingProcessState.Translating)
        var translationTaskId: String? = null
        ragRepository.translateAsync(transcriptionText!!, translateTarget, token).collect { result ->
             when (result) {
                is Resource.Success -> translationTaskId = result.data
                is Resource.Error -> {
                    emit(MeetingProcessState.Error("Translation request failed: ${result.message}"))
                    return@collect
                }
                is Resource.Loading -> emit(MeetingProcessState.Translating)
            }
        }

        if (translationTaskId == null) return@flow

        // 4. Poll Translation
        var translatedText: String? = null
        ragRepository.pollTranslation(translationTaskId!!, token).collect { result ->
             when (result) {
                is Resource.Success -> translatedText = result.data
                is Resource.Error -> {
                    emit(MeetingProcessState.Error("Translation failed: ${result.message}"))
                    return@collect
                }
                is Resource.Loading -> emit(MeetingProcessState.Translating)
            }
        }

        if (translatedText == null) return@flow

        if (translatedText == null) return@flow

        // 5. Summary
        emit(MeetingProcessState.Summarizing)
        var summaryTaskId: String? = null
        // Use "meeting_minutes" style as default
        ragRepository.generateSummaryAsync(translatedText!!, "meeting_minutes", null, token).collect { result ->
             when (result) {
                is Resource.Success -> summaryTaskId = result.data
                is Resource.Error -> {
                    emit(MeetingProcessState.Error("Summary request failed: ${result.message}"))
                    return@collect
                }
                is Resource.Loading -> emit(MeetingProcessState.Summarizing)
            }
        }

        if (summaryTaskId == null) return@flow

        // 6. Poll Summary
        ragRepository.pollSummary(summaryTaskId!!, token).collect { result ->
             when (result) {
                is Resource.Success -> {
                    val summaryData = result.data
                    if (summaryData != null) {
                        emit(MeetingProcessState.Success(summaryData.summary, summaryData.pdfUrl))
                    }
                }
                is Resource.Error -> emit(MeetingProcessState.Error("Summary generation failed: ${result.message}"))
                is Resource.Loading -> emit(MeetingProcessState.Summarizing)
            }
        }
    }
}
