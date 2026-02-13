package com.example.whisper_android.presentation.dashboard

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.Groups
import androidx.compose.material.icons.outlined.SmartToy
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
import com.example.whisper_android.data.di.NetworkModule
import com.example.whisper_android.presentation.components.DashboardFeatureCard

@Composable
fun DashboardScreen(
    onNavigateToRegister: () -> Unit,
    onNavigateToUpload: () -> Unit,
    onNavigateToStreaming: () -> Unit,
    onNavigateToEdge: () -> Unit,
    viewModel: DashboardViewModel = androidx.lifecycle.viewmodel.compose.viewModel { 
        DashboardViewModel(NetworkModule.authenticateUseCase) 
    }
) {
    val uiState by viewModel.uiState.collectAsState()
    val context = LocalContext.current
    
    var hasMicPermission by remember {
        mutableStateOf(
            androidx.core.content.ContextCompat.checkSelfPermission(
                context,
                android.Manifest.permission.RECORD_AUDIO
            ) == android.content.pm.PackageManager.PERMISSION_GRANTED
        )
    }

    val launcher = androidx.activity.compose.rememberLauncherForActivityResult(
        androidx.activity.result.contract.ActivityResultContracts.RequestPermission()
    ) { isGranted ->
        hasMicPermission = isGranted
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(
                Brush.linearGradient(
                    colors = listOf(
                        MaterialTheme.colorScheme.primaryContainer,
                        MaterialTheme.colorScheme.primary
                    )
                )
            )
    ) {
        if (uiState.isLoading) {
            CircularProgressIndicator(modifier = Modifier.align(Alignment.Center), color = Color.White)
        } else if (uiState.isAuthenticated) {
            DashboardContent(
                onNavigateToStreaming = onNavigateToStreaming,
                onNavigateToEdge = onNavigateToEdge
            )
        } else {
            // Error handling (keep existing UI for errors)
            Column(
                modifier = Modifier.align(Alignment.Center).padding(24.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(16.dp)
            ) {
                Text(
                    text = uiState.error ?: "Authentication Failed",
                    style = MaterialTheme.typography.bodyLarge,
                    color = Color.White,
                    textAlign = TextAlign.Center
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
    onNavigateToEdge: () -> Unit
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .statusBarsPadding() // Add padding for transparent status bar
            .padding(24.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.SpaceEvenly
    ) {
        // Header Section
        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            modifier = Modifier.padding(top = 16.dp)
        ) {
            Text(
                text = "Select Workspace",
                fontSize = 40.sp,
                fontWeight = FontWeight.Bold,
                color = Color.White,
                textAlign = TextAlign.Center,
                lineHeight = 44.sp,
                style = androidx.compose.ui.text.TextStyle(
                    shadow = androidx.compose.ui.graphics.Shadow(
                        color = Color.Black.copy(alpha = 0.3f),
                        offset = androidx.compose.ui.geometry.Offset(2f, 2f),
                        blurRadius = 6f
                    )
                )
            )
        }

        // Cards Section
        BoxWithConstraints(
            modifier = Modifier
                .fillMaxWidth()
                .padding(vertical = 16.dp), // Reduced from 48.dp
            contentAlignment = Alignment.Center
        ) {
            val isTablet = maxWidth > 600.dp
            
            if (isTablet) {
                Row(
                    modifier = Modifier.fillMaxWidth(0.95f),
                    horizontalArrangement = Arrangement.spacedBy(24.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    DashboardFeatureCard(
                        title = "Meeting Transcriber & Summary",
                        description = "Record, transcribe, and generate summaries of your meetings.",
                        icon = {
                            Icon(
                                imageVector = Icons.Outlined.Groups,
                                contentDescription = null,
                                tint = MaterialTheme.colorScheme.primary,
                                modifier = Modifier.size(72.dp) // Reduced icon
                            )
                        },
                        onClick = onNavigateToStreaming,
                        modifier = Modifier.weight(1f).height(240.dp) // Height reduced
                    )
                    DashboardFeatureCard(
                        title = "AI Assistant",
                        description = "Real-time conversational AI for assistance and tasks.",
                        icon = {
                            Icon(
                                imageVector = Icons.Outlined.SmartToy,
                                contentDescription = null,
                                tint = MaterialTheme.colorScheme.primary,
                                modifier = Modifier.size(72.dp) // Reduced icon
                            )
                        },
                        onClick = onNavigateToEdge,
                        modifier = Modifier.weight(1f).height(240.dp) // Height reduced
                    )
                }
            } else {
                Column(
                    modifier = Modifier.fillMaxWidth(),
                    verticalArrangement = Arrangement.spacedBy(12.dp)
                ) {
                    DashboardFeatureCard(
                        title = "Meeting Transcriber", // Shortened
                        description = "Transcribe and summarize meetings.", // Shortened
                        icon = {
                            Icon(
                                imageVector = Icons.Outlined.Groups,
                                contentDescription = null,
                                tint = MaterialTheme.colorScheme.primary,
                                modifier = Modifier.size(56.dp)
                            )
                        },
                        onClick = onNavigateToStreaming,
                        modifier = Modifier.height(180.dp)
                    )
                    DashboardFeatureCard(
                        title = "AI Assistant",
                        description = "Conversational AI for tasks.", // Shortened
                        icon = {
                            Icon(
                                imageVector = Icons.Outlined.SmartToy,
                                contentDescription = null,
                                tint = MaterialTheme.colorScheme.primary,
                                modifier = Modifier.size(56.dp)
                            )
                        },
                        onClick = onNavigateToEdge,
                        modifier = Modifier.height(180.dp)
                    )
                }
            }
        }

        // Footer
        Text(
            text = "Powered by Senso",
            fontSize = 16.sp,
            color = Color.White.copy(alpha = 0.8f),
            fontWeight = FontWeight.Medium
        )
    }
}
