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
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.tooling.preview.Preview

import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import com.example.whisper_android.presentation.register.RegisterScreen
import com.example.whisper_android.presentation.dashboard.DashboardScreen
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
        com.example.whisper_android.data.di.NetworkModule.init(this)
        
        androidx.core.view.WindowCompat.setDecorFitsSystemWindows(window, false)
        enableEdgeToEdge()
        setContent {
            SmartMeetingRoomWhisperDemoTheme {
                // Simple state-based navigation (replace with Navigation Component for complex apps)
                var currentScreen by remember { mutableStateOf("register") }

                when (currentScreen) {
                    "register" -> RegisterScreen(onNavigateToDashboard = { currentScreen = "dashboard" })
                    "dashboard" -> DashboardScreen(
                        onNavigateToRegister = { currentScreen = "register" },
                        onNavigateToUpload = { /* Deprecated */ },
                        onNavigateToStreaming = { currentScreen = "meeting" },
                        onNavigateToEdge = { currentScreen = "assistant" }
                    )
                    "meeting" -> com.example.whisper_android.presentation.meeting.MeetingTranscriberScreen(
                        onNavigateBack = { currentScreen = "dashboard" }
                    )
                    "assistant" -> com.example.whisper_android.presentation.assistant.AiAssistantScreen(
                        onNavigateBack = { currentScreen = "dashboard" }
                    )
                }
            }
        }
    }
}

@Composable
fun Greeting(name: String, modifier: Modifier = Modifier) {
    Text(
        text = "Hello $name!",
        style = MaterialTheme.typography.displayMedium,
        modifier = modifier
    )
}

@Preview(showBackground = true)
@Composable
fun GreetingPreview() {
    SmartMeetingRoomWhisperDemoTheme {
        Greeting("Android")
    }
}
