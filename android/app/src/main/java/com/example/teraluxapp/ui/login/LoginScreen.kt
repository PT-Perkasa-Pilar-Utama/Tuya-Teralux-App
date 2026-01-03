package com.example.teraluxapp.ui.login

import androidx.compose.animation.core.*
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Lock
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.blur
import androidx.compose.ui.draw.scale
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.platform.LocalContext
import androidx.hilt.navigation.compose.hiltViewModel
import com.example.teraluxapp.utils.DeviceInfoUtils

@Composable
fun LoginScreen(
    onLoginSuccess: (String, String) -> Unit,
    viewModel: LoginViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    val context = LocalContext.current
    val macAddress = remember { DeviceInfoUtils.getMacAddress(context) }
    
    // Animated background
    val infiniteTransition = rememberInfiniteTransition(label = "background")
    val animatedOffset by infiniteTransition.animateFloat(
        initialValue = 0f,
        targetValue = 1f,
        animationSpec = infiniteRepeatable(
            animation = tween(3000, easing = LinearEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "offset"
    )

    LaunchedEffect(uiState) {
        if (uiState is LoginUiState.Success) {
            val success = uiState as LoginUiState.Success
            onLoginSuccess(success.token, success.uid)
        }
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(
                Brush.verticalGradient(
                    colors = listOf(
                        Color(0xFF0F172A), // Dark slate
                        Color(0xFF1E293B), // Slate
                        Color(0xFF334155)  // Light slate
                    )
                )
            )
    ) {
        // Animated floating orbs
        Box(
            modifier = Modifier
                .offset(x = (100 * animatedOffset).dp, y = (50 * animatedOffset).dp)
                .size(200.dp)
                .alpha(0.3f)
                .blur(50.dp)
                .background(
                    Color(0xFF3B82F6),
                    CircleShape
                )
        )
        
        Box(
            modifier = Modifier
                .align(Alignment.BottomEnd)
                .offset(x = (-50 * animatedOffset).dp, y = (-100 * animatedOffset).dp)
                .size(250.dp)
                .alpha(0.2f)
                .blur(60.dp)
                .background(
                    Color(0xFF8B5CF6),
                    CircleShape
                )
        )

        Column(
            modifier = Modifier
                .fillMaxSize()
                .verticalScroll(rememberScrollState())
                .padding(32.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            // Logo with glow effect
            Surface(
                modifier = Modifier.size(100.dp),
                shape = CircleShape,
                color = Color(0xFF3B82F6).copy(alpha = 0.2f),
                shadowElevation = 20.dp
            ) {
                Box(
                    contentAlignment = Alignment.Center,
                    modifier = Modifier.background(
                        Brush.radialGradient(
                            colors = listOf(
                                Color(0xFF3B82F6).copy(alpha = 0.3f),
                                Color.Transparent
                            )
                        )
                    )
                ) {
                    Icon(
                        imageVector = Icons.Default.Lock,
                        contentDescription = null,
                        modifier = Modifier.size(50.dp),
                        tint = Color(0xFF60A5FA)
                    )
                }
            }
            
            Spacer(modifier = Modifier.height(32.dp))
            
            // App Title
            Text(
                text = "Teralux",
                fontSize = 42.sp,
                fontWeight = FontWeight.Bold,
                color = Color.White,
                letterSpacing = 1.sp
            )
            
            Text(
                text = "Smart Home Assistant",
                fontSize = 16.sp,
                color = Color.White.copy(alpha = 0.7f),
                fontWeight = FontWeight.Light
            )
            
            Spacer(modifier = Modifier.height(16.dp))
            
            // MAC Address chip
            Surface(
                color = Color.White.copy(alpha = 0.1f),
                shape = RoundedCornerShape(20.dp),
                border = androidx.compose.foundation.BorderStroke(
                    1.dp,
                    Color.White.copy(alpha = 0.2f)
                )
            ) {
                Text(
                    text = "Device: $macAddress",
                    style = MaterialTheme.typography.labelSmall,
                    color = Color.White.copy(alpha = 0.8f),
                    modifier = Modifier.padding(horizontal = 16.dp, vertical = 8.dp),
                    fontSize = 11.sp
                )
            }

            Spacer(modifier = Modifier.height(64.dp))

            // Login Button
            if (uiState is LoginUiState.Loading) {
                CircularProgressIndicator(
                    color = Color(0xFF60A5FA),
                    modifier = Modifier.size(48.dp)
                )
            } else {
                Button(
                    onClick = { viewModel.login() },
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(56.dp),
                    shape = RoundedCornerShape(16.dp),
                    colors = ButtonDefaults.buttonColors(
                        containerColor = Color(0xFF3B82F6)
                    ),
                    elevation = ButtonDefaults.buttonElevation(
                        defaultElevation = 8.dp,
                        pressedElevation = 12.dp
                    )
                ) {
                    Text(
                        text = "Sign In with Tuya",
                        fontSize = 18.sp,
                        fontWeight = FontWeight.SemiBold,
                        letterSpacing = 0.5.sp
                    )
                }
            }

            // Error Message
            if (uiState is LoginUiState.Error) {
                Spacer(modifier = Modifier.height(24.dp))
                Surface(
                    color = Color(0xFFEF4444).copy(alpha = 0.2f),
                    shape = RoundedCornerShape(12.dp),
                    border = androidx.compose.foundation.BorderStroke(
                        1.dp,
                        Color(0xFFEF4444).copy(alpha = 0.5f)
                    )
                ) {
                    Text(
                        text = (uiState as LoginUiState.Error).message,
                        color = Color(0xFFFCA5A5),
                        modifier = Modifier.padding(16.dp),
                        style = MaterialTheme.typography.bodySmall
                    )
                }
            }
            
            Spacer(modifier = Modifier.height(32.dp))
            
            // Footer
            Text(
                text = "Berjaya Inovasi Global",
                fontSize = 12.sp,
                color = Color.White.copy(alpha = 0.5f),
                fontWeight = FontWeight.Light
            )
        }
    }
}
