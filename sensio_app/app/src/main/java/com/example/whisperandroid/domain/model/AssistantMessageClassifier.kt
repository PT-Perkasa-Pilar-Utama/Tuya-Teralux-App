package com.example.whisperandroid.domain.model

import com.example.whisperandroid.presentation.assistant.ParsedAssistantChatResult

/**
 * Shared message classification logic for assistant transport layers.
 * Used by both foreground (AiAssistantViewModel) and background (BackgroundAssistantCoordinator) flows.
 *
 * This classifier provides a single source of truth for interpreting MQTT message payloads
 * and determining terminal vs. non-terminal states across all assistant entry points.
 *
 * Canonical terminal-state mapping for transcription outcomes:
 * - Completed(text) => continue to chat
 * - Rejected(reason="audio_silent"|"hallucination", audioClass, providerSkipped) =>
 *   show "voice not clear" message (user-recoverable, NOT service issue)
 * - Failed(message) => show service-issue message (technical/backend failure)
 * - Pending => continue waiting/polling
 *
 * Applied uniformly across:
 * - MQTT foreground (AiAssistantViewModel)
 * - MQTT background (BackgroundAssistantCoordinator)
 * - HTTP fallback foreground
 * - HTTP fallback background
 */
object AssistantMessageClassifier {

    /**
     * Classification result categories for assistant messages.
     */
    enum class MessageClassification {
        /** Non-terminal: whisper ACK, in-progress sync - continue waiting */
        ACK_PROGRESS,

        /** Terminal: completed with response text - show to user */
        TERMINAL_SUCCESS,

        /** Terminal: ASR gate rejection (silent audio, hallucination) - user-recoverable */
        TERMINAL_REJECTED,

        /** Terminal: technical/service failure - show service issue */
        TERMINAL_FAILED,

        /** Terminal: guard/skill blocked - show identity fallback */
        TERMINAL_BLOCKED,

        /** Terminal: idempotency cached response - process as final */
        DUPLICATE_CACHED,

        /** Non-terminal: idempotency still processing - continue waiting */
        DUPLICATE_IN_PROGRESS
    }

    /**
     * Result of message classification.
     *
     * @param classification The classification category
     * @param responseText The response text to show (if any)
     * @param rejectionReason Rejection reason from backend (if rejected)
     * @param audioClass Audio classification from backend (if available)
     * @param providerSkipped Whether provider was skipped (if available)
     * @param source Source identifier from backend
     */
    data class ClassificationResult(
        val classification: MessageClassification,
        val responseText: String?,
        val rejectionReason: String?,
        val audioClass: String?,
        val providerSkipped: Boolean?,
        val source: String?
    )

    /**
     * Classifies a parsed assistant message into a canonical category.
     *
     * @param parsedResult The parsed MQTT/HTTP response
     * @param topic The MQTT topic (used for context, e.g., "whisper/answer" vs "chat/answer")
     * @return Classification result with category and metadata
     */
    fun classify(parsedResult: ParsedAssistantChatResult, topic: String): ClassificationResult {
        // Handle idempotency states first (highest priority)
        if (parsedResult.isDupInProgress) {
            return ClassificationResult(
                classification = MessageClassification.DUPLICATE_IN_PROGRESS,
                responseText = null,
                rejectionReason = null,
                audioClass = null,
                providerSkipped = null,
                source = parsedResult.source
            )
        }

        if (parsedResult.isDupCached) {
            return ClassificationResult(
                classification = MessageClassification.DUPLICATE_CACHED,
                responseText = parsedResult.responseText,
                rejectionReason = null,
                audioClass = null,
                providerSkipped = null,
                source = parsedResult.source
            )
        }

        // Handle WHISPER_REJECTED source (backend ASR gate rejection)
        if (parsedResult.source == "WHISPER_REJECTED" && parsedResult.isBlocked) {
            return ClassificationResult(
                classification = MessageClassification.TERMINAL_REJECTED,
                responseText = null,
                rejectionReason = null, // Extract from data if needed in future
                audioClass = null,
                providerSkipped = null,
                source = parsedResult.source
            )
        }

        // Handle guard/skill blocked (identity fallback)
        if (parsedResult.isBlocked) {
            return ClassificationResult(
                classification = MessageClassification.TERMINAL_BLOCKED,
                responseText = null,
                rejectionReason = null,
                audioClass = null,
                providerSkipped = null,
                source = parsedResult.source
            )
        }

        // Handle successful response with text
        if (parsedResult.responseText != null) {
            return ClassificationResult(
                classification = MessageClassification.TERMINAL_SUCCESS,
                responseText = parsedResult.responseText,
                rejectionReason = null,
                audioClass = null,
                providerSkipped = null,
                source = parsedResult.source
            )
        }

        // Handle whisper/answer ACK (non-terminal progress)
        if (topic.endsWith("whisper/answer") && !parsedResult.isBlocked) {
            return ClassificationResult(
                classification = MessageClassification.ACK_PROGRESS,
                responseText = null,
                rejectionReason = null,
                audioClass = null,
                providerSkipped = null,
                source = parsedResult.source
            )
        }

        // Default to failed for empty terminal payloads without other classification
        return ClassificationResult(
            classification = MessageClassification.TERMINAL_FAILED,
            responseText = null,
            rejectionReason = null,
            audioClass = null,
            providerSkipped = null,
            source = parsedResult.source
        )
    }

    /**
     * Returns a user-friendly message for a given classification.
     *
     * @param classification The message classification
     * @param language User's language preference ("en" or other, defaults to Indonesian)
     * @param customMessage Optional custom message from backend
     * @return User-friendly message or null if no message should be shown
     */
    fun getUserMessage(
        classification: MessageClassification,
        language: String,
        customMessage: String? = null
    ): String? {
        return when (classification) {
            MessageClassification.TERMINAL_REJECTED -> {
                // ASR gate rejection (silent audio, hallucination) - user-recoverable
                when (language) {
                    "en" -> "I couldn't hear you clearly. Please try speaking again."
                    else -> "Suara kurang jelas. Coba bicara lagi ya."
                }
            }
            MessageClassification.TERMINAL_FAILED -> {
                // Technical/service failure - service issue message
                customMessage ?: when (language) {
                    "en" -> "Sorry, the AI service or network is having trouble right now. Please try again shortly."
                    else -> "Maaf, koneksi atau layanan AI sedang bermasalah. Coba lagi sebentar ya."
                }
            }
            MessageClassification.TERMINAL_SUCCESS,
            MessageClassification.DUPLICATE_CACHED -> {
                // Use provided custom message (assistant response text)
                customMessage
            }
            MessageClassification.TERMINAL_BLOCKED -> {
                // Guard blocked - typically no message, UI shows identity fallback
                null
            }
            MessageClassification.ACK_PROGRESS,
            MessageClassification.DUPLICATE_IN_PROGRESS -> {
                // Non-terminal states - no message yet
                null
            }
        }
    }

    /**
     * Determines if a classification represents a terminal (final) state.
     * Terminal states should complete the request and allow new input.
     *
     * @param classification The classification to check
     * @return true if terminal, false if non-terminal (still waiting)
     */
    fun isTerminal(classification: MessageClassification): Boolean {
        return when (classification) {
            MessageClassification.TERMINAL_SUCCESS,
            MessageClassification.TERMINAL_REJECTED,
            MessageClassification.TERMINAL_FAILED,
            MessageClassification.TERMINAL_BLOCKED,
            MessageClassification.DUPLICATE_CACHED -> true
            MessageClassification.ACK_PROGRESS,
            MessageClassification.DUPLICATE_IN_PROGRESS -> false
        }
    }
}
