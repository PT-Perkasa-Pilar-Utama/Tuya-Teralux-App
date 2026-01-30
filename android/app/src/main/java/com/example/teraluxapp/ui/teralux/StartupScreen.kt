package com.example.teraluxapp.ui.teralux

import androidx.compose.animation.core.*
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Text
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.blur
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.lifecycle.viewmodel.compose.hiltViewModel

@OptIn(com.google.accompanist.permissions.ExperimentalPermissionsApi::class)
@Composable
fun StartupScreen(
    onDeviceRegistered: () -> Unit,
    onDeviceNotRegistered: (String) -> Unit,
    viewModel: StartupViewModel = hiltViewModel()
) {
    val context = androidx.compose.ui.platform.LocalContext.current
    
    // Define permissions to request based on SDK version
    val permissionsToRequest = remember {
        mutableListOf(android.Manifest.permission.RECORD_AUDIO).apply {
            if (android.os.Build.VERSION.SDK_INT >= android.os.Build.VERSION_CODES.TIRAMISU) {
                add(android.Manifest.permission.POST_NOTIFICATIONS)
            }
        }
    }

    val permissionState = com.google.accompanist.permissions.rememberMultiplePermissionsState(
        permissions = permissionsToRequest
    )
    
    // Check if all permissions are granted
    if (!permissionState.allPermissionsGranted) {
        // Show Permission Request UI
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(
                    Brush.verticalGradient(
                        colors = listOf(
                            Color(0xFF0F172A),
                            Color(0xFF1E293B),
                            Color(0xFF334155)
                        )
                    )
                ),
            contentAlignment = Alignment.Center
        ) {
             Column(
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.Center,
                modifier = Modifier.padding(32.dp)
            ) {
                Text(
                    text = "Permission Required",
                    fontSize = 20.sp,
                    fontWeight = FontWeight.Bold,
                    color = Color.White
                )
                Spacer(modifier = Modifier.height(16.dp))
                Text(
                    text = "Please grant Microphone access to use Voice Control features.",
                    fontSize = 16.sp,
                    color = Color.White.copy(alpha = 0.8f),
                    textAlign = androidx.compose.ui.text.style.TextAlign.Center
                )
                Spacer(modifier = Modifier.height(32.dp))
                androidx.compose.material3.Button(
                    onClick = { permissionState.launchMultiplePermissionRequest() },
                    colors = androidx.compose.material3.ButtonDefaults.buttonColors(
                        containerColor = Color(0xFF3B82F6)
                    )
                ) {
                    Text("Grant Permissions")
                }
            }
        }
        return // Stop here until permissions are granted
    }

    // --- Original Startup Logic (Runs only after permissions granted) ---
    val uiState by viewModel.uiState.collectAsState()
    
    // Animated background
    val infiniteTransition = rememberInfiniteTransition(label = "background")
    val animatedOffset by infiniteTransition.animateFloat(
        initialValue = 0f,
        targetValue = 1f,
        animationSpec = infiniteRepeatable(
            animation = tween(3500, easing = LinearEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "offset"
    )

    LaunchedEffect(uiState) {
        when (val state = uiState) {
            is StartupUiState.DeviceRegistered -> onDeviceRegistered()
            is StartupUiState.DeviceNotRegistered -> onDeviceNotRegistered(state.macAddress)
            else -> { /* Loading or Error - stay on screen */ }
        }
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(
                Brush.verticalGradient(
                    colors = listOf(
                        Color(0xFF0F172A),
                        Color(0xFF1E293B),
                        Color(0xFF334155)
                    )
                )
            )
    ) {
        // Animated floating orbs
        Box(
            modifier = Modifier
                .offset(x = (120 * animatedOffset).dp, y = (80 * animatedOffset).dp)
                .size(220.dp)
                .alpha(0.25f)
                .blur(55.dp)
                .background(
                    Color(0xFF8B5CF6),
                    CircleShape
                )
        )
        
        Box(
            modifier = Modifier
                .align(Alignment.BottomEnd)
                .offset(x = (-80 * animatedOffset).dp, y = (-120 * animatedOffset).dp)
                .size(200.dp)
                .alpha(0.2f)
                .blur(50.dp)
                .background(
                    Color(0xFF3B82F6),
                    CircleShape
                )
        )
        
        Column(
            modifier = Modifier.fillMaxSize(),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            CircularProgressIndicator(
                color = Color(0xFF60A5FA),
                modifier = Modifier.size(56.dp),
                strokeWidth = 4.dp
            )
            
            Spacer(modifier = Modifier.height(32.dp))
            
            Text(
                text = when (uiState) {
                    is StartupUiState.Loading -> "Checking device..."
                    is StartupUiState.Error -> (uiState as StartupUiState.Error).message
                    else -> "Loading..."
                },
                fontSize = 18.sp,
                fontWeight = FontWeight.Medium,
                color = Color.White.copy(alpha = 0.9f),
                letterSpacing = 0.5.sp
            )
            
            Spacer(modifier = Modifier.height(8.dp))
            
            Text(
                text = "Please wait",
                fontSize = 14.sp,
                color = Color.White.copy(alpha = 0.6f),
                fontWeight = FontWeight.Light
            )
        }
    }
}
