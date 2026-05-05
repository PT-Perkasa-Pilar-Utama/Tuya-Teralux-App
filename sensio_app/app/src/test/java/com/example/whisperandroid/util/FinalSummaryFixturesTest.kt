package com.example.whisperandroid.util

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class FinalSummaryFixturesTest {

    companion object {
        internal val headingHeavySummary = """
# Quarterly Review Meeting

## Attendance
- John Doe (Project Lead)
- Jane Smith (Product Manager)
- Bob Wilson (Engineering)

## Agenda
1. Q3 performance review
2. Q4 planning discussion
3. Budget allocation

### Key Decisions Made
- Approved 15% budget increase for Q4
- Prioritize mobile app development
- Increase team capacity by 2 engineers

#### Next Steps
- Schedule follow-up with stakeholders
- Prepare detailed Q4 roadmap
- Finalize hiring plan by end of month

## Notes
Meeting adjourned at 4:30 PM.
        """.trimIndent()

        internal val listHeavySummary = """
# Sprint Planning Summary

## Action Items
1. Implement user authentication – deadline Friday
2. Update API documentation – deadline Monday
3. Review pull requests — scheduled for Wednesday
4. Deploy to staging environment — needs approval

## Key Highlights
- 23 features completed this sprint
- 8 bugs resolved with — priority fixes
- Team velocity increased 15%

## Decisions Made
- Switch to new database provider — better performance
- Improve CI/CD pipeline — faster deployments
- Add monitoring dashboard — clearer metrics

## Attendees
* John (Engineering Lead)
* Sarah (QA Lead)
* Mike (DevOps)
+ External consultant
        """.trimIndent()

        internal val tableContainingSummary = """
# Project Status Meeting

## Decisions Made

| Decision | Owner | Deadline | Status |
| --- | --- | --- | --- |
| Finalize architecture | John | 2026-03-20 | Approved |
| Review security audit | Jane | 2026-03-25 | Pending |
| Deploy to production | Bob | 2026-04-01 | Scheduled |

## Resource Allocation

| Resource | Current | Required | Gap |
| --- | --- | --- | --- |
| Engineers | 5 | 8 | 3 |
| QA Staff | 2 | 4 | 2 |
| Designers | 1 | 2 | 1 |

## Action Items
1. Schedule interviews — top priority
2. Submit budget request — due Friday
3. Update project timeline — by Monday
        """.trimIndent()

        internal val multilineParagraphSummary = """
# Weekly Team Sync

## Discussion Summary

The team discussed the current project status and identified several areas for improvement.

The main concern raised was about timeline constraints. We need to ensure all deliverables
are completed by the end of Q4. The team agreed to increase communication frequency and
hold more frequent sync meetings to address blockers early.

Another key point was resource allocation. We have — limited capacity but high demand.
The solution proposed was to prioritize based on business value and — focus on critical
paths first.

## Decisions

1. Increase daily standup frequency — from weekly to daily
2. Add weekly report — to track progress
3. Schedule biweekly reviews — for stakeholder alignment

## Next Steps

Follow up with management about the budget request. Also — schedule interviews for the
open positions. The team should — focus on completing the current sprint goals before
moving to new initiatives.
        """.trimIndent()

        internal val malformedButCommonSummary = """
# Meeting Summary

## Attendees
- John Doe
- Jane Smith

## Notes

This is a summary with   extra spaces.

Another line with trailing spaces.

| Col A | Col B |
| --- | --- |
| Val 1 | Val 2 |

1. First item
2. Second item


Some content after extra blank lines.

Final paragraph here.
        """.trimIndent()

        internal val knownBrokenRealSummary = """
# Ringkasan Pertemuan Tim

## Poin Penting
- Pertumbuhan pasar mencapai 15% di kuartal ini.
- Alokasi anggaran baru sudah disetujui.
- Perlu fokus pada kemitraan strategis bulan depan — terutama dengan partner utama.

### Action Items
1. Kirim dokumen anggaran ke tim finance — deadline Jumat
2. Jadwalkan meeting dengan partner eksternal — tanggal belum ditentukan
3. Review proposal dari vendor — perlu persetujuan manajemen

## Catatan Tambahan

Diskusi berjalan dengan lancar. Semua pihak setuju dengan proposal yang diajukan.
Beberapa poin penting yang perlu diperhatikan — masalah teknis yang masih pending
dan sumber daya yang terbatas.

| Item | Status | Priority |
| --- | --- | --- |
| Budget allocation | Approved | High |
| Partnership proposal | Under review | Medium |
| Technical setup | In progress | High |
        """.trimIndent()
    }

    @Test
    fun `Fixture1 - Heading-heavy summary preserves heading levels`() {
        val normalized = normalizeMeetingSummaryMarkdown(headingHeavySummary)

        assertTrue(normalized.contains("# Quarterly Review Meeting"))
        assertTrue(normalized.contains("## Attendance"))
        assertTrue(normalized.contains("### Key Decisions Made"))
        assertTrue(normalized.contains("#### Next Steps"))
    }

    @Test
    fun `Fixture1 - Heading-heavy summary preserves list content`() {
        val normalized = normalizeMeetingSummaryMarkdown(headingHeavySummary)

        assertTrue(normalized.contains("- John Doe"))
        assertTrue(normalized.contains("1. Q3 performance review"))
    }

    @Test
    fun `Fixture2 - List-heavy summary preserves numbered items`() {
        val normalized = normalizeMeetingSummaryMarkdown(listHeavySummary)

        assertTrue(normalized.contains("1. Implement user authentication"))
        assertTrue(normalized.contains("2. Update API documentation"))
        assertTrue(normalized.contains("3. Review pull requests"))
        assertTrue(normalized.contains("4. Deploy to staging environment"))
    }

    @Test
    fun `Fixture2 - List-heavy summary preserves all bullet types and dashes`() {
        val normalized = normalizeMeetingSummaryMarkdown(listHeavySummary)

        assertTrue(normalized.contains("— scheduled for Wednesday"))
        assertTrue(normalized.contains("* John"))
        assertTrue(normalized.contains("+ External consultant"))
    }

    @Test
    fun `Fixture3 - Table-containing summary preserves table structure`() {
        val normalized = normalizeMeetingSummaryMarkdown(tableContainingSummary)

        assertTrue(normalized.contains("| Decision | Owner | Deadline | Status |"))
        assertTrue(normalized.contains("| --- | --- | --- | --- |"))
        assertTrue(normalized.contains("| Finalize architecture | John | 2026-03-20 | Approved |"))
        assertTrue(normalized.contains("| Resource | Current | Required | Gap |"))
    }

    @Test
    fun `Fixture4 - Multiline paragraph summary preserves paragraph breaks`() {
        val normalized = normalizeMeetingSummaryMarkdown(multilineParagraphSummary)

        assertFalse(normalized.contains("\n\n\n"))
        assertTrue(normalized.contains("\n\n"))
        assertTrue(normalized.contains("The main concern raised was about timeline constraints."))
        assertTrue(normalized.contains("Another key point was resource allocation."))
    }

    @Test
    fun `Fixture5 - Malformed summary preserves readability and table`() {
        val normalized = normalizeMeetingSummaryMarkdown(malformedButCommonSummary)

        assertTrue(normalized.contains("# Meeting Summary"))
        assertTrue(normalized.contains("- John Doe"))
        assertTrue(normalized.contains("1. First item"))
        assertTrue(normalized.contains("| Col A | Col B |"))
        assertFalse(normalized.contains("\n\n\n"))
    }

    @Test
    fun `Fixture6 - Known-broken sample preserves boundaries and table structure`() {
        val normalized = normalizeMeetingSummaryMarkdown(knownBrokenRealSummary)

        assertTrue(normalized.contains("- Pertumbuhan pasar"))
        assertTrue(normalized.contains("1. Kirim dokumen"))
        assertTrue(normalized.contains("Diskusi berjalan dengan lancar."))
        assertTrue(normalized.contains("Beberapa poin penting yang perlu diperhatikan"))
        assertTrue(normalized.contains("| Item | Status | Priority |"))
        assertTrue(normalized.contains("— terutama dengan partner utama"))
    }

    @Test
    fun `Parity - Meeting screen and Summary Preview share identical render prep for all fixtures`() {
        val allFixtures = allFixtures()

        allFixtures.forEach { fixture ->
            val meetingPrepared = prepareForMeetingScreen(fixture)
            val previewPrepared = prepareForSummaryPreview(fixture)

            assertEquals(previewPrepared, meetingPrepared)
        }
    }

    @Test
    fun `Parity - Supported markdown fixtures parse into stable block boundaries`() {
        val preparedByFixture = allFixtures().associateWith(::prepareForMeetingScreen)

        for ((fixture, prepared) in preparedByFixture) {
            val blocks = parseMarkdownIntoBlocks(prepared)
            
            assertTrue(blocks.any { it is MarkdownBlock.Heading })
            
            if (fixture.contains("|")) {
                assertTrue(blocks.any { it is MarkdownBlock.Table })
            }
            
            if (fixture.contains("- ") || fixture.contains("1. ") || fixture.contains("* ")) {
                assertTrue(blocks.any { it is MarkdownBlock.ListBlock })
            }
            
            if (fixture.trim().isNotEmpty()) {
                assertTrue(blocks.isNotEmpty())
            }
        }
        
        val headingBlocks = parseMarkdownIntoBlocks(preparedByFixture.getValue(headingHeavySummary))
        assertTrue(headingBlocks.first() is MarkdownBlock.Heading)
        
        val tableBlocks = parseMarkdownIntoBlocks(preparedByFixture.getValue(tableContainingSummary))
        assertTrue(tableBlocks.any { it is MarkdownBlock.Table })
        
        val knownBrokenBlocks = parseMarkdownIntoBlocks(preparedByFixture.getValue(knownBrokenRealSummary))
        assertTrue(knownBrokenBlocks.any { it is MarkdownBlock.Table })
    }

    @Test
    fun `Parity - Known-broken real sample keeps standalone paragraph and table blocks`() {
        val blocks = parseMarkdownIntoBlocks(prepareForMeetingScreen(knownBrokenRealSummary))

        assertTrue(blocks.any { it is MarkdownBlock.Heading })
        assertTrue(blocks.any { it is MarkdownBlock.ListBlock })
        assertTrue(blocks.any { it is MarkdownBlock.Paragraph })
        assertTrue(blocks.any { it is MarkdownBlock.Table })
        assertTrue(blocks.any { it is MarkdownBlock.Paragraph && it.text.contains("Diskusi berjalan") })
    }

    @Test
    fun `Parity - Malformed but common summary remains readable after shared render prep`() {
        val prepared = prepareForMeetingScreen(malformedButCommonSummary)
        val blocks = parseMarkdownIntoBlocks(prepared)

        assertEquals(prepareForSummaryPreview(malformedButCommonSummary), prepared)
        assertTrue(prepared.contains("This is a summary with   extra spaces."))
        assertTrue(prepared.contains("Some content after extra blank lines."))
        assertTrue(blocks.any { it is MarkdownBlock.Paragraph })
        assertTrue(blocks.any { it is MarkdownBlock.Table })
        assertTrue(blocks.any { it is MarkdownBlock.ListBlock })
    }

    private fun prepareForMeetingScreen(summary: String): String =
        normalizeMeetingSummaryMarkdown(summary)

    private fun prepareForSummaryPreview(summary: String): String =
        normalizeMeetingSummaryMarkdown(summary)

    private fun allFixtures(): List<String> = listOf(
        headingHeavySummary,
        listHeavySummary,
        tableContainingSummary,
        multilineParagraphSummary,
        malformedButCommonSummary,
        knownBrokenRealSummary,
    )
}
