package com.example.whisper_android.presentation.dashboard

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.whisper_android.data.di.NetworkModule
import com.example.whisper_android.presentation.components.DashboardButton

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

    LaunchedEffect(uiState.isAuthenticated, uiState.error) {
        if (!uiState.isLoading && !uiState.isAuthenticated) {
            onNavigateToRegister()
        }
    }
    
    Scaffold { paddingValues ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(
                    Brush.verticalGradient(
                        colors = listOf(
                            MaterialTheme.colorScheme.surface,
                            MaterialTheme.colorScheme.primaryContainer.copy(alpha = 0.1f)
                        )
                    )
                )
                .padding(paddingValues),
            contentAlignment = Alignment.TopCenter
        ) {
            if (uiState.isLoading) {
                 CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
            } else if (uiState.isAuthenticated) {
                DashboardContent(
                    hasMicPermission = hasMicPermission,
                    onRequestPermission = {
                        launcher.launch(android.Manifest.permission.RECORD_AUDIO)
                    },
                    onNavigateToUpload = onNavigateToUpload,
                    onNavigateToStreaming = onNavigateToStreaming,
                    onNavigateToEdge = onNavigateToEdge
                )
            } else {
                 Text(
                     "Redirecting to Login...",
                     modifier = Modifier.align(Alignment.Center)
                 )
            }
        }
    }
}

@Composable
fun DashboardContent(
    hasMicPermission: Boolean,
    onRequestPermission: () -> Unit,
    onNavigateToUpload: () -> Unit,
    onNavigateToStreaming: () -> Unit,
    onNavigateToEdge: () -> Unit
) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .verticalScroll(rememberScrollState())
            .padding(horizontal = 24.dp, vertical = 40.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(20.dp)
    ) {
        // --- Header Section ---
        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            modifier = Modifier.padding(bottom = 24.dp)
        ) {
            Text(
                text = "Welcome to Whisper Demo",
                style = MaterialTheme.typography.headlineLarge.copy(
                    fontWeight = FontWeight.ExtraBold,
                    letterSpacing = (-0.5).sp
                ),
                color = MaterialTheme.colorScheme.onSurface,
                textAlign = TextAlign.Center
            )
            Text(
                text = "Choose your transcription interface",
                style = MaterialTheme.typography.bodyLarge,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                modifier = Modifier.padding(top = 8.dp)
            )
        }

        // --- Permission Banner ---
        if (!hasMicPermission) {
            Card(
                colors = CardDefaults.cardColors(
                    containerColor = MaterialTheme.colorScheme.errorContainer.copy(alpha = 0.8f)
                ),
                shape = RoundedCornerShape(20.dp),
                modifier = Modifier.fillMaxWidth()
            ) {
                Row(
                    modifier = Modifier.padding(16.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Icon(
                        imageVector = Icons.Default.MicOff,
                        contentDescription = null,
                        tint = MaterialTheme.colorScheme.error,
                        modifier = Modifier.size(32.dp)
                    )
                    Spacer(modifier = Modifier.size(16.dp))
                    Column(modifier = Modifier.weight(1f)) {
                        Text(
                            text = "Microphone Required",
                            style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                            color = MaterialTheme.colorScheme.onErrorContainer
                        )
                        Text(
                            text = "Enable permission to record audio.",
                            style = MaterialTheme.typography.bodySmall,
                            color = MaterialTheme.colorScheme.onErrorContainer
                        )
                    }
                    TextButton(
                        onClick = onRequestPermission,
                        colors = ButtonDefaults.textButtonColors(
                            contentColor = MaterialTheme.colorScheme.error
                        )
                    ) {
                        Text("Grant", fontWeight = FontWeight.Bold)
                    }
                }
            }
        }

        // --- Feature Selection ---
        DashboardButton(
            text = "Upload Files",
            subtitle = "Long-form processing via Cloud",
            icon = Icons.Default.Upload,
            color = MaterialTheme.colorScheme.primary,
            onClick = onNavigateToUpload
        )

        DashboardButton(
            text = "Realtime Streaming",
            subtitle = "Low-latency message streaming",
            icon = Icons.Default.CloudUpload,
            color = MaterialTheme.colorScheme.secondary,
            onClick = onNavigateToStreaming
        )

        DashboardButton(
            text = "Edge Computing",
            subtitle = "Private on-device transcription",
            icon = Icons.Default.Memory,
            color = MaterialTheme.colorScheme.tertiary,
            onClick = onNavigateToEdge
        )
        
        Spacer(modifier = Modifier.height(24.dp))
        
        Text(
            text = "Powered by Senso",
            style = MaterialTheme.typography.labelSmall,
            color = MaterialTheme.colorScheme.outline
        )
    }
}
