package com.example.whisper_android.presentation.register

import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.lifecycle.viewmodel.compose.viewModel
import com.example.whisper_android.data.di.NetworkModule
import com.example.whisper_android.presentation.components.ToastObserver

@Composable
fun RegisterScreen(
    onNavigateToDashboard: () -> Unit,
    viewModel: RegisterViewModel = viewModel { 
        RegisterViewModel(
            NetworkModule.registerUseCase,
            NetworkModule.getTeraluxByMacUseCase
        ) 
    }
) {
    val uiState by viewModel.uiState.collectAsState()
    var name by remember { mutableStateOf("") }
    var roomId by remember { mutableStateOf("") }

    // Reusable Toast Observer
    ToastObserver(
        message = uiState.message,
        onToastShown = { viewModel.clearMessage() }
    )

    // Side effect for navigation on success
    LaunchedEffect(uiState.isSuccess) {
        if (uiState.isSuccess) {
            onNavigateToDashboard()
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
            Column(
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(16.dp),
                modifier = Modifier.fillMaxWidth()
            ) {
                Text(
                    text = "Welcome to Teralux",
                    style = MaterialTheme.typography.headlineMedium,
                    color = MaterialTheme.colorScheme.primary
                )

                OutlinedTextField(
                    value = name,
                    onValueChange = { name = it; viewModel.clearError() },
                    label = { Text("Name") },
                    singleLine = true,
                    modifier = Modifier.fillMaxWidth()
                )

                OutlinedTextField(
                    value = roomId,
                    onValueChange = { roomId = it; viewModel.clearError() },
                    label = { Text("Room ID") },
                    singleLine = true,
                    modifier = Modifier.fillMaxWidth()
                )

                if (uiState.isLoading) {
                    CircularProgressIndicator()
                } else {
                    Button(
                        onClick = { viewModel.register(name, roomId) },
                        modifier = Modifier.fillMaxWidth(),
                        enabled = name.isNotBlank() && roomId.isNotBlank()
                    ) {
                        Text("Register")
                    }
                }

                if (uiState.error != null) {
                    Text(
                        text = uiState.error!!,
                        color = MaterialTheme.colorScheme.error,
                        style = MaterialTheme.typography.bodyMedium
                    )
                }
            }
        }
    }
}
