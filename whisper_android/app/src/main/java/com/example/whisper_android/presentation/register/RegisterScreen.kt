package com.example.whisper_android.presentation.register

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.clickable
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.text.style.TextAlign
import androidx.lifecycle.viewmodel.compose.viewModel
import com.example.whisper_android.data.di.NetworkModule
import com.example.whisper_android.presentation.components.*
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import android.Manifest
import android.content.pm.PackageManager
import androidx.compose.ui.platform.LocalContext
import androidx.core.content.ContextCompat

@Composable
fun RegisterScreen(
    onNavigateToDashboard: () -> Unit
) {
    val context = LocalContext.current
    val application = context.applicationContext as android.app.Application
    val viewModel: RegisterViewModel = viewModel {
        RegisterViewModel(
            application,
            NetworkModule.registerUseCase,
            NetworkModule.getTeraluxByMacUseCase,
            NetworkModule.authenticateUseCase
        )
    }
    val uiState by viewModel.uiState.collectAsState()
    var name by remember { mutableStateOf("") }
    var roomId by remember { mutableStateOf("") }


    // Permission Launcher
    val permissionLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.RequestMultiplePermissions()
    ) { permissions ->
        val recordGranted = permissions[Manifest.permission.RECORD_AUDIO] ?: false
        val storageGranted = permissions[Manifest.permission.WRITE_EXTERNAL_STORAGE] ?: false
        // We just request them, the UI will update based on checkSelfPermission later
    }

    // Proactive Request on Launch
    LaunchedEffect(Unit) {
        val permissionsToRequest = mutableListOf<String>()
        if (ContextCompat.checkSelfPermission(context, Manifest.permission.RECORD_AUDIO) != PackageManager.PERMISSION_GRANTED) {
            permissionsToRequest.add(Manifest.permission.RECORD_AUDIO)
        }
        if (ContextCompat.checkSelfPermission(context, Manifest.permission.WRITE_EXTERNAL_STORAGE) != PackageManager.PERMISSION_GRANTED) {
            permissionsToRequest.add(Manifest.permission.WRITE_EXTERNAL_STORAGE)
        }
        
        if (permissionsToRequest.isNotEmpty()) {
            permissionLauncher.launch(permissionsToRequest.toTypedArray())
        }
    }

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

    val primaryColor = MaterialTheme.colorScheme.primary
    val containerColor = MaterialTheme.colorScheme.primaryContainer
    val surfaceColor = MaterialTheme.colorScheme.surface

    val bgGradient = Brush.linearGradient(
        colors = listOf(
            MaterialTheme.colorScheme.background,
            MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.8f),
            MaterialTheme.colorScheme.background
        ),
        start = androidx.compose.ui.geometry.Offset(0f, 0f),
        end = androidx.compose.ui.geometry.Offset(2000f, 2000f)
    )

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(bgGradient)
    ) {
        BoxWithConstraints(
            modifier = Modifier.fillMaxSize(),
            contentAlignment = Alignment.Center
        ) {
            val isTablet = maxWidth > 600.dp

            if (isTablet) {
                Row(
                    modifier = Modifier
                        .padding(32.dp)
                        .fillMaxWidth(0.9f),
                    horizontalArrangement = Arrangement.SpaceEvenly,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    // Left Side: Branding & Microcopy
                    Column(
                        modifier = Modifier
                            .weight(1.2f)
                            .padding(end = 48.dp),
                        horizontalAlignment = Alignment.Start,
                        verticalArrangement = Arrangement.Center
                    ) {
                        // Logo is clickable
                        Box(modifier = Modifier.clickable { viewModel.checkRegistration() }) {
                             WhisperLogo()
                        }
                        
                        // Text is NOT clickable
                        Spacer(modifier = Modifier.height(32.dp))
                        Text(
                            text = "Secure Your\nConversation.",
                            fontSize = 48.sp,
                            lineHeight = 56.sp,
                            fontWeight = FontWeight.Black,
                            color = MaterialTheme.colorScheme.onBackground,
                            style = androidx.compose.ui.text.TextStyle(
                                shadow = androidx.compose.ui.graphics.Shadow(
                                    color = Color.Black.copy(alpha = 0.3f),
                                    offset = androidx.compose.ui.geometry.Offset(2f, 2f),
                                    blurRadius = 8f
                                )
                            )
                        )
                        Spacer(modifier = Modifier.height(16.dp))
                        Text(
                            text = "Experience private, high-fidelity transcription for your enterprise meetings.",
                            fontSize = 18.sp,
                            color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.7f),
                            lineHeight = 26.sp,
                            fontWeight = FontWeight.Medium
                        )
                    }

                    // Right Side: Card Form
                    RegisterCard(
                        name = name,
                        onNameChange = { name = it; viewModel.clearError() },
                        roomId = roomId,
                        onRoomIdChange = { roomId = it; viewModel.clearError() },
                        isLoading = uiState.isLoading,
                        error = uiState.error,
                        onRegisterClick = { viewModel.register(name, roomId) },
                        modifier = Modifier.weight(1f)
                    )
                }
            } else {
                // Mobile Layout (Vertical Stack)
                Column(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(top = 16.dp, start = 24.dp, end = 24.dp, bottom = 24.dp)
                        .padding(WindowInsets.statusBars.asPaddingValues()), // Add status bar padding
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.Center
                ) {
                    // Wrapper for Logo (Clickable)
                    Column(
                        horizontalAlignment = Alignment.CenterHorizontally,
                        modifier = Modifier.clickable { viewModel.checkRegistration() }
                    ) {
                        WhisperLogo()
                    }
                    
                    Spacer(modifier = Modifier.height(40.dp))
                    
                    // Text (Not Clickable)
                    Column(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalAlignment = Alignment.CenterHorizontally
                    ) {
                        Text(
                            text = "Secure Your Conversation",
                            fontSize = 24.sp, // Reduced font size to fit
                            fontWeight = FontWeight.Black,
                            color = MaterialTheme.colorScheme.onBackground,
                            textAlign = TextAlign.Center, // Explicitly center text
                            style = androidx.compose.ui.text.TextStyle(
                                shadow = androidx.compose.ui.graphics.Shadow(
                                    color = Color.Black.copy(alpha = 0.2f),
                                    offset = androidx.compose.ui.geometry.Offset(1f, 1f),
                                    blurRadius = 4f
                                )
                            )
                        )
                        Spacer(modifier = Modifier.height(8.dp))
                        Text(
                            text = "Private meeting transcription",
                            fontSize = 16.sp,
                            color = MaterialTheme.colorScheme.onBackground.copy(alpha = 0.6f),
                            fontWeight = FontWeight.Normal,
                            textAlign = TextAlign.Center // Explicitly center text
                        )
                    }
                    Spacer(modifier = Modifier.height(48.dp))
                    RegisterCard(
                        name = name,
                        onNameChange = { name = it; viewModel.clearError() },
                        roomId = roomId,
                        onRoomIdChange = { roomId = it; viewModel.clearError() },
                        isLoading = uiState.isLoading,
                        error = uiState.error,
                        onRegisterClick = { viewModel.register(name, roomId) }
                    )
                }
            }
        }
    }
}

