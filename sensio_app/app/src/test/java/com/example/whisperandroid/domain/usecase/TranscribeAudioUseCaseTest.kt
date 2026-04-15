package com.example.whisperandroid.domain.usecase

import com.example.whisperandroid.data.remote.dto.TranscriptionResultText
import com.example.whisperandroid.data.remote.dto.TranscriptionStatusDto
import com.example.whisperandroid.domain.model.TranscriptionPollingOutcome
import com.example.whisperandroid.domain.repository.Resource
import com.example.whisperandroid.domain.repository.WhisperRepository
import junit.framework.TestCase.assertEquals
import junit.framework.TestCase.assertTrue
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow
import kotlinx.coroutines.test.runTest
import org.junit.Test

class TranscribeAudioUseCaseTest {

    @Test
    fun `getResult with completed status and transcription text returns Completed`() = runTest {
        // Arrange
        val expectedText = "Hello world"
        val repository = createMockRepository(
            status = "completed",
            transcription = expectedText,
            refinedText = null
        )
        val useCase = TranscribeAudioUseCase(repository)

        // Act
        val results = mutableListOf<TranscriptionPollingOutcome>()
        useCase.getResult("task-123", "token").collect { results.add(it) }

        // Assert
        val completedOutcome = results.find { it is TranscriptionPollingOutcome.Completed }
        assertTrue("Should have Completed outcome", completedOutcome is TranscriptionPollingOutcome.Completed)
        assertEquals(expectedText, (completedOutcome as TranscriptionPollingOutcome.Completed).text)
    }

    @Test
    fun `getResult with completed status and refined_text prefers refined_text over transcription`() = runTest {
        // Arrange
        val rawTranscription = "hello"
        val refinedText = "Hello world"
        val repository = createMockRepository(
            status = "completed",
            transcription = rawTranscription,
            refinedText = refinedText
        )
        val useCase = TranscribeAudioUseCase(repository)

        // Act
        val results = mutableListOf<TranscriptionPollingOutcome>()
        useCase.getResult("task-123", "token").collect { results.add(it) }

        // Assert
        val completedOutcome = results.find { it is TranscriptionPollingOutcome.Completed }
        assertTrue("Should have Completed outcome", completedOutcome is TranscriptionPollingOutcome.Completed)
        assertEquals(refinedText, (completedOutcome as TranscriptionPollingOutcome.Completed).text)
    }

    @Test
    fun `getResult with completed status, blank text, and transcript_valid=false returns Rejected`() = runTest {
        // Arrange
        val rejectionReason = "empty_after_normalization"
        val audioClass = "silent"
        val repository = createMockRepositoryWithRejection(
            status = "completed",
            transcriptValid = false,
            transcriptRejectionReason = rejectionReason,
            audioClass = audioClass,
            providerSkipped = true
        )
        val useCase = TranscribeAudioUseCase(repository)

        // Act
        val results = mutableListOf<TranscriptionPollingOutcome>()
        useCase.getResult("task-123", "token").collect { results.add(it) }

        // Assert
        val rejectedOutcome = results.find { it is TranscriptionPollingOutcome.Rejected }
        assertTrue("Should have Rejected outcome", rejectedOutcome is TranscriptionPollingOutcome.Rejected)
        val rejected = rejectedOutcome as TranscriptionPollingOutcome.Rejected
        assertEquals(rejectionReason, rejected.reason)
        assertEquals(audioClass, rejected.audioClass)
        assertEquals(true, rejected.providerSkipped)
    }

    @Test
    fun `getResult with completed status, blank text, and no rejection metadata returns Failed`() = runTest {
        // Arrange
        val repository = createMockRepository(
            status = "completed",
            transcription = "",
            refinedText = null
        )
        val useCase = TranscribeAudioUseCase(repository)

        // Act
        val results = mutableListOf<TranscriptionPollingOutcome>()
        useCase.getResult("task-123", "token").collect { results.add(it) }

        // Assert
        val failedOutcome = results.find { it is TranscriptionPollingOutcome.Failed }
        assertTrue("Should have Failed outcome", failedOutcome is TranscriptionPollingOutcome.Failed)
        assertTrue(
            (failedOutcome as TranscriptionPollingOutcome.Failed).message.contains(
                "no usable transcript",
                ignoreCase = true
            )
        )
    }

