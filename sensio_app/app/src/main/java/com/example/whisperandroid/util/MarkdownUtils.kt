package com.example.whisperandroid.util

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

/**
 * Normalizes markdown text for meeting summary rendering.
 *
 * This function preserves important Markdown structures while cleaning up excessive whitespace:
 * - Preserves Markdown table rows (lines starting with |)
 * - Preserves table separator rows (e.g., | --- | --- |)
 * - Preserves numbered lists and headings
 * - Preserves lines containing en-dashes (–) or hyphens (-)
 * - Only collapses excessive blank lines (3+ newlines to 2)
 * - Trims trailing whitespace from each line
 *
 * @param markdown The raw markdown string to normalize
 * @return Normalized markdown string safe for rendering
 */
fun normalizeMeetingSummaryMarkdown(markdown: String): String =
    markdown
        .lines()
        .joinToString("\n") { line ->
            // Preserve lines that are part of tables, lists, headings, or contain dashes
            when {
                // Preserve table rows (lines starting with |)
                line.trimStart().startsWith("|") -> line.trimEnd()
                // Preserve table separator rows
                line.trimStart().matches(Regex("^\\|?[\\s\\-:|]+\\|?\\s*$")) -> line.trimEnd()
                // Preserve headings (# symbols)
                line.trimStart().startsWith("#") -> line.trimEnd()
                // Preserve numbered lists (1. 2. 3. etc.)
                line.trimStart().matches(Regex("^\\d+\\.\\s+.*")) -> line.trimEnd()
                // Preserve bullet lists (-, *, +)
                line.trimStart().matches(Regex("^[-*+]\\s+.*")) -> line.trimEnd()
                // Preserve lines with en-dashes or em-dashes in content
                line.contains("–") || line.contains("—") -> line.trimEnd()
                // For other lines, just trim trailing whitespace
                else -> line.trimEnd()
            }
        }
        // Collapse excessive blank lines (3+ newlines to 2)
        .replace(Regex("\n{3,}"), "\n\n")
        .trim()
