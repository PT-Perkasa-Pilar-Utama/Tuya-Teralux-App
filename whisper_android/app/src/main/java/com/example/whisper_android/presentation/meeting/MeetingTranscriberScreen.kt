package com.example.whisper_android.presentation.meeting

import androidx.compose.animation.*
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBackIosNew
import androidx.compose.material.icons.filled.DeleteOutline
import androidx.compose.material.icons.filled.Download
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.core.content.ContextCompat
import android.Manifest
import android.content.pm.PackageManager
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import dev.jeziellago.compose.markdowntext.MarkdownText
import com.example.whisper_android.presentation.components.LanguagePillToggle

@Composable
fun MeetingTranscriberScreen(
    onNavigateBack: () -> Unit
) {
    var isRecording by remember { mutableStateOf(false) }
    var isProcessing by remember { mutableStateOf(false) }
    var summaryLanguage by remember { mutableStateOf("id") }
    var displaySummary by remember { mutableStateOf("") }
    var pdfUrl by remember { mutableStateOf<String?>(null) }
    
    val context = LocalContext.current
    val scope = rememberCoroutineScope()
    
    val hasPermission = ContextCompat.checkSelfPermission(
        context,
        Manifest.permission.RECORD_AUDIO
    ) == PackageManager.PERMISSION_GRANTED

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(
                Brush.verticalGradient(
                    colors = listOf(
                        Color(0xFF67E8F9), // Cyan 300
                        Color(0xFF06B6D4)  // Cyan 500
                    )
                )
            )
            .statusBarsPadding()
    ) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(horizontal = 8.dp, vertical = 2.dp), // Minimal vertical padding
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            // Header
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(bottom = 0.dp), // Zero extra padding
                contentAlignment = Alignment.Center
            ) {
                IconButton(
                    onClick = onNavigateBack,
                    modifier = Modifier.align(Alignment.CenterStart)
                ) {
                    Icon(
                        imageVector = Icons.Default.ArrowBackIosNew,
                        contentDescription = "Back",
                        tint = Color(0xFF0D9488),
                        modifier = Modifier.size(18.dp) // Smaller icon
                    )
                }
                
                Text(
                    text = "Meeting Transcriber & Summary",
                    fontSize = 18.sp,
                    fontWeight = FontWeight.Bold,
                    color = Color(0xFF0F766E),
                    textAlign = TextAlign.Center
                )
            }

            // Main Transcription Card
            ElevatedCard(
                modifier = Modifier
                    .weight(1f)
                    .fillMaxWidth()
                    .padding(horizontal = 4.dp),
                shape = RoundedCornerShape(32.dp),
                colors = CardDefaults.elevatedCardColors(
                    containerColor = Color.White.copy(alpha = 0.95f)
                ),
                elevation = CardDefaults.elevatedCardElevation(defaultElevation = 0.dp)
            ) {
                Column(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(16.dp) // Reduced from 24
                        .verticalScroll(rememberScrollState())
                ) {
                    // Language selection at top right
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.End
                    ) {
                        LanguagePillToggle(
                            selectedLanguage = summaryLanguage,
                            onLanguageSelected = { summaryLanguage = it }
                        )
                    }
                    
                    Spacer(modifier = Modifier.height(8.dp))

                    if (displaySummary.isEmpty() && !isProcessing) {
                        Box(
                            modifier = Modifier.weight(1f).fillMaxWidth(),
                            contentAlignment = Alignment.Center
                        ) {
                            Text(
                                text = "Record audio to generate a summary...",
                                style = MaterialTheme.typography.bodyLarge,
                                color = Color.Gray,
                                textAlign = TextAlign.Center
                            )
                        }
                    } else if (isProcessing) {
                        Box(
                            modifier = Modifier.weight(1f).fillMaxWidth(),
                            contentAlignment = Alignment.Center
                        ) {
                            Column(horizontalAlignment = Alignment.CenterHorizontally) {
                                CircularProgressIndicator(color = Color(0xFFFF9800))
                                Spacer(modifier = Modifier.height(8.dp))
                                Text(
                                    text = "Generating summary...",
                                    style = MaterialTheme.typography.bodyMedium.copy(fontSize = 13.sp),
                                    color = Color(0xFFFF9800)
                                )
                            }
                        }
                    } else {
                        MarkdownText(
                            markdown = displaySummary,
                            style = MaterialTheme.typography.bodyMedium.copy(
                                color = Color.Black,
                                fontSize = 13.sp, // Reduced from 14
                                lineHeight = 18.sp // Reduced from 20
                            ),
                            modifier = Modifier.fillMaxWidth()
                        )
                        
                        if (pdfUrl != null) {
                            Spacer(modifier = Modifier.height(4.dp)) // Tightened
                            Button(
                                onClick = { /* Handle PDF Download */ },
                                modifier = Modifier.fillMaxWidth(),
                                shape = RoundedCornerShape(12.dp),
                                colors = ButtonDefaults.buttonColors(
                                    containerColor = Color(0xFF0D9488),
                                    contentColor = Color.White
                                )
                            ) {
                                Icon(Icons.Default.Download, contentDescription = null)
                                Spacer(modifier = Modifier.width(8.dp))
                                Text("Download Summary (PDF)")
                            }
                        }
                    }
                }
            }

            Spacer(modifier = Modifier.height(4.dp)) // Added small gap

            // Bottom Centered Control Pill
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(bottom = 4.dp), // Shifted to bottom edge
                contentAlignment = Alignment.Center
            ) {
                Surface(
                    modifier = Modifier
                        .height(48.dp) // Ultra-compact height
                        .widthIn(min = 220.dp),
                    shape = RoundedCornerShape(24.dp),
                    color = Color.White.copy(alpha = 0.85f),
                    tonalElevation = 6.dp,
                    shadowElevation = 8.dp
                ) {
                    Row(
                        modifier = Modifier
                            .padding(horizontal = 10.dp),
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(10.dp)
                    ) {
                        // Mic Button (The Primary Focal Point)
                        com.example.whisper_android.presentation.components.MicButton(
                            isRecording = isRecording,
                            hasPermission = hasPermission,
                            isProcessing = isProcessing,
                            onClick = { 
                                if (!isRecording) isRecording = true 
                            },
                            size = 38.dp // Compact size
                        )

                        // Stop Button (Red Square) - Only shown when recording or processing
                        if (isRecording || isProcessing) {
                            IconButton(
                                onClick = { 
                                    if (isRecording) {
                                        isRecording = false
                                        isProcessing = true
                                        
                                        // Simulation of processing and result delivery
                                        scope.launch {
                                            delay(3000) // Thinking...
                                            displaySummary = if (summaryLanguage == "id") {
                                                "# Ringkasan Pertemuan\n" +
                                                "**Topik Utama:** Evaluasi Q3 dan Perencanaan Q4.\n" +
                                                "### Poin Penting:\n" +
                                                "- Pertumbuhan pasar mencapai 15% di kuartal ini.\n" +
                                                "- Alokasi anggaran baru sudah disetujui.\n" +
                                                "- Perlu fokus pada kemitraan strategis bulan depan.\n" +
                                                "### Action Items:\n" +
                                                "1. Kirim dokumen anggaran ke tim finance.\n" +
                                                "2. Jadwalkan meeting dengan partner eksternal."
                                            } else {
                                                "# Meeting Summary\n" +
                                                "**Main Topic:** Q3 Performance Review & Q4 Planning.\n" +
                                                "### Key Highlights:\n" +
                                                "- Market share growth reached 15% this quarter.\n" +
                                                "- New budget allocation has been approved.\n" +
                                                "- Strategic partnerships need focus next month.\n" +
                                                "### Action Items:\n" +
                                                "1. Send budget documents to the finance team.\n" +
                                                "2. Schedule meeting with external partners."
                                            }
                                            pdfUrl = "/api/static/reports/summary.pdf"
                                            isProcessing = false
                                        }
                                    }
                                },
                                modifier = Modifier.size(32.dp)
                            ) {
                                Surface(
                                    modifier = Modifier.size(16.dp),
                                    color = Color(0xFFEF5350), // Red
                                    shape = RoundedCornerShape(2.dp)
                                ) {}
                            }
                        }

                        // Delete Button (Teal Icon)
                        IconButton(
                            onClick = { 
                                displaySummary = ""
                                pdfUrl = null
                                isRecording = false
                                isProcessing = false
                            },
                            modifier = Modifier.size(32.dp)
                        ) {
                            Icon(
                                imageVector = Icons.Default.DeleteOutline,
                                contentDescription = "Clear",
                                tint = Color(0xFF0D9488),
                                modifier = Modifier.size(20.dp)
                            )
                        }

                        // Status Text
                        Text(
                            text = when {
                                isRecording -> "Recording..."
                                isProcessing -> "Thinking..."
                                else -> "Ready"
                            },
                            fontSize = 14.sp,
                            color = Color(0xFF0D9488),
                            fontWeight = FontWeight.Medium,
                            modifier = Modifier.padding(end = 4.dp)
                        )
                    }
                }
            }
        }
    }
}