@Composable
fun RegisterCard(
    name: String,
    onNameChange: (String) -> Unit,
    roomId: String,
    onRoomIdChange: (String) -> Unit,
    isLoading: Boolean,
    error: String?,
    onRegisterClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    OutlinedCard(
        modifier = modifier
            .padding(8.dp)
            .wrapContentHeight(),
        shape = RoundedCornerShape(32.dp),
        colors = CardDefaults.outlinedCardColors(
            containerColor = MaterialTheme.colorScheme.surface.copy(alpha = 0.7f)
        ),
        border = androidx.compose.foundation.BorderStroke(
            width = 1.dp,
            color = MaterialTheme.colorScheme.outlineVariant.copy(alpha = 0.3f)
        )
    ) {
        Column(
            modifier = Modifier.padding(32.dp),
            verticalArrangement = Arrangement.spacedBy(20.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            WhisperTextField(
                value = name,
                onValueChange = onNameChange,
                label = "Name"
            )

            WhisperTextField(
                value = roomId,
                onValueChange = onRoomIdChange,
                label = "Room ID",
                isRoomId = true
            )

            if (isLoading) {
                CircularProgressIndicator(color = MaterialTheme.colorScheme.primary)
            } else {
                WhisperButton(
                    text = "Register",
                    onClick = onRegisterClick,
                    enabled = name.isNotBlank() && roomId.isNotBlank()
                )
            }

            if (error != null) {
                Text(
                    text = error,
                    color = Color.Red.copy(alpha = 0.7f),
                    fontSize = 14.sp
                )
            }
        }
    }
}
