package com.example.whisperandroid.util

/**
 * Represents a parsed block of markdown content.
 */
sealed class MarkdownBlock {
    data class Table(val rows: kotlin.collections.List<kotlin.collections.List<String>>, val headerCount: Int) : MarkdownBlock()
    data class Heading(val text: String, val level: Int) : MarkdownBlock()
    data class ListBlock(val items: kotlin.collections.List<String>, val isOrdered: Boolean) : MarkdownBlock()
    data class Paragraph(val text: String) : MarkdownBlock()
}

/**
 * Parses markdown text into structured blocks for rendering.
 * 
 * This parser identifies:
 * - Markdown tables (header + separator + data rows)
 * - Headings (#, ##, ###, etc.)
 * - Lists (ordered and unordered)
 * - Paragraphs (default)
 * 
 * Parsing is intentionally conservative - only clearly valid table structures
 * are treated as tables to avoid false positives.
 * 
 * @param markdown The markdown text to parse
 * @return List of markdown blocks in order
 */
fun parseMarkdownIntoBlocks(markdown: String): List<MarkdownBlock> {
    val lines = markdown.lines()
    val blocks = mutableListOf<MarkdownBlock>()
    var i = 0
    
    while (i < lines.size) {
        val line = lines[i]
        val trimmedLine = line.trim()
        
        // Skip empty lines
        if (trimmedLine.isEmpty()) {
            i++
            continue
        }
        
        // Check for table
        if (isTableStart(lines, i)) {
            val tableResult = parseTable(lines, i)
            if (tableResult != null) {
                blocks.add(tableResult.second)
                i = tableResult.first
                continue
            }
        }
        
        // Check for heading
        val headingMatch = Regex("^#{1,6}\\s+(.*)$").find(trimmedLine)
        if (headingMatch != null) {
            val level = trimmedLine.takeWhile { it == '#' }.length
            val text = trimmedLine.drop(level).trim()
            blocks.add(MarkdownBlock.Heading(text, level))
            i++
            continue
        }
        
        // Check for ordered list
        val orderedMatch = Regex("^\\d+\\.\\s+(.*)$").find(trimmedLine)
        if (orderedMatch != null) {
            val listItems = mutableListOf<String>()
            var j = i
            while (j < lines.size) {
                val listMatch = Regex("^\\d+\\.\\s+(.*)$").find(lines[j].trim())
                if (listMatch != null) {
                    listItems.add(listMatch.groupValues[1])
                    j++
                } else {
                    break
                }
            }
            blocks.add(MarkdownBlock.ListBlock(listItems, isOrdered = true))
            i = j
            continue
        }
        
        // Check for unordered list
        val unorderedMatch = Regex("^[-*+]\\s+(.*)$").find(trimmedLine)
        if (unorderedMatch != null) {
            val listItems = mutableListOf<String>()
            var j = i
            while (j < lines.size) {
                val listMatch = Regex("^[-*+]\\s+(.*)$").find(lines[j].trim())
                if (listMatch != null) {
                    listItems.add(listMatch.groupValues[1])
                    j++
                } else {
                    break
                }
            }
            blocks.add(MarkdownBlock.ListBlock(listItems, isOrdered = false))
            i = j
            continue
        }
        
        // Default: paragraph
        val paragraphLines = mutableListOf<String>()
        var j = i
        while (j < lines.size) {
            val currentLine = lines[j].trim()
            if (currentLine.isEmpty() || 
                currentLine.startsWith("#") ||
                currentLine.startsWith("|") ||
                Regex("^\\d+\\.\\s+").matches(currentLine) ||
                Regex("^[-*+]\\s+").matches(currentLine)
            ) {
                break
            }
            paragraphLines.add(currentLine)
            j++
        }
        
        if (paragraphLines.isNotEmpty()) {
            blocks.add(MarkdownBlock.Paragraph(paragraphLines.joinToString(" ")))
            i = j
        } else {
            i++
        }
    }
    
    return blocks
}

/**
 * Checks if a table starts at the given line index.
 */
private fun isTableStart(lines: List<String>, index: Int): Boolean {
    if (index >= lines.size) return false
    
    val line = lines[index].trim()
    
    // First line should start with |
    if (!line.startsWith("|")) return false
    
    // Check if next non-empty line is a separator
    var separatorIndex = index + 1
    while (separatorIndex < lines.size && lines[separatorIndex].trim().isEmpty()) {
        separatorIndex++
    }
    
    if (separatorIndex >= lines.size) return false
    
    val separatorLine = lines[separatorIndex].trim()
    return isTableSeparator(separatorLine)
}

/**
 * Checks if a line is a valid table separator (e.g., | --- | --- |).
 */
private fun isTableSeparator(line: String): Boolean {
    val trimmed = line.trim()
    if (!trimmed.startsWith("|") && !trimmed.startsWith("-")) return false
    
    // Split by | and check each cell
    val cells = trimmed.split("|").map { it.trim() }.filter { it.isNotEmpty() }
    if (cells.isEmpty()) return false
    
    // Each cell should be dashes, optionally with : for alignment
    return cells.all { cell ->
        cell.matches(Regex("^:?-+:?$"))
    }
}

/**
 * Parses a table from the given starting line.
 * Returns Pair(endIndex, Table) where endIndex is the next line to process.
 */
private fun parseTable(lines: List<String>, startIndex: Int): Pair<Int, MarkdownBlock.Table>? {
    val rows = mutableListOf<kotlin.collections.List<String>>()
    var i = startIndex
    
    // Parse header row
    val headerLine = lines[i].trim()
    val headerCells = parseTableCells(headerLine)
    if (headerCells.isEmpty()) return null
    rows.add(headerCells)
    i++
    
    // Skip empty lines to find separator
    while (i < lines.size && lines[i].trim().isEmpty()) {
        i++
    }
    
    if (i >= lines.size) return null
    
    // Parse separator row
    val separatorLine = lines[i].trim()
    if (!isTableSeparator(separatorLine)) return null
    i++
    
    // Parse data rows
    while (i < lines.size) {
        val line = lines[i].trim()
        
        // Empty line ends table
        if (line.isEmpty()) break
        
        // Non-table line ends table
        if (!line.startsWith("|")) break
        
        val cells = parseTableCells(line)
        if (cells.isEmpty()) break
        
        rows.add(cells)
        i++
    }
    
    // Need at least header + 1 data row
    if (rows.size < 2) return null

    return Pair(i, MarkdownBlock.Table(rows, headerCount = headerCells.size))
}

/**
 * Parses table cells from a row, handling escaped pipes.
 */
private fun parseTableCells(row: String): kotlin.collections.List<String> {
    val trimmed = row.trim()
    if (!trimmed.startsWith("|")) return emptyList()
    
    // Remove leading and trailing |
    val content = trimmed.removePrefix("|").removeSuffix("|").trim()
    if (content.isEmpty()) return emptyList()
    
    // Split by | and trim each cell
    return content.split("|").map { cell ->
        cell.trim()
    }
}

/**
 * Detects if markdown text contains valid table structures.
 * 
 * @param markdown The markdown text to check
 * @return true if at least one valid table is detected
 */
fun containsMarkdownTable(markdown: String): Boolean {
    val blocks = parseMarkdownIntoBlocks(markdown)
    return blocks.any { it is MarkdownBlock.Table }
}
