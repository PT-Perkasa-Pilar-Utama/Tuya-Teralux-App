package com.example.whisperandroid.data.repository

import com.example.whisperandroid.common.util.getErrorMessage
import com.example.whisperandroid.data.remote.api.EmailApi
import com.example.whisperandroid.data.remote.dto.SendEmailByMacRequestDto
import com.example.whisperandroid.data.remote.dto.SendEmailRequestDto
import com.example.whisperandroid.domain.repository.EmailRepository
import com.example.whisperandroid.domain.repository.Resource
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow
import retrofit2.HttpException

class EmailRepositoryImpl(
    private val api: EmailApi
) : EmailRepository {
    override suspend fun sendEmailByMac(
        macAddress: String,
        subject: String,
        template: String,
        token: String,
        attachmentPath: String?,
        overrideEmails: List<String>?
    ): Flow<Resource<Boolean>> = flow {
        emit(Resource.Loading())
        try {
            val dataMap = if (overrideEmails != null && overrideEmails.isNotEmpty()) {
                mapOf("email" to overrideEmails)
            } else {
                null
            }

            val request = SendEmailByMacRequestDto(
                subject = subject,
                template = template,
                data = dataMap,
                attachmentPath = attachmentPath
            )
            val response = api.sendEmailByMac("Bearer $token", macAddress, request)

            if (response.status && response.data != null) {
                val taskId = response.data.taskId
                pollEmailStatus(taskId, token).collect { pollResource ->
                    emit(pollResource)
                }
            } else {
                emit(Resource.Error(response.message))
            }
        } catch (e: HttpException) {
            emit(Resource.Error(e.getErrorMessage()))
        } catch (e: Exception) {
            emit(Resource.Error(e.localizedMessage ?: "Unknown error occurred"))
        }
    }

    override suspend fun sendEmail(
        to: List<String>,
        subject: String,
        template: String,
        token: String,
        attachmentPath: String?
    ): Flow<Resource<Boolean>> = flow {
        emit(Resource.Loading())
        try {
            val request = SendEmailRequestDto(
                to = to,
                subject = subject,
                template = template,
                attachmentPath = attachmentPath
            )
            val response = api.sendEmail("Bearer $token", request)

            if (response.status && response.data != null) {
                // Return the taskId via the store, then poll
                val taskId = response.data.taskId
                pollEmailStatus(taskId, token).collect { pollResource ->
                    emit(pollResource)
                }
            } else {
                emit(Resource.Error(response.message))
            }
        } catch (e: HttpException) {
            emit(Resource.Error(e.getErrorMessage()))
        } catch (e: Exception) {
            emit(Resource.Error(e.localizedMessage ?: "Unknown error occurred"))
        }
    }

    override suspend fun pollEmailStatus(
        taskId: String,
        token: String
    ): Flow<Resource<Boolean>> = flow {
        emit(Resource.Loading())
        var attempts = 0
        val maxAttempts = 30 // 60 seconds

        while (attempts < maxAttempts) {
            try {
                val response = api.getEmailStatus("Bearer $token", taskId)
                if (response.status && response.data != null) {
                    val status = response.data.status
                    when (status) {
                        "completed" -> {
                            emit(Resource.Success(true))
                            return@flow
                        }
                        "failed" -> {
                            emit(Resource.Error(response.data.error ?: "Email task failed"))
                            return@flow
                        }
                        else -> {
                            // pending or sending — keep polling
                            emit(Resource.Loading())
                        }
                    }
                } else {
                    emit(Resource.Error(response.message))
                    return@flow
                }
            } catch (e: HttpException) {
                emit(Resource.Error(e.getErrorMessage()))
                return@flow
            } catch (e: Exception) {
                emit(Resource.Error(e.localizedMessage ?: "Connection error"))
                return@flow
            }

            delay(2000)
            attempts++
        }

        emit(Resource.Error("Polling timeout"))
    }
}
