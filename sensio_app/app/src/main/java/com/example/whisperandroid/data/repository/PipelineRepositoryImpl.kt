package com.example.whisperandroid.data.repository

import android.util.Log
import com.example.whisperandroid.data.remote.api.PipelineApi
import com.example.whisperandroid.data.remote.dto.PipelineStatusDto
import com.example.whisperandroid.domain.repository.PipelineRepository
import com.example.whisperandroid.domain.repository.Resource
import java.io.File
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow
import okhttp3.MediaType.Companion.toMediaTypeOrNull
import okhttp3.MultipartBody
import okhttp3.RequestBody.Companion.asRequestBody

class PipelineRepositoryImpl(
    private val api: PipelineApi
) : PipelineRepository {
    override suspend fun executePipeline(
        audioFile: File,
        language: String?,
        targetLanguage: String?,
        summarize: Boolean,
        refine: Boolean?,
        diarize: Boolean,
        context: String?,
        style: String?,
        date: String?,
        location: String?,
        participants: String?,
        macAddress: String?,
        token: String,
        idempotencyKey: String?
    ): Flow<Resource<String>> = flow {
        emit(Resource.Loading())
        try {
            val requestFile = audioFile.asRequestBody("audio/*".toMediaTypeOrNull())
            val audioPart = MultipartBody.Part.createFormData("audio", audioFile.name, requestFile)

            val response = api.executePipeline(
                audio = audioPart,
                language = language,
                targetLanguage = targetLanguage,
                summarize = summarize,
                refine = refine,
                diarize = diarize,
                context = context,
                style = style,
                date = date,
                location = location,
                participants = participants,
                macAddress = macAddress,
                token = "Bearer $token",
                idempotencyKey = idempotencyKey
            )

            val taskId = response.data?.taskId
            if (response.status && taskId != null) {
                emit(Resource.Success(taskId))
            } else {
                emit(Resource.Error(response.message))
            }
        } catch (e: Exception) {
            Log.e("PipelineRepo", "Execute error: ${e.message}")
            emit(Resource.Error("Pipeline execution failed: ${e.message}"))
        }
    }

    override suspend fun pollPipelineStatus(
        taskId: String,
        token: String
    ): Flow<Resource<PipelineStatusDto>> = flow {
        emit(Resource.Loading())
        try {
            val response = api.getPipelineStatus(taskId, "Bearer $token")
            val statusData = response.data
            if (response.status && statusData != null) {
                emit(Resource.Success(statusData))
            } else {
                emit(Resource.Error(response.message))
            }
        } catch (e: Exception) {
            Log.e("PipelineRepo", "Poll error: ${e.message}")
            emit(Resource.Error("Poll error: ${e.message}"))
        }
    }

    override suspend fun executePipelineByUpload(
        sessionId: String,
        token: String,
        language: String?,
        targetLanguage: String?,
        summarize: Boolean,
        refine: Boolean?,
        diarize: Boolean,
        context: String?,
        style: String?,
        date: String?,
        location: String?,
        participants: List<String>?,
        macAddress: String?,
        idempotencyKey: String?
    ): Flow<Resource<String>> = flow {
        emit(Resource.Loading())
        try {
            val response = api.executePipelineByUpload(
                com.example.whisperandroid.data.remote.dto.PipelineSubmitByUploadRequestDto(
                    sessionId = sessionId,
                    language = language,
                    targetLanguage = targetLanguage,
                    summarize = summarize,
                    refine = refine,
                    diarize = diarize,
                    context = context,
                    style = style,
                    date = date,
                    location = location,
                    participants = participants,
                    macAddress = macAddress,
                    idempotencyKey = idempotencyKey
                ),
                "Bearer $token"
            )
            val taskId = response.data?.taskId
            if (response.status && taskId != null) {
                emit(Resource.Success(taskId))
            } else {
                emit(Resource.Error(response.message))
            }
        } catch (e: Exception) {
            emit(Resource.Error("Pipeline submission failed: ${e.message}"))
        }
    }
}
