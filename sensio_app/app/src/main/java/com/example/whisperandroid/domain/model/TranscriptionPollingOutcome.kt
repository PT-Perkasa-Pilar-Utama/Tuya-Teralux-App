package com.example.whisperandroid.domain.model

/**
 * Represents the outcome of a transcription polling operation.
 *
 * This sealed class provides explicit states for handling different transcription results:
 * - [Pending]: Transcription is still being processed
 * - [Completed]: Transcription completed with usable text
 * - [Rejected]: Transcription completed but rejected by ASR quality gate (user-recoverable)
 * - [Failed]: Transcription failed due to technical error (service issue)
 */
sealed class TranscriptionPollingOutcome {
    /**
     * Transcription is still being processed. Continue polling.
     */
    object Pending : TranscriptionPollingOutcome()

    /**
     * Transcription completed successfully with usable text.
     *
     * @param text The transcribed text (preferably refined_text over raw transcription)
     */
    data class Completed(val text: String) : TranscriptionPollingOutcome()

    /**
     * Transcription completed but rejected by ASR quality gate.
     * This is a user-recoverable state (e.g., silent audio, hallucination detected).
     *
     * @param reason The rejection reason from the ASR quality gate
     * @param audioClass Classification of the audio (e.g., "silent", "near_silent", "active")
     * @param providerSkipped Whether the external provider was skipped due to silence detection
     */
    data class Rejected(
        val reason: String,
        val audioClass: String?,
        val providerSkipped: Boolean?
    ) : TranscriptionPollingOutcome()

    /**
     * Transcription failed due to a technical error.
     * This indicates a service issue, not a user error.
     *
     * @param message Error message describing the failure
     */
    data class Failed(val message: String) : TranscriptionPollingOutcome()
}
