package com.example.whisper_android.presentation.summary

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Download
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import dev.jeziellago.compose.markdowntext.MarkdownText

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SummaryPreviewScreen(
    onNavigateBack: () -> Unit,
    viewModel: SummaryViewModel = remember { SummaryViewModel() }
) {
    val context = LocalContext.current
    val summaries by viewModel.summaries.collectAsState()
    val selectedLanguage by viewModel.selectedLanguage.collectAsState()
    
    val currentSummary = when (selectedLanguage) {
        "id" -> summaries.idSummary
        else -> summaries.enSummary
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Meeting Summary Preview", color = Color.White) },
                colors = TopAppBarDefaults.topAppBarColors(
                    containerColor = MaterialTheme.colorScheme.primary
                ),
                navigationIcon = {
                    IconButton(onClick = onNavigateBack) {
                        Icon(
                            imageVector = Icons.AutoMirrored.Filled.ArrowBack,
                            contentDescription = "Back",
                            tint = Color.White
                        )
                    }
                }
            )
        }
    ) { paddingValues ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(paddingValues)
                .verticalScroll(rememberScrollState())
                .padding(horizontal = 8.dp, vertical = 6.dp)
        ) {
            // Language & Download Row
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(bottom = 6.dp),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                // Language Buttons
                Row(
                    modifier = Modifier
                        .background(Color.LightGray.copy(alpha = 0.2f), RoundedCornerShape(20.dp))
                        .padding(4.dp),
                    horizontalArrangement = Arrangement.spacedBy(4.dp)
                ) {
                    Button(
                        onClick = { viewModel.selectLanguage("id") },
                        modifier = Modifier
                            .height(32.dp)
                            .widthIn(min = 50.dp),
                        colors = ButtonDefaults.buttonColors(
                            containerColor = if (selectedLanguage == "id") MaterialTheme.colorScheme.primary else Color.White
                        ),
                        shape = RoundedCornerShape(16.dp),
                        contentPadding = PaddingValues(horizontal = 8.dp, vertical = 0.dp)
                    ) {
                        Text(
                            "ID",
                            fontSize = 12.sp,
                            color = if (selectedLanguage == "id") Color.White else Color.Black,
                            fontWeight = FontWeight.Bold
                        )
                    }

                    Button(
                        onClick = { viewModel.selectLanguage("en") },
                        modifier = Modifier
                            .height(32.dp)
                            .widthIn(min = 50.dp),
                        colors = ButtonDefaults.buttonColors(
                            containerColor = if (selectedLanguage == "en") MaterialTheme.colorScheme.primary else Color.White
                        ),
                        shape = RoundedCornerShape(16.dp),
                        contentPadding = PaddingValues(horizontal = 8.dp, vertical = 0.dp)
                    ) {
                        Text(
                            "EN",
                            fontSize = 12.sp,
                            color = if (selectedLanguage == "en") Color.White else Color.Black,
                            fontWeight = FontWeight.Bold
                        )
                    }
                }

                // Download Button
                Button(
                    onClick = { /* TODO: Download PDF */ },
                    modifier = Modifier.height(32.dp),
                    colors = ButtonDefaults.buttonColors(
                        containerColor = MaterialTheme.colorScheme.primary
                    ),
                    shape = RoundedCornerShape(16.dp),
                    contentPadding = PaddingValues(horizontal = 12.dp, vertical = 0.dp)
                ) {
                    Icon(
                        imageVector = Icons.Default.Download,
                        contentDescription = "Download",
                        modifier = Modifier.size(16.dp),
                        tint = Color.White
                    )
                    Spacer(modifier = Modifier.width(4.dp))
                    Text("PDF", fontSize = 12.sp, color = Color.White, fontWeight = FontWeight.Bold)
                }
            }

            // Summary Content
            if (currentSummary.isNotEmpty()) {
                MarkdownText(
                    markdown = currentSummary
                        .replace(Regex("^-+\\s*$", RegexOption.MULTILINE), "")
                        .replace(Regex("^.*–.*$", RegexOption.MULTILINE), "")
                        .replace("\n\n\n", "\n\n")
                        .replace(Regex("\n{3,}"), "\n\n")
                        .trim(),
                    style = MaterialTheme.typography.bodyLarge.copy(
                        color = Color.DarkGray,
                        fontSize = 13.sp,
                        lineHeight = 16.sp
                    ),
                    modifier = Modifier.fillMaxWidth()
                )
            } else {
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(vertical = 32.dp),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        "No summary available",
                        style = MaterialTheme.typography.bodyLarge,
                        color = Color.Gray,
                        textAlign = TextAlign.Center
                    )
                }
            }
        }
    }
}
