package com.example.whisper_android

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.tooling.preview.Preview
import com.example.whisper_android.presentation.dashboard.DashboardScreen
import com.example.whisper_android.presentation.register.RegisterScreen
import com.example.whisper_android.ui.theme.SmartMeetingRoomWhisperDemoTheme

class MainActivity : ComponentActivity() {
    companion object {
        lateinit var appContext: android.content.Context
            private set
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        appContext = applicationContext

        // Initialize dependency injection (Manual)
        com.example.whisper_android.data.di.NetworkModule
            .init(this)

        // Dynamic Orientation Locking
        if (com.example.whisper_android.util.DeviceUtils
                .isTeralux()
        ) {
            requestedOrientation = android.content.pm.ActivityInfo.SCREEN_ORIENTATION_LANDSCAPE
        } else if (com.example.whisper_android.util.DeviceUtils
                .isPhone(this)
        ) {
            requestedOrientation = android.content.pm.ActivityInfo.SCREEN_ORIENTATION_PORTRAIT
        } else {
            // Tablet: Allow rotation
            requestedOrientation = android.content.pm.ActivityInfo.SCREEN_ORIENTATION_UNSPECIFIED
        }

        androidx.core.view.WindowCompat
            .setDecorFitsSystemWindows(window, false)
        enableEdgeToEdge()
        setContent {
            SmartMeetingRoomWhisperDemoTheme {
                val context = androidx.compose.ui.platform.LocalContext.current
                var permissionsGranted by remember { androidx.compose.runtime.mutableStateOf(false) }

                val launcher =
                    androidx.activity.compose.rememberLauncherForActivityResult(
                        contract =
                            androidx.activity.result.contract.ActivityResultContracts
                                .RequestMultiplePermissions(),
                    ) { permissions ->
                        // Check if all permissions are granted
                        val allGranted = permissions.entries.all { it.value }
                        permissionsGranted = allGranted
                    }

                androidx.compose.runtime.LaunchedEffect(Unit) {
                    val permissionsToRequest =
                        mutableListOf(
                            android.Manifest.permission.RECORD_AUDIO,
                        )

                    if (android.os.Build.VERSION.SDK_INT >= android.os.Build.VERSION_CODES.TIRAMISU) {
                        permissionsToRequest.add(android.Manifest.permission.READ_MEDIA_AUDIO)
                    } else {
                        permissionsToRequest.add(android.Manifest.permission.READ_EXTERNAL_STORAGE)
                        permissionsToRequest.add(android.Manifest.permission.WRITE_EXTERNAL_STORAGE)
                    }

                    launcher.launch(permissionsToRequest.toTypedArray())
                }

                // Simple state-based navigation (replace with Navigation Component for complex apps)
                var currentScreen by remember { mutableStateOf("register") }

                when (currentScreen) {
                    "register" -> {
                        RegisterScreen(onNavigateToDashboard = { currentScreen = "dashboard" })
                    }

                    "dashboard" -> {
                        DashboardScreen(
                            onNavigateToRegister = { currentScreen = "register" },
                            onNavigateToUpload = { /* Deprecated */ },
                            onNavigateToStreaming = { currentScreen = "meeting" },
                            onNavigateToEdge = { currentScreen = "assistant" },
                        )
                    }

                    "meeting" -> {
                        com.example.whisper_android.presentation.meeting.MeetingTranscriberScreen(
                            onNavigateBack = { currentScreen = "dashboard" },
                        )
                    }

                    "assistant" -> {
                        com.example.whisper_android.presentation.assistant.AiAssistantScreen(
                            onNavigateBack = { currentScreen = "dashboard" },
                        )
                    }
                }
            }
        }
    }
}

@Composable
fun Greeting(
    name: String,
    modifier: Modifier = Modifier,
) {
    Text(
        text = "Hello $name!",
        style = MaterialTheme.typography.displayMedium,
        modifier = modifier,
    )
}

@Preview(showBackground = true)
@Composable
fun GreetingPreview() {
    SmartMeetingRoomWhisperDemoTheme {
        Greeting("Android")
    }
}
