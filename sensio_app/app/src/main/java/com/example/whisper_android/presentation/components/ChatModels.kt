package com.example.whisper_android.presentation.components

enum class MessageRole { USER, ASSISTANT }

data class TranscriptionMessage(
    val text: String,
    val role: MessageRole
)
