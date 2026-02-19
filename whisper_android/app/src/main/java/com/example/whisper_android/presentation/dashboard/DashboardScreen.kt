package com.example.whisper_android.presentation.dashboard

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.asPaddingValues
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.systemBars
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.Groups
import androidx.compose.material.icons.outlined.SmartToy
import androidx.compose.material3.Button
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.whisper_android.data.di.NetworkModule
import com.example.whisper_android.presentation.components.DashboardFeatureCard

@Composable
fun DashboardScreen(
    onNavigateToRegister: () -> Unit,
    onNavigateToUpload: () -> Unit,
    onNavigateToStreaming: () -> Unit,
    onNavigateToEdge: () -> Unit,
    viewModel: DashboardViewModel =
        androidx.lifecycle.viewmodel.compose.viewModel {
            DashboardViewModel(NetworkModule.authenticateUseCase)
        },
) {
    val uiState by viewModel.uiState.collectAsState()
    val context = LocalContext.current

    var hasMicPermission by remember {
        mutableStateOf(
            androidx.core.content.ContextCompat.checkSelfPermission(
                context,
                android.Manifest.permission.RECORD_AUDIO,
            ) == android.content.pm.PackageManager.PERMISSION_GRANTED,
        )
    }

    val launcher =
        androidx.activity.compose.rememberLauncherForActivityResult(
            androidx.activity.result.contract.ActivityResultContracts
                .RequestPermission(),
        ) { isGranted ->
            hasMicPermission = isGranted
        }

    Box(
        modifier =
            Modifier
                .fillMaxSize()
                .background(MaterialTheme.colorScheme.background) // Slate950 from theme
                .padding(WindowInsets.systemBars.asPaddingValues()), // Add system bars padding
    ) {
        // Optional: Add a subtle overlay gradient for depth
        Box(
            modifier =
                Modifier
                    .fillMaxSize()
                    .background(
                        Brush.radialGradient(
                            colors =
                                listOf(
                                    MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.15f),
                                    Color.Transparent,
                                ),
                            center =
                                androidx.compose.ui.geometry
                                    .Offset(0f, 0f),
                            radius = 2000f,
                        ),
                    ),
        )
        if (uiState.isLoading) {
            Box(
                modifier = Modifier.fillMaxSize(),
                contentAlignment = Alignment.Center,
            ) {
                CircularProgressIndicator(color = Color.White)
            }
        } else if (uiState.isAuthenticated) {
            DashboardContent(
                onNavigateToStreaming = onNavigateToStreaming,
                onNavigateToEdge = onNavigateToEdge,
            )
        } else {
            // Error handling (keep existing UI for errors)
            Column(
                modifier =
                    Modifier
                        .align(Alignment.Center)
                        .padding(24.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(16.dp),
            ) {
                Text(
                    text = uiState.error ?: "Authentication Failed",
                    style = MaterialTheme.typography.bodyLarge,
                    color = Color.White,
                    textAlign = TextAlign.Center,
                )
                Button(onClick = { viewModel.authenticate() }) {
                    Text("Retry Login")
                }
            }
        }
    }
}

@Composable
fun DashboardContent(
    onNavigateToStreaming: () -> Unit,
    onNavigateToEdge: () -> Unit,
) {
    Column(
        modifier =
            Modifier
                .fillMaxSize()
                .padding(24.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        // Header Section
        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            modifier = Modifier.padding(top = 32.dp),
        ) {
            Text(
                text = "Select Workspace",
                fontSize = 32.sp,
                fontWeight = FontWeight.Black,
                color = MaterialTheme.colorScheme.onBackground,
                textAlign = TextAlign.Center,
                lineHeight = 40.sp,
                letterSpacing = (-0.5).sp,
                style =
                    MaterialTheme.typography.headlineMedium.copy(
                        shadow =
                            androidx.compose.ui.graphics.Shadow(
                                color = Color.Black.copy(alpha = 0.3f),
                                offset =
                                    androidx.compose.ui.geometry
                                        .Offset(2f, 2f),
                                blurRadius = 8f,
                            ),
                    ),
            )
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = "Choose your AI-powered environment",
                fontSize = 16.sp,
                color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.6f),
                fontWeight = FontWeight.Medium,
                textAlign = TextAlign.Center,
            )
        }

        Spacer(modifier = Modifier.weight(1f))

        // Cards Section
        BoxWithConstraints(
            modifier = Modifier.fillMaxWidth(),
            contentAlignment = Alignment.Center,
        ) {
            // Check orientation or width, but mainly we want to fix portrait
            val isTablet = maxWidth > 600.dp

            if (isTablet) {
                Row(
                    modifier = Modifier.fillMaxWidth(0.95f),
                    horizontalArrangement = Arrangement.spacedBy(24.dp),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    DashboardFeatureCard(
                        title = "Meeting Transcriber & Summary",
                        description = "Record, transcribe, and generate summaries of your meetings.",
                        icon = {
                            Icon(
                                imageVector = Icons.Outlined.Groups,
                                contentDescription = null,
                                tint = MaterialTheme.colorScheme.primary,
                                modifier = Modifier.size(72.dp),
                            )
                        },
                        onClick = onNavigateToStreaming,
                        modifier =
                            Modifier
                                .weight(1f)
                                .height(240.dp),
                    )
                    DashboardFeatureCard(
                        title = "AI Assistant",
                        description = "Real-time conversational AI for assistance and tasks.",
                        icon = {
                            Icon(
                                imageVector = Icons.Outlined.SmartToy,
                                contentDescription = null,
                                tint = MaterialTheme.colorScheme.primary,
                                modifier = Modifier.size(72.dp),
                            )
                        },
                        onClick = onNavigateToEdge,
                        modifier =
                            Modifier
                                .weight(1f)
                                .height(240.dp),
                    )
                }
            } else {
                // Phone / Portrait
                Column(
                    modifier = Modifier.fillMaxWidth(),
                    verticalArrangement = Arrangement.spacedBy(16.dp),
                ) {
                    DashboardFeatureCard(
                        title = "Meeting Transcriber",
                        description = "Transcribe and summarize meetings.",
                        icon = {
                            Icon(
                                imageVector = Icons.Outlined.Groups,
                                contentDescription = null,
                                tint = MaterialTheme.colorScheme.primary,
                                modifier = Modifier.size(64.dp), // Increased from 56
                            )
                        },
                        onClick = onNavigateToStreaming,
                        modifier = Modifier.height(200.dp), // Increased from 180
                    )
                    DashboardFeatureCard(
                        title = "AI Assistant",
                        description = "Conversational AI for tasks.",
                        icon = {
                            Icon(
                                imageVector = Icons.Outlined.SmartToy,
                                contentDescription = null,
                                tint = MaterialTheme.colorScheme.primary,
                                modifier = Modifier.size(64.dp), // Increased from 56
                            )
                        },
                        onClick = onNavigateToEdge,
                        modifier = Modifier.height(200.dp), // Increased from 180
                    )
                }
            }
        }

        Spacer(modifier = Modifier.weight(1f))

        // Footer
        Row(
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(8.dp),
            modifier = Modifier.padding(bottom = 16.dp),
        ) {
            Box(
                modifier =
                    Modifier
                        .size(4.dp)
                        .background(MaterialTheme.colorScheme.primary, CircleShape),
            )
            Text(
                text = "Powered by Sensio",
                fontSize = 14.sp,
                color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.4f),
                fontWeight = FontWeight.SemiBold,
                letterSpacing = 1.sp,
            )
        }
    }
}
