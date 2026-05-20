package com.example.whisperandroid.data.remote.api

import com.example.whisperandroid.data.remote.dto.FinalizeRecordingRequestDto
import com.example.whisperandroid.data.remote.dto.RecordingResponseDto
import com.example.whisperandroid.data.remote.dto.SpeechResponseDto
import com.example.whisperandroid.data.remote.dto.UploadRecordingUrlRequestDto
import com.example.whisperandroid.data.remote.dto.UploadRecordingUrlResponseDto
import retrofit2.http.Body
import retrofit2.http.Header
import retrofit2.http.POST

interface RecordingsApi {
    @POST("/api/recordings/upload-url")
    suspend fun createUploadUrl(
        @Body request: UploadRecordingUrlRequestDto,
        @Header("Authorization") token: String
    ): SpeechResponseDto<UploadRecordingUrlResponseDto>

    @POST("/api/recordings/finalize")
    suspend fun finalizeRecording(
        @Body request: FinalizeRecordingRequestDto,
        @Header("Authorization") token: String
    ): SpeechResponseDto<RecordingResponseDto>
}
