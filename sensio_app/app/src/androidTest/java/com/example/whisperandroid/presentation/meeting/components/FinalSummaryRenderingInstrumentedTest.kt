package com.example.whisperandroid.presentation.meeting.components

import androidx.activity.ComponentActivity
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.test.junit4.createAndroidComposeRule
import androidx.compose.ui.test.onAllNodesWithText
import androidx.test.ext.junit.runners.AndroidJUnit4
import com.example.whisperandroid.domain.usecase.MeetingProcessState
import com.example.whisperandroid.util.normalizeMeetingSummaryMarkdown
import dev.jeziellago.compose.markdowntext.MarkdownText
import org.junit.Assert.assertTrue
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
class FinalSummaryRenderingInstrumentedTest {

    @get:Rule
    val composeRule = createAndroidComposeRule<ComponentActivity>()

    @Test
    fun meetingSuccessContent_rendersKnownBrokenSummaryReadably() {
        composeRule.setContent {
            MaterialTheme {
                MeetingSuccessContent(
                    state = MeetingProcessState.Success(
                        summary = knownBrokenRealSummary,
                        pdfUrl = null
                    )
                )
            }
        }

        assertVisibleText("Meeting Summary")
        assertVisibleText("Ringkasan Pertemuan Tim")
        assertVisibleText("Poin Penting")
        assertVisibleText("Pertumbuhan pasar mencapai 15% di kuartal ini.")
        assertVisibleText("Kirim dokumen anggaran ke tim finance")
        assertVisibleText("Diskusi berjalan dengan lancar.")
        assertVisibleText("Budget allocation")
        assertVisibleText("Approved")
    }

    @Test
    fun meetingSuccessContent_rendersMalformedCommonSummaryReadably() {
        composeRule.setContent {
            MaterialTheme {
                MeetingSuccessContent(
                    state = MeetingProcessState.Success(
                        summary = malformedButCommonSummary,
                        pdfUrl = null
                    )
                )
            }
        }

        assertVisibleText("Meeting Summary")
        assertVisibleText("Attendees")
        assertVisibleText("John Doe")
        assertVisibleText("This is a summary with extra spaces.")
        assertVisibleText("Col A")
        assertVisibleText("Val 2")
        assertVisibleText("First item")
        assertVisibleText("Some content after extra blank lines.")
        assertVisibleText("Final paragraph here.")
    }

    @Test
    fun summaryPreviewParityTarget_rendersSameRepresentativeSections() {
        composeRule.setContent {
            MaterialTheme {
                SummaryPreviewParityContent(summary = knownBrokenRealSummary)
            }
        }

        assertVisibleText("Ringkasan Pertemuan Tim")
        assertVisibleText("Poin Penting")
        assertVisibleText("Pertumbuhan pasar mencapai 15% di kuartal ini.")
        assertVisibleText("Kirim dokumen anggaran ke tim finance")
        assertVisibleText("Diskusi berjalan dengan lancar.")
        assertVisibleText("Budget allocation")
        assertVisibleText("Approved")
    }

    private fun assertVisibleText(text: String) {
        val matches = composeRule.onAllNodesWithText(text, substring = true, useUnmergedTree = true)
            .fetchSemanticsNodes()

        assertTrue("Expected to find text: $text", matches.isNotEmpty())
    }

    @Composable
    private fun SummaryPreviewParityContent(summary: String) {
        MarkdownText(
            markdown = normalizeMeetingSummaryMarkdown(summary),
            style = MaterialTheme.typography.bodyLarge,
            modifier = Modifier.fillMaxWidth()
        )
    }

    private companion object {
        val malformedButCommonSummary = """
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

        val knownBrokenRealSummary = """
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
}
