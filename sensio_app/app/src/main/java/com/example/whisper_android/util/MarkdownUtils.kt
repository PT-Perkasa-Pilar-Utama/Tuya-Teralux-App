package com.example.whisper_android.util

/**
 * Strips basic Markdown symbols for plain text display in chat bubbles.
 */
fun parseMarkdownToText(markdown: String): String =
    markdown
        // Remove bold
        .replace(Regex("\\*\\*(.*?)\\*\\*"), "$1")
        // Remove italic
        .replace(Regex("\\*(.*?)\\*"), "$1")
        // Remove headers
        .replace(Regex("^#+\\s+", RegexOption.MULTILINE), "")
        // Remove links [text](url) -> text
        .replace(Regex("\\[(.*?)\\]\\(.*?\\)"), "$1")
        // Remove inline code
        .replace(Regex("`(.*?)`"), "$1")
        // Remove code blocks
        .replace(Regex("```[a-z]*\\n?(.*?)\\n?```", RegexOption.DOT_MATCHES_ALL), "$1")
        .trim()
