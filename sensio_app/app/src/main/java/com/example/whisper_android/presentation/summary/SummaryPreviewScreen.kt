package com.example.whisper_android.presentation.summary

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.foundation.clickable
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Download
import androidx.compose.material.icons.filled.Email
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.whisper_android.presentation.components.EmailInputDialog
import com.example.whisper_android.presentation.components.FeatureBackground
import com.example.whisper_android.presentation.components.FeatureHeader
import com.example.whisper_android.presentation.components.UiState
import dev.jeziellago.compose.markdowntext.MarkdownText

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SummaryPreviewScreen(
    onNavigateBack: () -> Unit,
    viewModel: SummaryViewModel = androidx.lifecycle.viewmodel.compose.viewModel()
) {
    val context = LocalContext.current
    val summaries by viewModel.summaries.collectAsState()
    val selectedLanguage by viewModel.selectedLanguage.collectAsState()
    val mqttStatus by viewModel.mqttStatus.collectAsState()
    var showEmailDialog by remember { mutableStateOf(false) }

    val currentSummary =
        when (selectedLanguage) {
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
                modifier =
                Modifier
                    .fillMaxSize()
                    .padding(paddingValues)
                    .verticalScroll(rememberScrollState())
                    .padding(horizontal = 8.dp, vertical = 6.dp)
            ) {
                // Language & Download Row
                Row(
                    modifier =
                    Modifier
                        .fillMaxWidth()
                        .padding(bottom = 6.dp),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    // Language Buttons
                    Row(
                        modifier =
                        Modifier
                            .background(
                                Color.LightGray.copy(alpha = 0.2f),
                                RoundedCornerShape(20.dp)
                            )
                            .padding(4.dp),
                        horizontalArrangement = Arrangement.spacedBy(4.dp)
                    ) {
                        Button(
                            onClick = { viewModel.selectLanguage("id") },
                            modifier =
                            Modifier
                                .height(32.dp)
                                .widthIn(min = 50.dp),
                            colors =
                            ButtonDefaults.buttonColors(
                                containerColor = if (selectedLanguage == "id") {
                                    MaterialTheme.colorScheme.primary
                                } else {
                                    Color.White
                                }
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
                            modifier =
                            Modifier
                                .height(32.dp)
                                .widthIn(min = 50.dp),
                            colors =
                            ButtonDefaults.buttonColors(
                                containerColor = if (selectedLanguage == "en") {
                                    MaterialTheme.colorScheme.primary
                                } else {
                                    Color.White
                                }
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

                    Spacer(modifier = Modifier.width(8.dp))

                    MqttStatusBadge(
                        status = mqttStatus,
                        onReconnectClick = { viewModel.reconnectMqtt() }
                    )

                    Spacer(modifier = Modifier.weight(1f))

                    // Email Button
                    Button(
                        onClick = { showEmailDialog = true },
                        modifier = Modifier.height(32.dp),
                        colors =
                        ButtonDefaults.buttonColors(
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
                        Text(
                            "Email",
                            fontSize = 12.sp,
                            color = Color.White,
                            fontWeight = FontWeight.Bold
                        )
                    }

                    Spacer(modifier = Modifier.width(8.dp))

                    // Download Button
                    Button(
                        onClick = { /* TODO: Download PDF */ },
                        modifier = Modifier.height(32.dp),
                        colors =
                        ButtonDefaults.buttonColors(
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
                        Text(
                            "PDF",
                            fontSize = 12.sp,
                            color = Color.White,
                            fontWeight = FontWeight.Bold
                        )
                    }
                }

                // Summary Content
                if (currentSummary.isNotEmpty()) {
                    MarkdownText(
                        markdown =
                        currentSummary
                            .replace(Regex("^-+\\s*$", RegexOption.MULTILINE), "")
                            .replace(Regex("^.*–.*$", RegexOption.MULTILINE), "")
                            .replace("\n\n\n", "\n\n")
                            .replace(Regex("\n{3,}"), "\n\n")
                            .trim(),
                        style =
                        MaterialTheme.typography.bodyLarge.copy(
                            color = Color.DarkGray,
                            fontSize = 13.sp,
                            lineHeight = 16.sp
                        ),
                        modifier = Modifier.fillMaxWidth()
                    )
                } else {
                    Box(
                        modifier =
                        Modifier
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
                onSend = { isMacMode, target, subject ->
                    if (isMacMode) {
                        viewModel.sendEmailByMac(target, subject)
                    } else {
                        viewModel.sendEmail(target, subject)
                    }
                    showEmailDialog = false
                }
            )
        }

        // Observe Email State
        val emailState by viewModel.emailState.collectAsState()
        LaunchedEffect(emailState) {
            when (val state = emailState) {
                is UiState.Success -> {
                    android.widget.Toast
                        .makeText(
                            context,
                            "Email sent successfully",
                            android.widget.Toast.LENGTH_SHORT
                        )
                        .show()
                    viewModel.resetEmailState()
                }

                is UiState.Error -> {
                    android.widget.Toast
                        .makeText(context, state.message, android.widget.Toast.LENGTH_LONG)
                        .show()
                    viewModel.resetEmailState()
                }

                else -> {}
            }
        }
    }
}

@Composable
private fun MqttStatusBadge(
    status: com.example.whisper_android.util.MqttHelper.MqttConnectionStatus,
    onReconnectClick: () -> Unit = {}
) {
    val isError = status == com.example.whisper_android.util.MqttHelper.MqttConnectionStatus.DISCONNECTED ||
        status == com.example.whisper_android.util.MqttHelper.MqttConnectionStatus.FAILED

    val color =
        when (status) {
            com.example.whisper_android.util.MqttHelper.MqttConnectionStatus.CONNECTED -> Color(0xFF4CAF50)
            com.example.whisper_android.util.MqttHelper.MqttConnectionStatus.CONNECTING -> Color(0xFFFFC107)
            com.example.whisper_android.util.MqttHelper.MqttConnectionStatus.DISCONNECTED -> Color(0xFFF44336)
            com.example.whisper_android.util.MqttHelper.MqttConnectionStatus.FAILED -> Color(0xFFD32F2F)
        }

    val text =
        when (status) {
            com.example.whisper_android.util.MqttHelper.MqttConnectionStatus.CONNECTED -> "Online"
            com.example.whisper_android.util.MqttHelper.MqttConnectionStatus.CONNECTING -> "Connecting"
            com.example.whisper_android.util.MqttHelper.MqttConnectionStatus.DISCONNECTED -> "Offline"
            com.example.whisper_android.util.MqttHelper.MqttConnectionStatus.FAILED -> "Error"
        }

    Row(
        verticalAlignment = Alignment.CenterVertically,
        modifier =
        Modifier
            .padding(start = 4.dp)
            .background(color.copy(alpha = 0.1f), RoundedCornerShape(12.dp))
            .then(
                if (isError) {
                    Modifier.clickable { onReconnectClick() }
                } else {
                    Modifier
                }
            )
            .padding(horizontal = 8.dp, vertical = 4.dp)
    ) {
        Box(
            modifier =
            Modifier
                .size(8.dp)
                .background(color, androidx.compose.foundation.shape.CircleShape)
        )
        Spacer(modifier = Modifier.width(6.dp))
        Text(
            text = text,
            fontSize = 11.sp,
            fontWeight = FontWeight.Bold,
            color = color
        )
        if (isError) {
            Spacer(modifier = Modifier.width(4.dp))
            Icon(
                imageVector = Icons.Default.Refresh,
                contentDescription = "Reconnect",
                tint = color,
                modifier = Modifier.size(12.dp)
            )
        }
    }
}

