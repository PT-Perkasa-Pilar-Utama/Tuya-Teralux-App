package com.example.teraluxapp.ui.devices

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.Edit
import androidx.compose.material.icons.filled.Power
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.teraluxapp.data.network.Command
import com.example.teraluxapp.data.network.CommandRequest
import com.example.teraluxapp.data.network.RetrofitClient
import kotlinx.coroutines.launch

/**
 * SwitchDeviceScreen for controlling switch/relay devices
 * Uses standard device command API: POST /api/tuya/devices/{id}/commands
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SwitchDeviceScreen(
    deviceId: String,
    deviceName: String,
    token: String,
    onBack: () -> Unit
) {
    val scope = rememberCoroutineScope()
    var isOn by remember { mutableStateOf(false) }
    var isProcessing by remember { mutableStateOf(false) }

    // Send switch command
    val sendSwitchCommand = { switchOn: Boolean ->
        scope.launch {
            isProcessing = true
            try {
                val cmd = Command("switch", switchOn)
                val request = CommandRequest(listOf(cmd))
                val response = RetrofitClient.instance.sendDeviceCommand(token, deviceId, request)
                if (response.isSuccessful) {
                    isOn = switchOn
                }
            } catch (e: Exception) {
                e.printStackTrace()
            } finally {
                isProcessing = false
            }
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(deviceName, fontWeight = FontWeight.Bold) },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "Back")
                    }
                },
                actions = {
                    IconButton(onClick = { /* Edit action */ }) {
                        Icon(Icons.Default.Edit, contentDescription = "Edit")
                    }
                },
                colors = TopAppBarDefaults.topAppBarColors(containerColor = Color.White)
            )
        },
        containerColor = Color.White
    ) { paddingValues ->
        Column(
            modifier = Modifier
                .padding(paddingValues)
                .fillMaxSize()
                .padding(24.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            // Large Power Icon
            Icon(
                imageVector = Icons.Default.Power,
                contentDescription = null,
                modifier = Modifier.size(120.dp),
                tint = if (isOn) Color(0xFF4CAF50) else Color.LightGray
            )

            Spacer(modifier = Modifier.height(32.dp))

            // Status Text
            Text(
                text = if (isOn) "ON" else "OFF",
                fontSize = 48.sp,
                fontWeight = FontWeight.Bold,
                color = if (isOn) Color(0xFF4CAF50) else Color.Gray
            )

            Spacer(modifier = Modifier.height(48.dp))

            // Toggle Button
            Button(
                onClick = { sendSwitchCommand(!isOn) },
                enabled = !isProcessing,
                modifier = Modifier
                    .fillMaxWidth(0.6f)
                    .height(56.dp),
                shape = RoundedCornerShape(28.dp),
                colors = ButtonDefaults.buttonColors(
                    containerColor = if (isOn) Color.Gray else Color(0xFF4CAF50)
                )
            ) {
                if (isProcessing) {
                    CircularProgressIndicator(
                        modifier = Modifier.size(24.dp),
                        color = Color.White
                    )
                } else {
                    Text(
                        text = if (isOn) "Turn Off" else "Turn On",
                        fontSize = 18.sp
                    )
                }
            }
        }
    }
}
