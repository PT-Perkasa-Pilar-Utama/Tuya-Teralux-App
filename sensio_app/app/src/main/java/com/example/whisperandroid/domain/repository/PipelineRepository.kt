package com.example.whisperandroid.domain.repository

import com.example.whisperandroid.data.remote.dto.PipelineStatusDto
import com.example.whisperandroid.domain.repository.Resource
import kotlinx.coroutines.flow.Flow
import java.io.File

interface PipelineRepository {
    suspend fun executePipeline(
        audioFile: File,
        language: String? = "id",
        targetLanguage: String? = "en",
        summarize: Boolean = true,
        refine: Boolean? = true,
        diarize: Boolean = false,
        context: String? = null,
        style: String? = "meeting_minutes",
        date: String? = null,
        location: String? = null,
        participants: String? = null,
        macAddress: String? = null,
        token: String,
        idempotencyKey: String? = null
    ): Flow<Resource<String>>

    suspend fun pollPipelineStatus(
        taskId: String,
        token: String
    ): Flow<Resource<PipelineStatusDto>>
}