    @Test
    fun `getResult with failed status returns Failed`() = runTest {
        // Arrange
        val errorMessage = "Service unavailable"
        val repository = createMockRepositoryWithFailure(
            status = "failed",
            error = errorMessage
        )
        val useCase = TranscribeAudioUseCase(repository)

        // Act
        val results = mutableListOf<TranscriptionPollingOutcome>()
        useCase.getResult("task-123", "token").collect { results.add(it) }

        // Assert
        val failedOutcome = results.find { it is TranscriptionPollingOutcome.Failed }
        assertTrue("Should have Failed outcome", failedOutcome is TranscriptionPollingOutcome.Failed)
        assertEquals(errorMessage, (failedOutcome as TranscriptionPollingOutcome.Failed).message)
    }

    @Test
    fun `getResult with pending status returns Pending`() = runTest {
        // Arrange
        val repository = createMockRepository(
            status = "pending",
            transcription = "",
            refinedText = null
        )
        val useCase = TranscribeAudioUseCase(repository)

        // Act
        val results = mutableListOf<TranscriptionPollingOutcome>()
        useCase.getResult("task-123", "token").collect { results.add(it) }

        // Assert
        val pendingOutcome = results.find { it is TranscriptionPollingOutcome.Pending }
        assertTrue("Should have Pending outcome", pendingOutcome is TranscriptionPollingOutcome.Pending)
    }

    @Test
    fun `getResult with repository error returns Failed`() = runTest {
        // Arrange
        val errorMessage = "Network error"
        val repository = object : WhisperRepository {
            override suspend fun transcribeAudio(
                file: java.io.File,
                token: String,
                language: String,
                macAddress: String?,
                idempotencyKey: String?
            ): Flow<Resource<String>> = flow { emit(Resource.Error("Not used")) }

            override suspend fun pollTranscription(
                taskId: String,
                token: String
            ): Flow<Resource<TranscriptionStatusDto>> = flow {
                emit(Resource.Loading())
                emit(Resource.Error(errorMessage))
            }

            override suspend fun getTranscriptionStatus(
                taskId: String,
                token: String
            ) = throw UnsupportedOperationException("Not used in test")

            override suspend fun transcribeByUpload(
                sessionId: String,
                token: String,
                language: String,
                macAddress: String?,
                idempotencyKey: String?,
                diarize: Boolean
            ): Flow<Resource<String>> = flow { emit(Resource.Error("Not used")) }
        }
        val useCase = TranscribeAudioUseCase(repository)

        // Act
        val results = mutableListOf<TranscriptionPollingOutcome>()
        useCase.getResult("task-123", "token").collect { results.add(it) }

        // Assert
        val failedOutcome = results.find { it is TranscriptionPollingOutcome.Failed }
        assertTrue("Should have Failed outcome", failedOutcome is TranscriptionPollingOutcome.Failed)
        assertEquals(errorMessage, (failedOutcome as TranscriptionPollingOutcome.Failed).message)
    }

    @Test
    fun `getResult with completed status and hallucination rejection returns Rejected`() = runTest {
        // Arrange
        val rejectionReason = "known_hallucination_phrase"
        val audioClass = "active"
        val repository = createMockRepositoryWithRejection(
            status = "completed",
            transcriptValid = false,
            transcriptRejectionReason = rejectionReason,
            audioClass = audioClass,
            providerSkipped = false
        )
        val useCase = TranscribeAudioUseCase(repository)

        // Act
        val results = mutableListOf<TranscriptionPollingOutcome>()
        useCase.getResult("task-123", "token").collect { results.add(it) }

        // Assert
        val rejectedOutcome = results.find { it is TranscriptionPollingOutcome.Rejected }
        assertTrue("Should have Rejected outcome", rejectedOutcome is TranscriptionPollingOutcome.Rejected)
        val rejected = rejectedOutcome as TranscriptionPollingOutcome.Rejected
        assertEquals(rejectionReason, rejected.reason)
        assertEquals(audioClass, rejected.audioClass)
        assertEquals(false, rejected.providerSkipped)
    }

