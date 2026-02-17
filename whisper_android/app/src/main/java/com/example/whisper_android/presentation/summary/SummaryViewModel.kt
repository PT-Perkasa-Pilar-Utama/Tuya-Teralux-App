package com.example.whisper_android.presentation.summary

import androidx.lifecycle.ViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow

data class SummariesData(
    val idSummary: String = "",
    val enSummary: String = ""
)

class SummaryViewModel : ViewModel() {
    private val _summaries = MutableStateFlow(SummariesData())
    val summaries: StateFlow<SummariesData> = _summaries

    private val _selectedLanguage = MutableStateFlow("id")
    val selectedLanguage: StateFlow<String> = _selectedLanguage

    init {
        loadSummaries()
    }

    private fun loadSummaries() {
        // Mock data - in real app this would load from assets/summaries.md
        val idSummary = """
# Ringkasan Pertemuan
**Topik Utama:** Evaluasi Q3 dan Perencanaan Q4.

### Poin Penting:
- Pertumbuhan pasar mencapai 15% di kuartal ini.
- Alokasi anggaran baru sudah disetujui.
- Perlu fokus pada kemitraan strategis bulan depan.

### Action Items:
1. Kirim dokumen anggaran ke tim finance.
2. Jadwalkan meeting dengan partner eksternal.
        """.trimIndent()

        val enSummary = """
# Meeting Summary
**Main Topic:** Q3 Performance Review & Q4 Planning.

### Key Highlights:
- Market share growth reached 15% this quarter.
- New budget allocation has been approved.
- Strategic partnerships need focus next month.

### Action Items:
1. Send budget documents to the finance team.
2. Schedule meeting with external partners.
        """.trimIndent()

        _summaries.value = SummariesData(idSummary = idSummary, enSummary = enSummary)
    }

    fun selectLanguage(lang: String) {
        _selectedLanguage.value = lang
    }
}
