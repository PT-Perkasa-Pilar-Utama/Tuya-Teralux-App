package com.example.whisper_android.presentation.components

import android.net.Uri
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.UploadFile
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color

@Composable
fun AudioFilePicker(
    onFileSelected: (Uri) -> Unit,
    modifier: Modifier = Modifier,
    enabled: Boolean = true,
    hasPermission: Boolean = true,
    onPermissionDenied: () -> Unit = {},
    onFallbackNeeded: () -> Unit = {},
    tint: Color = MaterialTheme.colorScheme.primary,
    disabledTint: Color = Color.Gray
) {
    val context = androidx.compose.ui.platform.LocalContext.current
    val launcher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.GetContent()
    ) { uri ->
        uri?.let { onFileSelected(it) }
    }

    IconButton(
        onClick = { 
            if (hasPermission) {
                try {
                    launcher.launch("audio/*")
                } catch (e: Exception) {
                    e.printStackTrace()
                    // Fallback to internal picker if system picker fails
                    onFallbackNeeded()
                }
            } else {
                onPermissionDenied()
            }
        },
        enabled = enabled,
        modifier = modifier
    ) {
        Icon(
            imageVector = Icons.Default.UploadFile,
            contentDescription = "Upload Audio File",
            tint = if (enabled) {
                if (hasPermission) tint else MaterialTheme.colorScheme.error // Red if permission missing
            } else disabledTint
        )
    }
}
