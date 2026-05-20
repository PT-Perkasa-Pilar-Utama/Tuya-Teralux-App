package com.example.whisperandroid.data.repository

import com.example.whisperandroid.data.remote.ProgressRequestBody
import com.example.whisperandroid.data.remote.api.RecordingsApi
import com.example.whisperandroid.data.remote.dto.FinalizeRecordingRequestDto
import com.example.whisperandroid.data.remote.dto.UploadRecordingUrlRequestDto
import com.example.whisperandroid.domain.repository.RecordingUploadRepository
import com.example.whisperandroid.domain.repository.RecordingUploadState
import java.io.File
import kotlinx.coroutines.channels.trySendBlocking
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.channelFlow
import okhttp3.MediaType.Companion.toMediaTypeOrNull
import okhttp3.OkHttpClient
import okhttp3.Request

class RecordingUploadRepositoryImpl(
    private val recordingsApi: RecordingsApi,
    private val uploadClient: OkHttpClient
) : RecordingUploadRepository {
    override fun uploadRecording(
        file: File,
        token: String,
        macAddress: String,
        contentType: String
    ): Flow<RecordingUploadState> = channelFlow {
        trySendBlocking(RecordingUploadState.Loading("Preparing upload..."))

        val uploadUrlResponse = recordingsApi.createUploadUrl(
            request = UploadRecordingUrlRequestDto(
                filename = file.name,
                contentType = contentType
            ),
            token = "Bearer $token"
        )

        val uploadData = uploadUrlResponse.data
            ?: throw IllegalStateException(uploadUrlResponse.message ?: "Failed to create upload URL")

        file.inputStream().use { inputStream ->
            val requestBody = ProgressRequestBody(
                inputStream = inputStream,
                contentType = contentType.toMediaTypeOrNull(),
                contentLength = file.length(),
                onProgress = { uploadedBytes ->
                    val percent = if (file.length() > 0L) {
                        (uploadedBytes.toFloat() / file.length().toFloat()) * 100f
                    } else {
                        0f
                    }
                    trySendBlocking(RecordingUploadState.Progress(uploadedBytes, file.length(), percent))
                }
            )

            val request = Request.Builder()
                .url(uploadData.uploadUrl)
                .put(requestBody)
                .header("Content-Type", contentType)
                .build()

            uploadClient.newCall(request).execute().use { response ->
                if (!response.isSuccessful) {
                    throw IllegalStateException("Upload failed with status ${response.code}")
                }
            }
        }

        trySendBlocking(RecordingUploadState.Finalizing(uploadData.objectKey))

        val finalizeResponse = recordingsApi.finalizeRecording(
            request = FinalizeRecordingRequestDto(
                filename = file.name,
                objectKey = uploadData.objectKey,
                macAddress = macAddress
            ),
            token = "Bearer $token"
        )

        val recording = finalizeResponse.data
            ?: throw IllegalStateException(finalizeResponse.message ?: "Failed to finalize recording")

        trySendBlocking(RecordingUploadState.Success(recording))
    }
}
