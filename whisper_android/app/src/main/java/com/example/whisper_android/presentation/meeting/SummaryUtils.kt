package com.example.whisper_android.presentation.meeting

import android.content.Context

object SummaryUtils {
    fun loadAndFormatSummary(
        context: Context,
        language: String
    ): String {
        return try {
            val content =
                context.assets
                    .open("summaries.md")
                    .bufferedReader()
                    .use { it.readText() }
            val sectionMarker = if (language == "id") "[ID]" else "[EN]"
            val otherMarker = if (language == "id") "[EN]" else "[ID]"

            val startIndex = content.indexOf(sectionMarker)
            if (startIndex == -1) return "Error: Language section not found"

            val contentStart = startIndex + sectionMarker.length
            val endIndex = content.indexOf(otherMarker, contentStart)

            val rawSection =
                if (endIndex == -1) {
                    content.substring(contentStart)
                } else {
                    content.substring(contentStart, endIndex)
                }

            formatToSingleBlock(rawSection.trim())
        } catch (e: Exception) {
            "Error loading summary: ${e.message}"
        }
    }

    private fun formatToSingleBlock(input: String): String =
        input
            .lines()
            .filter { it.isNotBlank() }
            .map { line ->
                line
                    .trim()
                    // Convert headers to bold
                    .replace(Regex("^#+\\s+(.*)$"), "**$1**")
                    // Convert bullets to manual dots
                    .replace(Regex("^[-\\*]\\s+(.*)$"), "• $1")
                    // Convert numbered lists to manual bold numbers
                    .replace(Regex("^(\\d+)\\.\\s+(.*)$"), "**$1.** $2")
                    // Ensure soft line break (double space)
                    .let { "$it  " }
            }.joinToString("\n")
}
