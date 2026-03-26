package com.example.whisperandroid.presentation.assistant

import com.example.whisperandroid.data.remote.dto.RAGChatResponseDto
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

class AssistantResponseParserTest {

    @Test
    fun `parseHttpAssistantResult with IDEMPOTENCY_IN_PROGRESS sets isDupInProgress to true`() {
        val response = RAGChatResponseDto(
            response = null,
            isBlocked = null,
            isControl = null,
            source = "IDEMPOTENCY_IN_PROGRESS"
        )

        val result = AssistantResponseParser.parseHttpAssistantResult(response)

        assertTrue(result.isDupInProgress)
        assertFalse(result.isDupCached)
        assertNull(result.responseText)
    }

    @Test
    fun `parseHttpAssistantResult with IDEMPOTENCY_CACHED sets isDupCached to true`() {
        val response = RAGChatResponseDto(
            response = "Cached response",
            isBlocked = false,
            isControl = false,
            source = "IDEMPOTENCY_CACHED"
        )

        val result = AssistantResponseParser.parseHttpAssistantResult(response)

        assertTrue(result.isDupCached)
        assertFalse(result.isDupInProgress)
        assertEquals("Cached response", result.responseText)
    }

    @Test
    fun `parseHttpAssistantResult with normal response sets both duplicate flags to false`() {
        val response = RAGChatResponseDto(
            response = "Normal response",
            isBlocked = false,
            isControl = false,
            source = "HTTP_HANDLER"
        )

        val result = AssistantResponseParser.parseHttpAssistantResult(response)

        assertFalse(result.isDupCached)
        assertFalse(result.isDupInProgress)
        assertEquals("Normal response", result.responseText)
    }

    @Test
    fun `parseHttpAssistantResult with MQTT_SYNC_DROP sets isDupInProgress to true`() {
        val response = RAGChatResponseDto(
            response = null,
            isBlocked = null,
            isControl = null,
            source = "MQTT_SYNC_DROP"
        )

        val result = AssistantResponseParser.parseHttpAssistantResult(response)

        assertTrue(result.isDupInProgress)
        assertFalse(result.isDupCached)
    }

    @Test
    fun `parseHttpAssistantResult with HTTP_DUP_DROP sets isDupCached to true`() {
        val response = RAGChatResponseDto(
            response = "Duplicate drop response",
            isBlocked = false,
            isControl = false,
            source = "HTTP_DUP_DROP"
        )

        val result = AssistantResponseParser.parseHttpAssistantResult(response)

        assertTrue(result.isDupCached)
        assertFalse(result.isDupInProgress)
    }

    @Test
    fun `parseHttpAssistantResult with null source sets both duplicate flags to false`() {
        val response = RAGChatResponseDto(
            response = "Response with null source",
            isBlocked = false,
            isControl = false,
            source = null
        )

        val result = AssistantResponseParser.parseHttpAssistantResult(response)

        assertFalse(result.isDupCached)
        assertFalse(result.isDupInProgress)
        assertEquals("Response with null source", result.responseText)
    }

    @Test
    fun `getCleanMessage returns service issue message for IDEMPOTENCY_IN_PROGRESS with no response`() {
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

        val cleanMessage = AssistantResponseParser.getCleanMessage(result, "en")

        // IDEMPOTENCY_IN_PROGRESS should return null (no message yet, still processing)
        assertNull(cleanMessage)
    }

    @Test
    fun `getCleanMessage returns parsed text for normal response`() {
        val result = ParsedAssistantChatResult(
            responseText = "**Hello**",
            isBlocked = false,
            isControl = false,
            source = "HTTP_HANDLER",
            isDupDrop = false,
            isValidationError = false,
            isDupCached = false,
            isDupInProgress = false
        )

        val cleanMessage = AssistantResponseParser.getCleanMessage(result, "en")

        assertEquals("Hello", cleanMessage)
    }

    @Test
    fun `parseMqttAssistantResult with IDEMPOTENCY_IN_PROGRESS sets isDupInProgress to true`() {
        val message = """
            {
                "data": {
                    "request_id": "test-123",
                    "response": null,
                    "source": "IDEMPOTENCY_IN_PROGRESS"
                }
            }
        """.trimIndent()

        val result = AssistantResponseParser.parseMqttAssistantResult(message)

        assertTrue(result?.isDupInProgress == true)
        assertFalse(result?.isDupCached == true)
        assertNull(result?.responseText)
    }

    @Test
    fun `parseMqttAssistantResult with MQTT_SYNC_DROP sets isDupInProgress to true`() {
        val message = """
            {
                "data": {
                    "request_id": "test-123",
                    "response": null,
                    "source": "MQTT_SYNC_DROP"
                }
            }
        """.trimIndent()

        val result = AssistantResponseParser.parseMqttAssistantResult(message)

        assertTrue(result?.isDupInProgress == true)
        assertFalse(result?.isDupCached == true)
    }

