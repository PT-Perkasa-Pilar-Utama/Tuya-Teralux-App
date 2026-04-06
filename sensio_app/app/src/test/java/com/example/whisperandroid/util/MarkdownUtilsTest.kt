package com.example.whisperandroid.util

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class MarkdownUtilsTest {

    @Test
    fun `normalizeMeetingSummaryMarkdown preserves Markdown tables`() {
        val input = """
            | Column A | Column B |
            | --- | --- |
            | Value 1 | Value 2 |
            | Value 3 | Value 4 |
        """.trimIndent()

        val result = normalizeMeetingSummaryMarkdown(input)

        assertTrue("Should preserve table header row", result.contains("| Column A | Column B |"))
        assertTrue("Should preserve table separator", result.contains("| --- | --- |"))
        assertTrue("Should preserve table data rows", result.contains("| Value 1 | Value 2 |"))
    }

    @Test
    fun `normalizeMeetingSummaryMarkdown preserves headings`() {
        val input = """
            # Main Heading
            ## Sub Heading
            ### Section Heading
            Regular text here
        """.trimIndent()

        val result = normalizeMeetingSummaryMarkdown(input)

        assertTrue("Should preserve H1 heading", result.contains("# Main Heading"))
        assertTrue("Should preserve H2 heading", result.contains("## Sub Heading"))
        assertTrue("Should preserve H3 heading", result.contains("### Section Heading"))
    }

    @Test
    fun `normalizeMeetingSummaryMarkdown preserves numbered lists`() {
        val input = """
            1. First item
            2. Second item
            3. Third item
        """.trimIndent()

        val result = normalizeMeetingSummaryMarkdown(input)

        assertTrue("Should preserve numbered list item 1", result.contains("1. First item"))
        assertTrue("Should preserve numbered list item 2", result.contains("2. Second item"))
        assertTrue("Should preserve numbered list item 3", result.contains("3. Third item"))
    }

    @Test
    fun `normalizeMeetingSummaryMarkdown preserves lines with en-dashes`() {
        val input = """
            This is a sentence – with an en-dash.
            Another line with – en-dash in middle.
            Regular line without dash.
        """.trimIndent()

        val result = normalizeMeetingSummaryMarkdown(input)

        assertTrue("Should preserve line with en-dash", result.contains("–"))
        assertTrue("Should preserve first line", result.contains("This is a sentence"))
        assertTrue("Should preserve second line", result.contains("Another line with"))
    }

    @Test
    fun `normalizeMeetingSummaryMarkdown preserves lines with em-dashes`() {
        val input = """
            This is a sentence — with an em-dash.
            Another line — em-dash in middle.
        """.trimIndent()

        val result = normalizeMeetingSummaryMarkdown(input)

        assertTrue("Should preserve line with em-dash", result.contains("—"))
    }

    @Test
    fun `normalizeMeetingSummaryMarkdown collapses excessive blank lines`() {
        val input = """
            First paragraph
            
            
            Second paragraph with extra blanks
            
            
            
            
            Third paragraph with many blanks
        """.trimIndent()

        val result = normalizeMeetingSummaryMarkdown(input)

        // Should collapse 3+ newlines to 2
        assertTrue("Should not have 3+ consecutive newlines", !result.contains("\n\n\n"))
        assertTrue("Should preserve paragraph separation", result.contains("\n\n"))
    }

    @Test
    fun `normalizeMeetingSummaryMarkdown trims trailing whitespace`() {
        val input = """
            Line with trailing spaces   
            Line with trailing tabs		
            Clean line
        """.trimIndent()

        val result = normalizeMeetingSummaryMarkdown(input)

        // Each line should be trimmed
        val lines = result.lines()
        assertTrue("First line should be trimmed", lines[0].endsWith("spaces"))
        assertTrue("Second line should be trimmed", lines[1].endsWith("tabs"))
        assertTrue("Third line should be unchanged", lines[2] == "Clean line")
    }

    @Test
    fun `normalizeMeetingSummaryMarkdown preserves bullet lists`() {
        val input = """
            - Bullet item one
            * Bullet item two
            + Bullet item three
        """.trimIndent()

        val result = normalizeMeetingSummaryMarkdown(input)

        assertTrue("Should preserve dash bullet", result.contains("- Bullet item one"))
        assertTrue("Should preserve asterisk bullet", result.contains("* Bullet item two"))
        assertTrue("Should preserve plus bullet", result.contains("+ Bullet item three"))
    }

    @Test
    fun `normalizeMeetingSummaryMarkdown handles complex meeting summary`() {
        val input = """
            # Meeting Summary

            ## Key Decisions

            | Decision | Owner | Deadline |
            | --- | --- | --- |
            | Implement feature A | John | 2026-03-20 |
            | Review architecture | Jane | 2026-03-25 |

            ## Action Items

            1. Complete the initial draft
            2. Review with the team – scheduled for Friday
            3. Deploy to staging

            ## Notes

            The team discussed the approach – everyone agreed on the proposal.
        """.trimIndent()

        val result = normalizeMeetingSummaryMarkdown(input)

        // Verify all structural elements are preserved
        assertTrue("Should preserve H1", result.contains("# Meeting Summary"))
        assertTrue("Should preserve H2", result.contains("## Key Decisions"))
        assertTrue("Should preserve table header", result.contains("| Decision | Owner | Deadline |"))
        assertTrue("Should preserve table separator", result.contains("| --- | --- | --- |"))
        assertTrue("Should preserve numbered list", result.contains("1. Complete the initial draft"))
        assertTrue("Should preserve line with en-dash", result.contains("–"))
    }

    @Test
    fun `normalizeMeetingSummaryMarkdown handles empty input`() {
        val input = ""
        val result = normalizeMeetingSummaryMarkdown(input)
        assertEquals("", result)
    }

    @Test
    fun `normalizeMeetingSummaryMarkdown handles whitespace-only input`() {
        val input = "   \n\n\n   "
        val result = normalizeMeetingSummaryMarkdown(input)
        assertEquals("", result)
    }

    // region: parseMarkdownIntoBlocks tests

    @Test
    fun `parseMarkdownIntoBlocks detects valid markdown table`() {
        val input = """
            | Column A | Column B |
            | --- | --- |
            | Value 1 | Value 2 |
            | Value 3 | Value 4 |
        """.trimIndent()

        val blocks = parseMarkdownIntoBlocks(input)

        assertEquals(1, blocks.size)
        assertTrue("Should detect table block", blocks[0] is MarkdownBlock.Table)
        val table = blocks[0] as MarkdownBlock.Table
        assertEquals(2, table.headerCount)
        assertEquals(3, table.rows.size) // header + 2 data rows
    }

    @Test
    fun `parseMarkdownIntoBlocks handles table with multiple columns`() {
        val input = """
            | Name | Age | City | Country |
            | --- | --- | --- | --- |
            | John | 30 | NYC | USA |
        """.trimIndent()

        val blocks = parseMarkdownIntoBlocks(input)

        assertEquals(1, blocks.size)
        val table = blocks[0] as MarkdownBlock.Table
        assertEquals(4, table.rows[0].size) // 4 columns in header
        assertEquals("Name", table.rows[0][0])
        assertEquals("USA", table.rows[1][3])
    }

    @Test
    fun `parseMarkdownIntoBlocks handles mixed content with table`() {
        val input = """
            # Meeting Summary

            Here is the summary.

            | Decision | Owner |
            | --- | --- |
            | Feature A | John |

            ## Next Steps

            1. Review code
            2. Deploy
        """.trimIndent()

        val blocks = parseMarkdownIntoBlocks(input)

        assertEquals(5, blocks.size)
        assertTrue("First block should be heading", blocks[0] is MarkdownBlock.Heading)
        assertTrue("Second block should be paragraph", blocks[1] is MarkdownBlock.Paragraph)
        assertTrue("Third block should be table", blocks[2] is MarkdownBlock.Table)
        assertTrue("Fourth block should be heading", blocks[3] is MarkdownBlock.Heading)
        assertTrue("Fifth block should be list", blocks[4] is MarkdownBlock.ListBlock)
    }

    @Test
    fun `parseMarkdownIntoBlocks ignores pipes in regular text`() {
        val input = "This | is | not | a | table"

        val blocks = parseMarkdownIntoBlocks(input)

        assertEquals(1, blocks.size)
        assertTrue("Should be paragraph, not table", blocks[0] is MarkdownBlock.Paragraph)
    }

    @Test
    fun `parseMarkdownIntoBlocks handles malformed table without separator`() {
        val input = """
            | Column A | Column B |
            | Value 1 | Value 2 |
        """.trimIndent()

        val blocks = parseMarkdownIntoBlocks(input)

        // Without separator row, should not be detected as table
        assertTrue("Should not detect as table", blocks.none { it is MarkdownBlock.Table })
    }

    @Test
    fun `parseMarkdownIntoBlocks handles table with empty cells`() {
        val input = """
            | A | B | C |
            | --- | --- | --- |
            | 1 | | 3 |
            | | 5 | |
        """.trimIndent()

        val blocks = parseMarkdownIntoBlocks(input)

        assertEquals(1, blocks.size)
        val table = blocks[0] as MarkdownBlock.Table
        assertEquals(3, table.rows.size)
        assertEquals("", table.rows[1][1]) // Empty cell
        assertEquals("", table.rows[2][0]) // Empty cell
    }

    @Test
    fun `parseMarkdownIntoBlocks handles unordered list`() {
        val input = """
            - Item one
            - Item two
            - Item three
        """.trimIndent()

        val blocks = parseMarkdownIntoBlocks(input)

        assertEquals(1, blocks.size)
        val list = blocks[0] as MarkdownBlock.ListBlock
        assertFalse("Should be unordered", list.isOrdered)
        assertEquals(3, list.items.size)
    }

    @Test
    fun `parseMarkdownIntoBlocks handles ordered list`() {
        val input = """
            1. First item
            2. Second item
            3. Third item
        """.trimIndent()

        val blocks = parseMarkdownIntoBlocks(input)

        assertEquals(1, blocks.size)
        val list = blocks[0] as MarkdownBlock.ListBlock
        assertTrue("Should be ordered", list.isOrdered)
        assertEquals(3, list.items.size)
    }

    @Test
    fun `parseMarkdownIntoBlocks handles multiple headings`() {
        val input = """
            # H1
            ## H2
            ### H3
        """.trimIndent()

        val blocks = parseMarkdownIntoBlocks(input)

        assertEquals(3, blocks.size)
        assertEquals(1, (blocks[0] as MarkdownBlock.Heading).level)
        assertEquals(2, (blocks[1] as MarkdownBlock.Heading).level)
        assertEquals(3, (blocks[2] as MarkdownBlock.Heading).level)
    }

    @Test
    fun `parseMarkdownIntoBlocks handles empty input`() {
        val input = ""
        val blocks = parseMarkdownIntoBlocks(input)
        assertEquals(0, blocks.size)
    }

    @Test
    fun `parseMarkdownIntoBlocks handles whitespace-only input`() {
        val input = "   \n\n\n   "
        val blocks = parseMarkdownIntoBlocks(input)
        assertEquals(0, blocks.size)
    }

    @Test
    fun `containsMarkdownTable returns true for valid table`() {
        val input = """
            | A | B |
            | --- | --- |
            | 1 | 2 |
        """.trimIndent()

        assertTrue(containsMarkdownTable(input))
    }

    @Test
    fun `containsMarkdownTable returns false for pipes in text`() {
        val input = "This | is | not | a | table"

        assertFalse(containsMarkdownTable(input))
    }

    @Test
    fun `containsMarkdownTable returns false for malformed table`() {
        val input = """
            | Column A | Column B |
            | Value 1 | Value 2 |
        """.trimIndent()

        assertFalse(containsMarkdownTable(input))
    }

    // endregion
}
