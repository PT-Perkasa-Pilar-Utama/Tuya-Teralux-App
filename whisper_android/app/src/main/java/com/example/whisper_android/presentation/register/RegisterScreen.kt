package com.example.whisper_android.presentation.register

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
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
import androidx.lifecycle.viewmodel.compose.viewModel
import com.example.whisper_android.data.di.NetworkModule
import com.example.whisper_android.presentation.components.*

@Composable
fun RegisterScreen(
    onNavigateToDashboard: () -> Unit,
    viewModel: RegisterViewModel = viewModel { 
        RegisterViewModel(
            NetworkModule.registerUseCase,
            NetworkModule.getTeraluxByMacUseCase,
            NetworkModule.authenticateUseCase
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

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(
                Brush.linearGradient(
                    colors = listOf(
                        Color(0xFFA5F3FC), // Cyan 200
                        Color(0xFFCFFAFE)  // Cyan 100
                    )
                )
            )
    ) {
        // Layout for both Mobile and Tablet (using adaptive approach)
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
                    // Left Side: Logo & Text
                    Column(
                        modifier = Modifier.weight(1f),
                        horizontalAlignment = Alignment.Start
                    ) {
                        WhisperLogo()
                        Spacer(modifier = Modifier.height(24.dp))
                        Text(
                            text = "Welcome to\nWhisper Demo",
                            fontSize = 64.sp,
                            lineHeight = 72.sp,
                            fontWeight = FontWeight.Bold,
                            color = Color.White,
                            style = androidx.compose.ui.text.TextStyle(
                                shadow = androidx.compose.ui.graphics.Shadow(
                                    color = Color.Black.copy(alpha = 0.3f),
                                    offset = androidx.compose.ui.geometry.Offset(2f, 2f),
                                    blurRadius = 4f
                                )
                            )
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
                        .padding(24.dp)
                        .fillMaxSize(),
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.Center
                ) {
                    WhisperLogo()
                    Spacer(modifier = Modifier.height(24.dp))
                    Text(
                        text = "Welcome to Whisper Demo",
                        fontSize = 32.sp,
                        fontWeight = FontWeight.Bold,
                        color = Color.White,
                        modifier = Modifier.align(Alignment.Start),
                        style = androidx.compose.ui.text.TextStyle(
                            shadow = androidx.compose.ui.graphics.Shadow(
                                color = Color.Black.copy(alpha = 0.3f),
                                offset = androidx.compose.ui.geometry.Offset(2f, 2f),
                                blurRadius = 4f
                            )
                        )
                    )
                    Spacer(modifier = Modifier.height(32.dp))
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
    ElevatedCard(
        modifier = modifier
            .padding(16.dp)
            .wrapContentHeight(),
        shape = RoundedCornerShape(24.dp),
        colors = CardDefaults.elevatedCardColors(
            containerColor = Color.White
        ),
        elevation = CardDefaults.elevatedCardElevation(
            defaultElevation = 8.dp
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
                CircularProgressIndicator(color = Color(0xFF06B6D4))
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
