package com.example.whisper_android.presentation.summary

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Download
import androidx.compose.material.icons.filled.Email
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
import com.example.whisper_android.presentation.components.*
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
    var showEmailDialog by remember { mutableStateOf(false) }
    
    val currentSummary = when (selectedLanguage) {
        "id" -> summaries.idSummary
        else -> summaries.enSummary
    }

    FeatureBackground {
        Scaffold(
            containerColor = Color.Transparent,
            topBar = {
                FeatureHeader(
                    title = "Meeting Summary Preview",
                    onNavigateBack = onNavigateBack
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

                // Email Button
                Button(
                    onClick = { showEmailDialog = true },
                    modifier = Modifier.height(32.dp),
                    colors = ButtonDefaults.buttonColors(
                        containerColor = MaterialTheme.colorScheme.primary
                    ),
                    shape = RoundedCornerShape(16.dp),
                    contentPadding = PaddingValues(horizontal = 12.dp, vertical = 0.dp)
                ) {
                    Icon(
                        imageVector = Icons.Default.Email,
                        contentDescription = "Email",
                        modifier = Modifier.size(16.dp),
                        tint = Color.White
                    )
                    Spacer(modifier = Modifier.width(4.dp))
                    Text("Email", fontSize = 12.sp, color = Color.White, fontWeight = FontWeight.Bold)
                }

                Spacer(modifier = Modifier.width(8.dp))

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
        
        if (showEmailDialog) {
            EmailInputDialog(
                onDismiss = { showEmailDialog = false },
                onSend = { email, subject ->
                    viewModel.sendEmail(email, subject)
                    showEmailDialog = false
                    // Ideally show loading/success state handling here or observe in main UI
                }
            )
        }
        
        // Observe Email State
        val emailState by viewModel.emailState.collectAsState()
        LaunchedEffect(emailState) {
            when(emailState) {
                 is UiState.Success -> {
                     android.widget.Toast.makeText(context, "Email sent successfully", android.widget.Toast.LENGTH_SHORT).show()
                     viewModel.resetEmailState()
                 }
                 is UiState.Error -> {
                     android.widget.Toast.makeText(context, (emailState as UiState.Error).message, android.widget.Toast.LENGTH_LONG).show()
                     viewModel.resetEmailState()
                 }
                 else -> {}
            }
        }
    }
}

@Composable
fun EmailInputDialog(
    onDismiss: () -> Unit,
    onSend: (String, String) -> Unit
) {
    var email by remember { mutableStateOf("") }
    var subject by remember { mutableStateOf("Meeting Summary") }
    var emailError by remember { mutableStateOf<String?>(null) }
    var subjectError by remember { mutableStateOf<String?>(null) }

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("Send via Email") },
        text = {
            Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                OutlinedTextField(
                    value = email,
                    onValueChange = { 
                        email = it 
                        if (it.isNotBlank()) emailError = null
                    },
                    label = { Text("Recipient Email") },
                    singleLine = true,
                    isError = emailError != null,
                    supportingText = { if (emailError != null) Text(emailError!!) },
                    modifier = Modifier.fillMaxWidth()
                )
                
                OutlinedTextField(
                    value = subject,
                    onValueChange = { 
                        subject = it
                        if (it.isNotBlank()) subjectError = null
                    },
                    label = { Text("Subject") },
                    singleLine = true,
                    isError = subjectError != null,
                    supportingText = { if (subjectError != null) Text(subjectError!!) },
                    modifier = Modifier.fillMaxWidth()
                )
            }
        },
        confirmButton = {
            Button(
                onClick = {
                    var isValid = true
                    if (email.isBlank()) {
                        emailError = "Email is required"
                        isValid = false
                    }
                    if (subject.isBlank()) {
                        subjectError = "Subject is required"
                        isValid = false
                    }
                    
                    if (isValid) {
                        onSend(email, subject)
                    }
                }
            ) {
                Text("Send")
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) {
                Text("Cancel")
            }
        }
    )
}
