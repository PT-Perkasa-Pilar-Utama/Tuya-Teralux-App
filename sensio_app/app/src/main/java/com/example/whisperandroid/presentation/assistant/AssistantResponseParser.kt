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
    val isValidationError: Boolean
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
            isDupDrop = data.source == "HTTP_DUP_DROP",
            isValidationError = false // HTTP usually doesn't return validation error as 200 Success
        )
    }

    private fun parseFromJson(json: JsonObject, raw: String): ParsedAssistantChatResult {
        val data = if (json.has("data") && !json.get("data").isJsonNull) {
            json.getAsJsonObject("data")
        } else {
            null
        }

        val responseText = if (data != null && data.has("response") && !data.get("response").isJsonNull) {
            data.get("response").asString
        } else if (json.has("message") && !json.get("message").isJsonNull) {
            json.get("message").asString
        } else if (raw.contains("Response: \"")) {
            raw.substringAfter("Response: \"").substringBeforeLast("\"")
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

        val isValidationError = json.has("message") && 
            !json.get("message").isJsonNull && 
            json.get("message").asString == "Validation Error"

        return ParsedAssistantChatResult(
            responseText = responseText,
            isBlocked = isBlocked,
            isControl = isControl,
            source = source,
            isDupDrop = source == "MQTT_SYNC_DROP" || source == "MQTT_DUP_DROP",
            isValidationError = isValidationError
        )
    }

    fun getCleanMessage(result: ParsedAssistantChatResult, language: String): String? {
        return when {
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
