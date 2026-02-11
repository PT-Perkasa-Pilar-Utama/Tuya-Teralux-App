package com.example.whisper_android.presentation.upload

import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.core.content.ContextCompat
import android.Manifest
import android.content.pm.PackageManager
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

import androidx.lifecycle.viewmodel.compose.viewModel
import com.example.whisper_android.data.di.NetworkModule
import com.example.whisper_android.presentation.components.FeatureScreenTemplate
import com.example.whisper_android.presentation.components.TranscriptionMessage
import com.example.whisper_android.presentation.components.ToastObserver
import com.example.whisper_android.MainActivity

@Composable
fun UploadScreen(
    onNavigateBack: () -> Unit,
    viewModel: UploadViewModel = viewModel {
        UploadViewModel(
            NetworkModule.speechRepository,
            AudioRecorder(MainActivity.appContext),
            NetworkModule.tokenManager,
            MainActivity.appContext.cacheDir
        )
    }
) {
    val uiState by viewModel.uiState.collectAsState()
    val context = LocalContext.current
    
    val hasPermission = ContextCompat.checkSelfPermission(
        context,
        Manifest.permission.RECORD_AUDIO
    ) == PackageManager.PERMISSION_GRANTED

    // Observe errors
    ToastObserver(
        message = uiState.error,
        onToastShown = { /* Error handled */ }
    )

    FeatureScreenTemplate(
        title = "Upload Audio",
        onNavigateBack = onNavigateBack,
        isRecording = uiState.isRecording,
        isProcessing = uiState.isProcessing,
        hasPermission = hasPermission,
        transcriptionResults = uiState.transcriptionResults,
        onMicClick = {
            if (hasPermission) {
                viewModel.toggleRecording()
            }
        },
        onClearChat = { viewModel.clearLog() },
        walkthroughContent = {
            Text(
                text = "🎯 Purpose",
                style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                color = MaterialTheme.colorScheme.primary
            )
            Text(
                text = "Long-Form Meeting & Audio Transcription (30 minutes to 4+ hours)",
                style = MaterialTheme.typography.bodyMedium
            )

            Text(
                text = "🔄 Flow Summary (Normal Upload)",
                style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                color = MaterialTheme.colorScheme.primary
            )
            Text(
                text = "1. The Android app (Kotlin) records audio using the microphone.\n" +
                        "2. The audio is saved as a complete audio file.\n" +
                        "3. The app sends the file to the backend (Golang) via REST API.\n" +
                        "4. The backend receives the audio file.\n" +
                        "5. The backend processes the file using whisper.cpp.\n" +
                        "6. The system returns the transcribed text.",
                style = MaterialTheme.typography.bodyMedium
            )

            Text(
                text = "✅ Advantages",
                style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                color = MaterialTheme.colorScheme.primary
            )
            Text(
                text = "• Simple architecture & easy to debug.\n" +
                        "• Stable for very long recordings.\n" +
                        "• No infrastructure message broker needed.\n" +
                        "• Better transcription consistency.",
                style = MaterialTheme.typography.bodyMedium
            )

            Text(
                text = "❌ Disadvantages",
                style = MaterialTheme.typography.titleMedium.copy(fontWeight = FontWeight.Bold),
                color = MaterialTheme.colorScheme.error
            )
            Text(
                text = "• Not real-time (must wait until finish).\n" +
                        "• Higher perceived latency for long sessions.\n" +
                        "• Large file uploads required.\n" +
                        "• Requires stable internet connection.",
                style = MaterialTheme.typography.bodyMedium
            )
        }
    )
}
