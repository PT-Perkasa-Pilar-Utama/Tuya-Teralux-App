package com.example.whisper_android.data.repository

import android.util.Log
import com.example.whisper_android.common.util.getErrorMessage
import com.example.whisper_android.data.remote.api.SpeechApi
import com.example.whisper_android.data.remote.dto.TranscriptionResultText
import com.example.whisper_android.domain.repository.Resource
import com.example.whisper_android.domain.repository.SpeechRepository
import java.io.File
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow
import okhttp3.MediaType.Companion.toMediaTypeOrNull
import okhttp3.MultipartBody
import okhttp3.RequestBody.Companion.asRequestBody

class SpeechRepositoryImpl(
    private val api: SpeechApi
) : SpeechRepository {
    override suspend fun transcribeAudio(
        file: File,
        token: String,
        language: String,
        macAddress: String?
    ): Flow<Resource<String>> =
        flow {
            emit(Resource.Loading())
            try {
                val requestFile = file.asRequestBody(getAudioMimeType(file).toMediaTypeOrNull())
                val body = MultipartBody.Part.createFormData("audio", file.name, requestFile)
                val languageBody = MultipartBody.Part.createFormData("language", language)
                val macPart = macAddress?.let {
                    MultipartBody.Part.createFormData("mac_address", it)
                }

                val response = api.transcribeAudio(body, languageBody, macPart, "Bearer $token")
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

    private fun getAudioMimeType(file: File): String =
        when (file.extension.lowercase()) {
            "wav" -> "audio/wav"
            "mp3" -> "audio/mpeg"
            "m4a" -> "audio/mp4"
            "aac" -> "audio/aac"
            "ogg" -> "audio/ogg"
            "flac" -> "audio/flac"
            else -> "application/octet-stream"
        }

    override suspend fun pollTranscription(
        taskId: String,
        token: String
    ): Flow<Resource<TranscriptionResultText>> =
        flow {
            emit(Resource.Loading())
            try {
                val response = api.getTranscriptionStatus(taskId, "Bearer $token")
                val statusDto = response.data
                val status = statusDto?.status?.lowercase()

                Log.d("SpeechRepo", "Check Task $taskId: $status")

                when (status) {
                    "completed" -> {
                        val result = statusDto.result
                        if (result != null) {
                            emit(Resource.Success(result))
                        } else {
                            emit(Resource.Error("Completed but no result found"))
                        }
                    }

                    "failed" -> {
                        emit(Resource.Error(statusDto.error ?: "Transcription task failed"))
                    }

                    else -> {
                        // Pending or Processing, emit Loading once
                        emit(Resource.Loading())
                    }
                }
            } catch (e: Exception) {
                Log.e("SpeechRepo", "Check error: ${e.message}")
                emit(Resource.Error("Check error: ${e.message}"))
            }
        }
}
