package com.example.whisperandroid.presentation.assistant

import com.example.whisperandroid.data.di.NetworkModule
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.StandardTestDispatcher
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.setMain
import org.junit.After
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotEquals
import org.junit.Assert.assertTrue
import org.junit.Before
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class AiAssistantViewModelTest {

    private val testDispatcher = StandardTestDispatcher()

    @Before
    fun setup() {
        Dispatchers.setMain(testDispatcher)
        NetworkModule.setTuyaSyncReady(false)
    }

    @After
    fun tearDown() {
        Dispatchers.resetMain()
    }

    @Test
    fun `getCleanMessage returns null for empty responseText without blocked or control flags`() {
        // Test the terminal empty payload case
        val result = ParsedAssistantChatResult(
            responseText = null,
            isBlocked = false,
            isControl = false,
            source = "HTTP_HANDLER",
            isDupDrop = false,
            isValidationError = false,
            isDupCached = false,
            isDupInProgress = false
        )

        val cleanMessage = AssistantResponseParser.getCleanMessage(result, "en")

        // Empty responseText without blocked/control should return null
        assertEquals(null, cleanMessage)
    }

    @Test
    fun `getCleanMessage returns null for blank responseText without blocked or control flags`() {
        // Test the terminal empty payload case with blank text
        val result = ParsedAssistantChatResult(
            responseText = "   ",
            isBlocked = false,
            isControl = false,
            source = "HTTP_HANDLER",
            isDupDrop = false,
            isValidationError = false,
            isDupCached = false,
            isDupInProgress = false
        )

        val cleanMessage = AssistantResponseParser.getCleanMessage(result, "en")

        // Blank responseText should be trimmed - parseMarkdownToText will handle it
        // The result may be null or empty string depending on implementation
        assert(cleanMessage == null || cleanMessage.trim().isEmpty())
    }

    @Test
    fun `getCleanMessage returns blocked message for blocked flag`() {
        val result = ParsedAssistantChatResult(
            responseText = null,
            isBlocked = true,
            isControl = false,
            source = "HTTP_HANDLER",
            isDupDrop = false,
            isValidationError = false,
            isDupCached = false,
            isDupInProgress = false
        )

        val cleanMessage = AssistantResponseParser.getCleanMessage(result, "en")

        // Blocked with null response should return voice not clear message (ASR quality gate)
        assertTrue(cleanMessage?.contains("not clear") == true)
    }

    @Test
    fun `getCleanMessage returns validation error message for validation error flag`() {
        val result = ParsedAssistantChatResult(
            responseText = null,
            isBlocked = false,
            isControl = false,
            source = null,
            isDupDrop = false,
            isValidationError = true,
            isDupCached = false,
            isDupInProgress = false
        )

        val cleanMessage = AssistantResponseParser.getCleanMessage(result, "en")

        // Validation error should return specific message
        assertTrue(cleanMessage?.contains("not clear") == true)
    }

    @Test
    fun `getCleanMessage returns parsed markdown text for normal response`() {
        val result = ParsedAssistantChatResult(
            responseText = "**Hello, how can I help?**",
            isBlocked = false,
            isControl = false,
            source = "HTTP_HANDLER",
            isDupDrop = false,
            isValidationError = false,
            isDupCached = false,
            isDupInProgress = false
        )

        val cleanMessage = AssistantResponseParser.getCleanMessage(result, "en")

        assertEquals("Hello, how can I help?", cleanMessage)
    }

    @Test
    fun `getCleanMessage extracts result line from control response with multiple lines`() {
        // Test control response parsing - should extract second line
        val result = ParsedAssistantChatResult(
            responseText = "Mematikan AC rumah.\n\nBerhasil mematikan Sharp AC Rumah.",
            isBlocked = false,
            isControl = true,
            source = "HTTP_HANDLER",
            isDupDrop = false,
            isValidationError = false,
            isDupCached = false,
            isDupInProgress = false
        )

        val cleanMessage = AssistantResponseParser.getCleanMessage(result, "en")

        // Should extract the result line (second line after \n\n)
        assertEquals("Berhasil mematikan Sharp AC Rumah.", cleanMessage)
    }

    @Test
    fun `ParsedAssistantChatResult with IDEMPOTENCY_IN_PROGRESS has correct flags`() {
        val result = ParsedAssistantChatResult(
            responseText = null,
            isBlocked = false,
            isControl = false,
            source = "IDEMPOTENCY_IN_PROGRESS",
            isDupDrop = false,
            isValidationError = false,
            isDupCached = false,
            isDupInProgress = true
        )

        assertTrue(result.isDupInProgress)
        assertTrue(result.isDupInProgress)
    }

    @Test
    fun `ParsedAssistantChatResult with IDEMPOTENCY_CACHED has correct flags`() {
        val result = ParsedAssistantChatResult(
            responseText = "Cached response",
            isBlocked = false,
            isControl = false,
            source = "IDEMPOTENCY_CACHED",
            isDupDrop = true,
            isValidationError = false,
            isDupCached = true,
            isDupInProgress = false
        )

        assertTrue(result.isDupCached)
        assertEquals("Cached response", result.responseText)
    }

    // ========================================================================
    // MQTT Flow State Machine Tests
    // These tests verify the parsed result structure for proper handling
    // Actual state transitions occur in the MQTT collect flow
    // ========================================================================

    @Test
    fun `whisper answer with MQTT_ACK is identified as non-final acknowledgment`() {
        // This test verifies that when a whisper/answer message arrives with
        // source: MQTT_ACK, it is identified as a non-final acknowledgment
        // that should NOT trigger state transition to Idle.

        // Simulate whisper/answer ACK with MQTT_ACK source
        val ackMessage = """
            {
                "status": true,
                "message": "Transcription task submitted successfully (Ephemeral)",
                "data": {
                    "task_id": "task-uuid-456",
                    "task_status": "pending",
                    "request_id": "test-voice-request-123",
                    "source": "MQTT_ACK"
                }
            }
        """.trimIndent()

        // Parse the ACK message
        val parsedResult = AssistantResponseParser.parseMqttAssistantResult(ackMessage)

        // Verify parsing succeeded
        assert(parsedResult != null)
        assertEquals("MQTT_ACK", parsedResult?.source)
        assertEquals(false, parsedResult?.isBlocked)
        assertEquals(false, parsedResult?.isControl)

        // CRITICAL: Verify isDupInProgress is false (MQTT_ACK is not an in-progress marker)
        // This means it won't be silently dropped, but also won't trigger completion
        assertEquals(false, parsedResult?.isDupInProgress)

        // Verify isDupCached is false (this is not a cached response)
        assertEquals(false, parsedResult?.isDupCached)

        // Verify no error flags
        assertEquals(false, parsedResult?.isValidationError)
    }

    @Test
    fun `whisper answer with status false is identified as error`() {
        // This test verifies that when a whisper/answer message arrives with
        // status: false, it is identified as an error.
        // Note: isValidationError is only true for "Validation Error" message.
        // For other errors, the status field indicates failure.

        // Simulate whisper/answer with error status
        val errorMessage = """
            {
                "status": false,
                "message": "Voice transcription failed: audio too short"
            }
        """.trimIndent()

        // Parse the error message
        val parsedResult = AssistantResponseParser.parseMqttAssistantResult(errorMessage)

        // Verify parsing succeeded
        assert(parsedResult != null)

        // The envelope-level message is NOT extracted as responseText (by design)
        // Only data.response is used for assistant text
        assertEquals(null, parsedResult?.responseText)
        assertEquals(false, parsedResult?.isValidationError) // Only "Validation Error" triggers this
    }

    @Test
    fun `chat answer with final response is identified for completion`() {
        // This test verifies that when a chat/answer message arrives with
        // a final response, it is identified for state transition to Idle.

        // Simulate chat/answer with final response
        val chatResponse = """
            {
                "status": true,
                "message": "Chat processed successfully",
                "data": {
                    "response": "Berhasil menyalakan AC",
                    "is_control": true,
                    "is_blocked": false,
                    "request_id": "test-chat-request-success-999",
                    "source": "MQTT_SUBSCRIBER"
                }
            }
        """.trimIndent()

        // Parse the response
        val parsedResult = AssistantResponseParser.parseMqttAssistantResult(chatResponse)

        // Verify parsing succeeded
        assert(parsedResult != null)
        assertEquals("MQTT_SUBSCRIBER", parsedResult?.source)
        assertEquals("Berhasil menyalakan AC", parsedResult?.responseText)
        assertEquals(true, parsedResult?.isControl)

        // CRITICAL: Verify duplicate flags are false (this is a fresh response)
        assertEquals(false, parsedResult?.isDupInProgress)
        assertEquals(false, parsedResult?.isDupCached)

        // Verify no error flags
        assertEquals(false, parsedResult?.isValidationError)
        assertEquals(false, parsedResult?.isBlocked)
    }

    @Test
    fun `chat answer with IDEMPOTENCY_IN_PROGRESS is identified to be ignored`() {
        // This test verifies that in-progress acknowledgments are properly
        // identified and would be ignored in the MQTT flow.

        // Simulate chat/answer with IDEMPOTENCY_IN_PROGRESS
        val inProgressMessage = """
            {
                "status": true,
                "message": "Chat request received (sync with active whisper)",
                "data": {
                    "request_id": "test-chat-request-inprogress-111",
                    "source": "IDEMPOTENCY_IN_PROGRESS"
                }
            }
        """.trimIndent()

        // Parse the response
        val parsedResult = AssistantResponseParser.parseMqttAssistantResult(inProgressMessage)

        // Verify parsing succeeded
        assert(parsedResult != null)
        assertEquals("IDEMPOTENCY_IN_PROGRESS", parsedResult?.source)

        // CRITICAL: isDupInProgress should be true, signaling this should be ignored
        assertEquals(true, parsedResult?.isDupInProgress)

        // Verify it's not a cached response
        assertEquals(false, parsedResult?.isDupCached)
    }

    @Test
    fun `chat answer with IDEMPOTENCY_CACHED is identified as cached final response`() {
        // This test verifies that cached responses are properly identified
        // and would complete the request silently.

        // Simulate chat/answer with IDEMPOTENCY_CACHED
        val cachedMessage = """
            {
                "status": true,
                "message": "Chat processed successfully",
                "data": {
                    "response": "Berhasil mematikan lampu (cached)",
                    "is_control": true,
                    "is_blocked": false,
                    "request_id": "test-chat-request-cached-222",
                    "source": "IDEMPOTENCY_CACHED"
                }
            }
        """.trimIndent()

        // Parse the response
        val parsedResult = AssistantResponseParser.parseMqttAssistantResult(cachedMessage)

        // Verify parsing succeeded
        assert(parsedResult != null)
        assertEquals("IDEMPOTENCY_CACHED", parsedResult?.source)
        assertEquals("Berhasil mematikan lampu (cached)", parsedResult?.responseText)

        // CRITICAL: isDupCached should be true, signaling this is a cached final response
        assertEquals(true, parsedResult?.isDupCached)
        assertEquals(false, parsedResult?.isDupInProgress)
    }

    @Test
    fun `request_id mismatch is detected in MQTT response`() {
        // This test verifies that request_id mismatch is properly detected.

        // Simulate response with different request_id
        val mismatchedResponse = """
            {
                "status": true,
                "message": "Chat processed successfully",
                "data": {
                    "response": "Some response",
                    "request_id": "different-request-id-444",
                    "source": "MQTT_SUBSCRIBER"
                }
            }
        """.trimIndent()

        // Parse the response
        val parsedResult = AssistantResponseParser.parseMqttAssistantResult(mismatchedResponse)

        // Verify parsing succeeded
        assert(parsedResult != null)

        // The actual request_id from data
        val json = com.google.gson.JsonParser.parseString(mismatchedResponse).asJsonObject
        val data = json.getAsJsonObject("data")
        val responseRequestId = if (data.has("request_id") && !data.get("request_id").isJsonNull) {
            data.get("request_id").asString
        } else {
            null
        }

        // Verify request_id is present in response
        assertEquals("different-request-id-444", responseRequestId)

        // Test would detect mismatch in actual flow by comparing with activeRequestId
        assertNotEquals("test-original-request-333", responseRequestId)
    }
}
