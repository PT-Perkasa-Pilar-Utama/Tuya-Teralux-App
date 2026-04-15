package com.example.whisperandroid.presentation.assistant

import com.example.whisperandroid.data.remote.dto.RAGChatResponseDto
import com.example.whisperandroid.util.parseMarkdownToText
import com.google.gson.JsonObject
import com.google.gson.JsonParser

data class ParsedAssistantChatResult(
    val responseText: String?,
    val isBlocked: Boolean,
    val isControl: Boolean,
    val source: String?,
    val isDupDrop: Boolean,
    val isValidationError: Boolean,
    val isDupCached: Boolean,
    val isDupInProgress: Boolean
)

object AssistantResponseParser {

    fun parseMqttAssistantResult(message: String): ParsedAssistantChatResult? {
        return try {
            val json = JsonParser.parseString(message).asJsonObject
            parseFromJson(json, message)
        } catch (e: Exception) {
            null
        }
    }

    fun parseHttpAssistantResult(data: RAGChatResponseDto): ParsedAssistantChatResult {
        return ParsedAssistantChatResult(
            responseText = data.response,
            isBlocked = data.isBlocked ?: false,
            isControl = data.isControl ?: false,
            source = data.source,
            isDupDrop = isIdempotencyCached(data.source),
            isValidationError = false, // HTTP usually doesn't return validation error as 200 Success
            isDupCached = isIdempotencyCached(data.source),
            isDupInProgress = isIdempotencyInProgress(data.source)
        )
    }

    private fun parseFromJson(json: JsonObject, raw: String): ParsedAssistantChatResult {
        val data = if (json.has("data") && !json.get("data").isJsonNull) {
            json.getAsJsonObject("data")
        } else {
            null
        }

        // CRITICAL: Only use data.response for assistant text.
        // Do NOT fall back to envelope-level "message" field - that is transport status only.
        // This prevents "Chat processed successfully" from appearing as assistant response.
        val responseText = if (data != null && data.has("response") && !data.get("response").isJsonNull) {
            data.get("response").asString
        } else {
            null
        }

        val isBlocked = data?.let {
            it.has("is_blocked") && !it.get("is_blocked").isJsonNull && it.get("is_blocked").asBoolean
        } ?: false

        val isControl = data?.let {
            it.has("is_control") && !it.get("is_control").isJsonNull && it.get("is_control").asBoolean
        } ?: false

        val source = data?.let {
            if (it.has("source") && !it.get("source").isJsonNull) it.get("source").asString else null
        }

        val isValidationError = json.has("message") && !json.get("message").isJsonNull && json.get("message").asString == "Validation Error"

        return ParsedAssistantChatResult(
            responseText = responseText,
            isBlocked = isBlocked,
            isControl = isControl,
            source = source,
            isDupDrop = isIdempotencyCached(source),
            isValidationError = isValidationError,
            isDupCached = isIdempotencyCached(source),
            isDupInProgress = isIdempotencyInProgress(source)
        )
    }

    /**
     * Checks if the source indicates a cached duplicate replay (idempotency completed response).
     * This is a FINAL state - the request can be silently completed.
     *
     * Canonical contract:
     * - "IDEMPOTENCY_CACHED": Duplicate request, returning cached completed response (FINAL)
     *
     * Legacy values (migration window):
     * - "MQTT_DUP_DROP": Legacy MQTT duplicate drop
     * - "HTTP_DUP_DROP": Legacy HTTP duplicate drop
     */
    private fun isIdempotencyCached(source: String?): Boolean {
        return source == "IDEMPOTENCY_CACHED" ||
            source == "MQTT_DUP_DROP" ||
            source == "HTTP_DUP_DROP"
    }

    /**
     * Checks if the source indicates a request is still in progress (idempotency non-final state).
     * This is a NON-FINAL state - the request should continue waiting, NOT complete.
     *
     * Canonical contract:
     * - "IDEMPOTENCY_IN_PROGRESS": Duplicate request, first request still processing (NON-FINAL)
     *
     * Legacy values (migration window):
     * - "MQTT_SYNC_DROP": Text query dropped because Whisper is active (treated as in-progress during migration)
     */
    private fun isIdempotencyInProgress(source: String?): Boolean {
        return source == "IDEMPOTENCY_IN_PROGRESS" ||
            source == "MQTT_SYNC_DROP"
    }

    /**
     * @deprecated Use isIdempotencyCached() or isIdempotencyInProgress() instead.
     * Kept for backward compatibility during migration.
     */
    @Deprecated("Use isIdempotencyCached() or isIdempotencyInProgress() instead", ReplaceWith("isIdempotencyCached(source)"))
    private fun isIdempotencyDuplicate(source: String?): Boolean {
        return isIdempotencyCached(source)
    }

    fun getCleanMessage(result: ParsedAssistantChatResult, language: String): String? {
        return when {
            // ASR Quality Gate: Empty transcript from silent/invalid audio
            // When response is empty/null AND transcript was blocked/invalid, show voice fallback
            result.isBlocked && (result.responseText == null || result.responseText.isBlank()) -> {
                if (language == "en") {
                    "Sorry, voice was not clear. Please try again."
                } else {
                    "Maaf, suara tidak terdengar jelas. Silakan coba lagi."
                }
            }
            result.isValidationError -> if (language == "en") {
                "Sorry, the voice was not clear. Please try again."
            } else {
                "Maaf, suara tidak terdengar dengan jelas. Silakan coba lagi."
            }
            result.responseText != null && result.responseText.isNotBlank() -> {
                val fullText = parseMarkdownToText(result.responseText).trim().removeSurrounding("\"")
                // For control responses with multiple lines, show only the result line (second line)
                // Example: "Mematikan AC rumah.\n\nBerhasil mematikan Sharp AC Rumah." -> show only "Berhasil..."
                val lines = fullText.split("\n\n").filter { it.isNotBlank() }
                if (lines.size >= 2 && lines[0].contains("ACTION:")) {
                    // First line has ACTION:, use the second line (result)
                    lines[1].trim()
                } else if (lines.size >= 2) {
                    // Multiple lines without ACTION:, use the last line (usually the result)
                    lines.last().trim()
                } else {
                    fullText
                }
            }
            result.isBlocked -> if (language == "en") {
                "Hi! I'm Sensio, your smart home assistant. How can I help you today?"
            } else {
                "Halo! Saya Sensio, asisten rumah pintar Anda. Ada yang bisa saya bantu?"
            }
            else -> null
        }
    }
}
