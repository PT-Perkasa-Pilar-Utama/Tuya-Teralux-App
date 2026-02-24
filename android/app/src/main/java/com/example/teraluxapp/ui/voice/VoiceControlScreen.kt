package com.example.teraluxapp.ui.voice

import android.Manifest
import android.util.Log
import androidx.compose.foundation.background
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Mic
import androidx.compose.material.icons.filled.MicOff
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.navigation.NavController
import com.example.teraluxapp.utils.AudioRecorderHelper
import com.example.teraluxapp.utils.MqttHelper
import com.google.accompanist.permissions.ExperimentalPermissionsApi
import com.google.accompanist.permissions.isGranted
import com.google.accompanist.permissions.rememberPermissionState
import kotlinx.coroutines.launch
import java.io.File

@OptIn(ExperimentalPermissionsApi::class)
@Composable
fun VoiceControlScreen(navController: NavController) {
    val context = LocalContext.current
    val coroutineScope = rememberCoroutineScope()
    
    // State
    var isRecording by remember { mutableStateOf(false) }
    var connectionStatus by remember { mutableStateOf("Ready") }
    
    // Helpers
    // NB: In a real app, these should be provided by DI (Hilt)
    val mqttHelper = remember { MqttHelper(context) }
    val audioRecorder = remember { AudioRecorderHelper(context) }

    // Permission
    val micPermissionState = rememberPermissionState(Manifest.permission.RECORD_AUDIO)

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        // Header
        Text(
            text = "Voice Control",
            style = MaterialTheme.typography.headlineMedium,
            fontWeight = FontWeight.Bold,
            modifier = Modifier.padding(bottom = 48.dp)
        )

        // Mic Button
        if (micPermissionState.status.isGranted) {
            Box(
                contentAlignment = Alignment.Center,
                modifier = Modifier
                    .size(160.dp)
                    .clip(CircleShape)
                    .background(if (isRecording) MaterialTheme.colorScheme.error else MaterialTheme.colorScheme.primary)
                    .pointerInput(Unit) {
                        detectTapGestures(
                            onPress = {
                                try {
                                    // Start Recording
                                    isRecording = true
                                    connectionStatus = "Listening..."
                                    val startResult = audioRecorder.startRecording()
                                    if (startResult == null) {
                                        isRecording = false
                                        connectionStatus = "Failed to start microphone"
                                    } else {
                                        awaitRelease()
                                    }
                                } catch (e: Exception) {
                                    Log.e("VoiceControl", "Recording error: ${e.message}", e)
                                    connectionStatus = "Error: ${e.message}"
                                } finally {
                                    // Stop Recording
                                    if (isRecording) {
                                        isRecording = false
                                        connectionStatus = "Processing..."
                                        val file = audioRecorder.stopRecording()
                                        
                                        if (file != null && file.exists() && file.length() > 0) {
                                            // Send to Backend
                                            Log.d("VoiceControl", "Sending file: ${file.length()} bytes")
                                            val bytes = file.readBytes()
                                            mqttHelper.publishAudio(bytes)
                                            connectionStatus = "Sent"
                                            file.delete()
                                        } else {
                                            connectionStatus = "Recording failed"
                                        }
                                    }
                                }
                            }
                        )
                    }
            ) {
                Icon(
                    imageVector = if (isRecording) Icons.Default.Mic else Icons.Default.MicOff,
                    contentDescription = "Mic",
                    tint = Color.White,
                    modifier = Modifier.size(80.dp)
                )
            }
            
            Spacer(modifier = Modifier.height(32.dp))

            Text(
                text = if (isRecording) "Release to Send" else "Hold to Speak",
                style = MaterialTheme.typography.bodyLarge
            )

            Spacer(modifier = Modifier.height(16.dp))

            // Status
            Text(
                text = "Status: $connectionStatus",
                style = MaterialTheme.typography.bodySmall,
                color = Color.Gray
            )
        } else {
            Button(onClick = { micPermissionState.launchPermissionRequest() }) {
                Text("Grant Mic Permission")
            }
        }
    }
}