    @Test
    fun `parseMqttAssistantResult with IDEMPOTENCY_CACHED sets isDupCached to true`() {
        val message = """
            {
                "data": {
                    "request_id": "test-123",
                    "response": "Cached response",
                    "source": "IDEMPOTENCY_CACHED"
                }
            }
        """.trimIndent()

        val result = AssistantResponseParser.parseMqttAssistantResult(message)

        assertTrue(result?.isDupCached == true)
        assertFalse(result?.isDupInProgress == true)
        assertEquals("Cached response", result?.responseText)
    }

    @Test
    fun `parseMqttAssistantResult with normal response sets both duplicate flags to false`() {
        val message = """
            {
                "data": {
                    "request_id": "test-123",
                    "response": "Normal response",
                    "source": "MQTT_HANDLER"
                }
            }
        """.trimIndent()

        val result = AssistantResponseParser.parseMqttAssistantResult(message)

        assertFalse(result?.isDupCached == true)
        assertFalse(result?.isDupInProgress == true)
        assertEquals("Normal response", result?.responseText)
    }

    @Test
    fun `parseHttpAssistantResult with MQTT_DUP_DROP sets isDupCached to true for migration`() {
        val response = RAGChatResponseDto(
            response = "Legacy MQTT duplicate drop response",
            isBlocked = false,
            isControl = false,
            source = "MQTT_DUP_DROP"
        )

        val result = AssistantResponseParser.parseHttpAssistantResult(response)

        assertTrue(result.isDupCached)
        assertFalse(result.isDupInProgress)
        assertEquals("Legacy MQTT duplicate drop response", result.responseText)
    }

    @Test
    fun `getCleanMessage returns null for IDEMPOTENCY_IN_PROGRESS (non-final state)`() {
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

        val cleanMessage = AssistantResponseParser.getCleanMessage(result, "en")

        // IDEMPOTENCY_IN_PROGRESS is non-final, should return null (no message yet)
        assertNull(cleanMessage)
    }

    @Test
    fun `getCleanMessage returns payload for IDEMPOTENCY_CACHED (final state)`() {
        val result = ParsedAssistantChatResult(
            responseText = "**Cached response text**",
            isBlocked = false,
            isControl = false,
            source = "IDEMPOTENCY_CACHED",
            isDupDrop = false,
            isValidationError = false,
            isDupCached = true,
            isDupInProgress = false
        )

        val cleanMessage = AssistantResponseParser.getCleanMessage(result, "en")

        // IDEMPOTENCY_CACHED is final, should return cleaned message
        assertEquals("Cached response text", cleanMessage)
    }

    @Test
    fun `parseMqttAssistantResult with MQTT_DUP_DROP sets isDupCached to true for migration`() {
        val message = """
            {
                "data": {
                    "request_id": "test-123",
                    "response": "Legacy MQTT duplicate drop",
                    "source": "MQTT_DUP_DROP"
                }
            }
        """.trimIndent()

        val result = AssistantResponseParser.parseMqttAssistantResult(message)

        assertTrue(result?.isDupCached == true)
        assertFalse(result?.isDupInProgress == true)
        assertEquals("Legacy MQTT duplicate drop", result?.responseText)
    }

    @Test
    fun `parseHttpAssistantResult ensures isDupCached and isDupInProgress are mutually exclusive`() {
        // Test IDEMPOTENCY_CACHED
        val cachedResponse = RAGChatResponseDto(
            response = "Cached",
            isBlocked = false,
            isControl = false,
            source = "IDEMPOTENCY_CACHED"
        )
        val cachedResult = AssistantResponseParser.parseHttpAssistantResult(cachedResponse)
        assertTrue(cachedResult.isDupCached)
        assertFalse(cachedResult.isDupInProgress)

        // Test IDEMPOTENCY_IN_PROGRESS
        val inProgressResponse = RAGChatResponseDto(
            response = null,
            isBlocked = null,
            isControl = null,
            source = "IDEMPOTENCY_IN_PROGRESS"
        )
        val inProgressResult = AssistantResponseParser.parseHttpAssistantResult(inProgressResponse)
        assertFalse(inProgressResult.isDupCached)
        assertTrue(inProgressResult.isDupInProgress)

        // Test normal response
        val normalResponse = RAGChatResponseDto(
            response = "Normal",
            isBlocked = false,
            isControl = false,
            source = "HTTP_HANDLER"
        )
        val normalResult = AssistantResponseParser.parseHttpAssistantResult(normalResponse)
        assertFalse(normalResult.isDupCached)
        assertFalse(normalResult.isDupInProgress)
    }
}
