package com.example.whisper_android.presentation.dashboard

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.size
import androidx.compose.ui.unit.dp
import androidx.compose.material.icons.filled.Upload
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material.icons.filled.DeviceHub
import com.example.whisper_android.data.di.NetworkModule

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

    LaunchedEffect(uiState.isAuthenticated, uiState.error) {
        if (!uiState.isLoading && !uiState.isAuthenticated) {
            onNavigateToRegister()
        }
    }
    
    Scaffold { paddingValues ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(paddingValues)
                .padding(16.dp),
            contentAlignment = Alignment.Center
        ) {
            if (uiState.isLoading) {
                 CircularProgressIndicator()
            } else if (uiState.isAuthenticated) {
                DashboardContent(
                    onNavigateToUpload = onNavigateToUpload,
                    onNavigateToStreaming = onNavigateToStreaming,
                    onNavigateToEdge = onNavigateToEdge
                )
            } else {
                 Text("Redirecting to Login...")
            }
        }
    }
}

@Composable
fun DashboardContent(
    onNavigateToUpload: () -> Unit,
    onNavigateToStreaming: () -> Unit,
    onNavigateToEdge: () -> Unit
) {
    Column(
        modifier = Modifier.fillMaxWidth(),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = androidx.compose.foundation.layout.Arrangement.spacedBy(16.dp)
    ) {
        Text(
            text = "Whisper Demo Dashboard",
            style = MaterialTheme.typography.headlineMedium,
            color = MaterialTheme.colorScheme.primary,
            modifier = Modifier.padding(bottom = 32.dp)
        )

        DashboardButton(
            text = "Upload Files",
            icon = androidx.compose.material.icons.Icons.Default.Upload,
            onClick = onNavigateToUpload
        )

        DashboardButton(
            text = "Realtime Streaming",
            icon = androidx.compose.material.icons.Icons.Default.PlayArrow,
            onClick = onNavigateToStreaming
        )

        DashboardButton(
            text = "Edge Computing",
            icon = androidx.compose.material.icons.Icons.Default.DeviceHub,
            onClick = onNavigateToEdge
        )
    }
}

@Composable
fun DashboardButton(
    text: String,
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    onClick: () -> Unit
) {
    androidx.compose.material3.ElevatedButton(
        onClick = onClick,
        modifier = Modifier
            .fillMaxWidth()
            .height(80.dp),
        shape = androidx.compose.foundation.shape.RoundedCornerShape(16.dp)
    ) {
        androidx.compose.foundation.layout.Row(
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = androidx.compose.foundation.layout.Arrangement.spacedBy(16.dp)
        ) {
            androidx.compose.material3.Icon(
                imageVector = icon,
                contentDescription = null,
                modifier = Modifier.size(32.dp)
            )
            Text(
                text = text,
                style = MaterialTheme.typography.titleLarge
            )
        }
    }
}
