package com.example.whisperandroid.presentation.meeting.components

data class MeetingFailureUiModel(
    val title: String,
    val body: String,
    val actionLabel: String? = null,
    val isRetrySuggested: Boolean = true
)

fun mapMeetingErrorToUiModel(rawMessage: String?): MeetingFailureUiModel {
    val lowerMessage = rawMessage?.lowercase() ?: ""

    return when {
        lowerMessage.contains("upload") || lowerMessage.contains("transport") -> {
            MeetingFailureUiModel(
                title = "Upload interrupted",
                body = "Your recording couldn't finish uploading. Check your connection and try again.",
                actionLabel = "Retry",
                isRetrySuggested = true
            )
        }
        lowerMessage.contains("timeout") || lowerMessage.contains("504") || lowerMessage.contains("deadline exceeded") -> {
            MeetingFailureUiModel(
                title = "The summary took too long to process",
                body = "We couldn't finish generating your meeting summary right now. Please try again in a moment.",
                actionLabel = "Retry",
                isRetrySuggested = true
            )
        }
        lowerMessage.contains("initiation failed") || lowerMessage.contains("start") -> {
            MeetingFailureUiModel(
                title = "Couldn't start the summary",
                body = "We couldn't start processing this recording. Please try again.",
                actionLabel = "Retry",
                isRetrySuggested = true
            )
        }
        lowerMessage.contains("unauthorized") || lowerMessage.contains("session expired") || lowerMessage.contains("token") -> {
            MeetingFailureUiModel(
                title = "Session expired",
                body = "Please sign in again, then retry your meeting summary.",
                actionLabel = "Sign In",
                isRetrySuggested = false
            )
        }
        lowerMessage.contains("network") || lowerMessage.contains("connection") -> {
            MeetingFailureUiModel(
                title = "Connection issue",
                body = "We lost the connection while processing your recording. Please try again when your network is stable.",
                actionLabel = "Retry",
                isRetrySuggested = true
            )
        }
        else -> {
            MeetingFailureUiModel(
                title = "Summary unavailable",
                body = "Something went wrong while preparing your meeting summary. Please try again.",
                actionLabel = "Retry",
                isRetrySuggested = true
            )
        }
    }
}
