package com.example.teraluxapp.ui.teralux

import androidx.compose.animation.core.*
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.DeviceHub
import androidx.compose.material3.*
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

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun RegisterTeraluxScreen(
    macAddress: String,
    onRegistrationSuccess: () -> Unit,
    viewModel: RegisterTeraluxViewModel = hiltViewModel()
) {
    var deviceName by remember { mutableStateOf("") }
    var roomId by remember { mutableStateOf("") }
    val uiState by viewModel.uiState.collectAsState()
    
    // Animated background
    val infiniteTransition = rememberInfiniteTransition(label = "background")
    val animatedOffset by infiniteTransition.animateFloat(
        initialValue = 0f,
        targetValue = 1f,
        animationSpec = infiniteRepeatable(
            animation = tween(4000, easing = LinearEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "offset"
    )

    LaunchedEffect(uiState) {
        if (uiState is RegisterUiState.Success) {
            onRegistrationSuccess()
        }
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(
                Brush.verticalGradient(
                    colors = listOf(
                        Color(0xFF1E293B), // Slate
                        Color(0xFF334155), // Light slate
                        Color(0xFF475569)  // Lighter slate
                    )
                )
            )
    ) {
        // Animated floating orbs (Moved slightly to be less intrusive)
        Box(
            modifier = Modifier
                .offset(x = (60 * animatedOffset).dp, y = (100 * animatedOffset).dp)
                .size(160.dp) // Slightly smaller
                .alpha(0.2f) // Lower opacity
                .blur(45.dp)
                .background(
                    Color(0xFF10B981),
                    CircleShape
                )
        )
        
        Box(
            modifier = Modifier
                .align(Alignment.BottomEnd)
                .offset(x = (-40 * animatedOffset).dp, y = (-60 * animatedOffset).dp)
                .size(180.dp) // Slightly smaller
                .alpha(0.15f) // Lower opacity
                .blur(55.dp)
                .background(
                    Color(0xFF06B6D4),
                    CircleShape
                )
        )

        Column(
            modifier = Modifier
                .fillMaxSize()
                .verticalScroll(rememberScrollState())
                .padding(24.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            // Icon with glow
            Surface(
                modifier = Modifier.size(60.dp),
                shape = CircleShape,
                color = Color(0xFF10B981).copy(alpha = 0.2f),
                shadowElevation = 12.dp
            ) {
                Box(
                    contentAlignment = Alignment.Center,
                    modifier = Modifier.background(
                        Brush.radialGradient(
                            colors = listOf(
                                Color(0xFF10B981).copy(alpha = 0.3f),
                                Color.Transparent
                            )
                        )
                    )
                ) {
                    Icon(
                        imageVector = Icons.Default.DeviceHub,
                        contentDescription = null,
                        modifier = Modifier.size(32.dp),
                        tint = Color(0xFF34D399)
                    )
                }
            }
            
            Spacer(modifier = Modifier.height(16.dp))
            
            // Title
            Text(
                text = "Register Device",
                fontSize = 28.sp,
                fontWeight = FontWeight.Bold,
                color = Color.White,
                letterSpacing = 0.5.sp
            )
            
            Text(
                text = "Let's get your device connected",
                fontSize = 13.sp,
                color = Color.White.copy(alpha = 0.7f),
                fontWeight = FontWeight.Light,
                modifier = Modifier.padding(top = 4.dp)
            )

            Spacer(modifier = Modifier.height(24.dp))

            // MAC Address Display
            Surface(
                modifier = Modifier.fillMaxWidth(),
                color = Color.White.copy(alpha = 0.08f),
                shape = RoundedCornerShape(12.dp),
                border = androidx.compose.foundation.BorderStroke(
                    1.dp,
                    Color.White.copy(alpha = 0.12f)
                )
            ) {
                Column(modifier = Modifier.padding(12.dp)) {
                    Text(
                        text = "MAC ADDRESS",
                        fontSize = 9.sp,
                        fontWeight = FontWeight.SemiBold,
                        color = Color(0xFF34D399),
                        letterSpacing = 1.2.sp
                    )
                    Spacer(modifier = Modifier.height(4.dp))
                    Text(
                        text = macAddress,
                        fontSize = 13.sp,
                        fontWeight = FontWeight.Medium,
                        color = Color.White.copy(alpha = 0.9f),
                        letterSpacing = 0.5.sp
                    )
                }
            }

            Spacer(modifier = Modifier.height(12.dp))

            // Room ID Input
            OutlinedTextField(
                value = roomId,
                onValueChange = { roomId = it },
                label = { 
                    Text(
                        "Room ID",
                        color = Color.White.copy(alpha = 0.7f)
                    ) 
                },
                placeholder = {
                    Text(
                        "e.g., Room 101",
                        color = Color.White.copy(alpha = 0.4f)
                    )
                },
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(16.dp),
                colors = OutlinedTextFieldDefaults.colors(
                    focusedBorderColor = Color(0xFF34D399),
                    unfocusedBorderColor = Color.White.copy(alpha = 0.2f),
                    focusedTextColor = Color.White,
                    unfocusedTextColor = Color.White,
                    cursorColor = Color(0xFF34D399),
                    focusedLeadingIconColor = Color(0xFF34D399),
                    unfocusedLeadingIconColor = Color.White.copy(alpha = 0.4f),
                    focusedLabelColor = Color(0xFF34D399),
                    unfocusedLabelColor = Color.White.copy(alpha = 0.7f)
                ),
                singleLine = true
            )

            Spacer(modifier = Modifier.height(12.dp))

            // Device Name Input
            OutlinedTextField(
                value = deviceName,
                onValueChange = { deviceName = it },
                label = { 
                    Text(
                        "Device Name",
                        color = Color.White.copy(alpha = 0.7f)
                    ) 
                },
                placeholder = {
                    Text(
                        "e.g., Living Room Hub",
                        color = Color.White.copy(alpha = 0.4f)
                    )
                },
                modifier = Modifier.fillMaxWidth(),
                shape = RoundedCornerShape(12.dp),
                colors = OutlinedTextFieldDefaults.colors(
                    focusedBorderColor = Color(0xFF34D399),
                    unfocusedBorderColor = Color.White.copy(alpha = 0.2f),
                    focusedTextColor = Color.White,
                    unfocusedTextColor = Color.White,
                    cursorColor = Color(0xFF34D399),
                    focusedLeadingIconColor = Color(0xFF34D399),
                    unfocusedLeadingIconColor = Color.White.copy(alpha = 0.4f),
                    focusedLabelColor = Color(0xFF34D399),
                    unfocusedLabelColor = Color.White.copy(alpha = 0.7f)
                ),
                singleLine = true
            )

            Spacer(modifier = Modifier.height(20.dp))

            // Register Button
            if (uiState is RegisterUiState.Loading) {
                CircularProgressIndicator(
                    color = Color(0xFF34D399),
                    modifier = Modifier.size(48.dp)
                )
            } else {
                Button(
                    onClick = {
                        if (deviceName.isNotBlank() && roomId.isNotBlank()) {
                            viewModel.registerDevice(macAddress, roomId, deviceName)
                        }
                    },
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(52.dp),
                    shape = RoundedCornerShape(12.dp),
                    colors = ButtonDefaults.buttonColors(
                        containerColor = Color(0xFF10B981),
                        disabledContainerColor = Color(0xFF10B981).copy(alpha = 0.3f),
                        contentColor = Color.White,
                        disabledContentColor = Color.White.copy(alpha = 0.5f)
                    ),
                    elevation = ButtonDefaults.buttonElevation(
                        defaultElevation = 0.dp,
                        pressedElevation = 4.dp
                    ),
                    enabled = deviceName.isNotBlank() && roomId.isNotBlank()
                ) {
                    Text(
                        text = "Register Device",
                        fontSize = 16.sp,
                        fontWeight = FontWeight.SemiBold,
                        letterSpacing = 0.5.sp
                    )
                }
            }

            // Error Message
            if (uiState is RegisterUiState.Error) {
                Spacer(modifier = Modifier.height(24.dp))
                Surface(
                    color = Color(0xFFEF4444).copy(alpha = 0.1f),
                    shape = RoundedCornerShape(12.dp),
                    border = androidx.compose.foundation.BorderStroke(
                        1.dp,
                        Color(0xFFEF4444).copy(alpha = 0.3f)
                    )
                ) {
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(12.dp),
                        verticalAlignment = Alignment.CenterVertically // Center error text vertically
                    ) {
                        Text(
                            text = (uiState as RegisterUiState.Error).message,
                            color = Color(0xFFFCA5A5),
                            style = MaterialTheme.typography.bodySmall,
                            modifier = Modifier.weight(1f)
                        )
                    }
                }
            }
        }
    }
}
