package com.example.whisperandroid.presentation.components

enum class MessageRole { USER, ASSISTANT }

data class TranscriptionMessage(
    val text: String,
    val role: MessageRole
)