    // Helper functions

    private fun createMockRepository(
        status: String,
        transcription: String,
        refinedText: String?
    ): WhisperRepository {
        return object : WhisperRepository {
            override suspend fun transcribeAudio(
                file: java.io.File,
                token: String,
                language: String,
                macAddress: String?,
                idempotencyKey: String?
            ): Flow<Resource<String>> = flow { emit(Resource.Error("Not used in this test")) }

            override suspend fun pollTranscription(
                taskId: String,
                token: String
            ): Flow<Resource<TranscriptionStatusDto>> = flow {
                emit(Resource.Loading())
                emit(
                    Resource.Success(
                        TranscriptionStatusDto(
                            status = status,
                            result = TranscriptionResultText(
                                transcription = transcription,
                                refinedText = refinedText
                            ),
                            error = null
                        )
                    )
                )
            }

            override suspend fun getTranscriptionStatus(
                taskId: String,
                token: String
            ) = throw UnsupportedOperationException("Not used in test")

            override suspend fun transcribeByUpload(
                sessionId: String,
                token: String,
                language: String,
                macAddress: String?,
                idempotencyKey: String?,
                diarize: Boolean
            ): Flow<Resource<String>> = flow { emit(Resource.Error("Not used in this test")) }
        }
    }

    private fun createMockRepositoryWithRejection(
        status: String,
        transcriptValid: Boolean,
        transcriptRejectionReason: String,
        audioClass: String,
        providerSkipped: Boolean
    ): WhisperRepository {
        return object : WhisperRepository {
            override suspend fun transcribeAudio(
                file: java.io.File,
                token: String,
                language: String,
                macAddress: String?,
                idempotencyKey: String?
            ): Flow<Resource<String>> = flow { emit(Resource.Error("Not used in this test")) }

            override suspend fun pollTranscription(
                taskId: String,
                token: String
            ): Flow<Resource<TranscriptionStatusDto>> = flow {
                emit(Resource.Loading())
                emit(
                    Resource.Success(
                        TranscriptionStatusDto(
                            status = status,
                            result = TranscriptionResultText(
                                transcription = "",
                                refinedText = null,
                                transcriptValid = transcriptValid,
                                transcriptRejectionReason = transcriptRejectionReason,
                                audioClass = audioClass,
                                providerSkipped = providerSkipped
                            ),
                            error = null
                        )
                    )
                )
            }

            override suspend fun getTranscriptionStatus(
                taskId: String,
                token: String
            ) = throw UnsupportedOperationException("Not used in test")

            override suspend fun transcribeByUpload(
                sessionId: String,
                token: String,
                language: String,
                macAddress: String?,
                idempotencyKey: String?,
                diarize: Boolean
            ): Flow<Resource<String>> = flow { emit(Resource.Error("Not used in this test")) }
        }
    }

    private fun createMockRepositoryWithFailure(
        status: String,
        error: String
    ): WhisperRepository {
        return object : WhisperRepository {
            override suspend fun transcribeAudio(
                file: java.io.File,
                token: String,
                language: String,
                macAddress: String?,
                idempotencyKey: String?
            ): Flow<Resource<String>> = flow { emit(Resource.Error("Not used in this test")) }

            override suspend fun pollTranscription(
                taskId: String,
                token: String
            ): Flow<Resource<TranscriptionStatusDto>> = flow {
                emit(Resource.Loading())
                emit(
                    Resource.Success(
                        TranscriptionStatusDto(
                            status = status,
                            result = null,
                            error = error
                        )
                    )
                )
            }

            override suspend fun getTranscriptionStatus(
                taskId: String,
                token: String
            ) = throw UnsupportedOperationException("Not used in test")

            override suspend fun transcribeByUpload(
                sessionId: String,
                token: String,
                language: String,
                macAddress: String?,
                idempotencyKey: String?,
                diarize: Boolean
            ): Flow<Resource<String>> = flow { emit(Resource.Error("Not used in this test")) }
        }
    }
}
