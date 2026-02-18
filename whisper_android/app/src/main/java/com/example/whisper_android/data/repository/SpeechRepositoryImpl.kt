package com.example.whisper_android.data.repository

import android.util.Log
import com.example.whisper_android.common.util.getErrorMessage
import com.example.whisper_android.data.remote.api.SpeechApi
import com.example.whisper_android.data.remote.dto.TranscriptionResultText
import com.example.whisper_android.domain.repository.Resource
import com.example.whisper_android.domain.repository.SpeechRepository
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow
import okhttp3.MediaType.Companion.toMediaTypeOrNull
import okhttp3.MultipartBody
import okhttp3.RequestBody.Companion.asRequestBody
import java.io.File

class SpeechRepositoryImpl(
    private val api: SpeechApi
) : SpeechRepository {

    override suspend fun transcribeAudio(file: File, token: String, language: String): Flow<Resource<String>> = flow {
        emit(Resource.Loading())
        try {
            val requestFile = file.asRequestBody(getAudioMimeType(file).toMediaTypeOrNull())
            val body = MultipartBody.Part.createFormData("audio", file.name, requestFile)
            val languageBody = MultipartBody.Part.createFormData("language", language)
            
            val response = api.transcribeAudio(body, languageBody, "Bearer $token")
            val taskId = response.data?.taskId
            
            if (response.status && taskId != null) {
                emit(Resource.Success(taskId))
            } else {
                emit(Resource.Error(response.message))
            }
        } catch (e: retrofit2.HttpException) {
             emit(Resource.Error(e.getErrorMessage()))
        } catch (e: Exception) {
            emit(Resource.Error("Transcribe failed: ${e.message}"))
        }
    }

    private fun getAudioMimeType(file: File): String {
        return when (file.extension.lowercase()) {
            "wav" -> "audio/wav"
            "mp3" -> "audio/mpeg"
            "m4a" -> "audio/mp4"
            "aac" -> "audio/aac"
            "ogg" -> "audio/ogg"
            "flac" -> "audio/flac"
            else -> "application/octet-stream"
        }
    }

    override suspend fun pollTranscription(taskId: String, token: String): Flow<Resource<TranscriptionResultText>> = flow {
        emit(Resource.Loading())
        while (true) {
            try {
                val response = api.getTranscriptionStatus(taskId, "Bearer $token")
                val statusWrapper = response.data?.taskStatus
                val status = statusWrapper?.status?.lowercase()

                Log.d("SpeechRepo", "Polling Task $taskId: $status")

                when (status) {
                    "completed" -> {
                        val result = statusWrapper.result
                        if (result != null) {
                            emit(Resource.Success(result))
                            return@flow
                        } else {
                            emit(Resource.Error("Completed but no result found"))
                            return@flow
                        }
                    }
                    "failed" -> {
                        emit(Resource.Error("Transcription task failed"))
                        return@flow
                    }
                    else -> {
                        // Pending or Processing, continue polling
                        delay(2000)
                    }
                }
            } catch (e: Exception) {
                Log.e("SpeechRepo", "Polling error: ${e.message}")
                // Retry on error instead of failing immediately
                delay(2000)
            }
        }
    }
}
