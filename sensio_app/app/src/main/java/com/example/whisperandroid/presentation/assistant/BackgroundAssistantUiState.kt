package com.example.whisperandroid.presentation.assistant

data class BackgroundAssistantUiState(
    val state: State = State.Hidden,
    val recognizedText: String? = null,
    val assistantText: String? = null,
    val errorText: String? = null,
    val startedAtMs: Long = 0L,
    val micLevel: Float = 0f,
    val sessionId: String = ""
) {
    enum class State {
        Hidden,
        Greeting,
        Listening,
        Processing,
        Result,
        Error
    }
}
