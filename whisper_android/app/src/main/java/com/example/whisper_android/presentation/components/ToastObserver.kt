package com.example.whisper_android.presentation.components

import android.content.Context
import android.widget.Toast
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.ui.platform.LocalContext

/**
 * A reusable component that observes a [message] and displays a [Toast].
 *
 * @param message The message to display. If null, nothing is shown.
 * @param onToastShown Callback to clear the message from the state after it's been displayed.
 */
@Composable
fun ToastObserver(
    message: String?,
    onToastShown: () -> Unit,
    context: Context = LocalContext.current
) {
    LaunchedEffect(message) {
        message?.let {
            Toast.makeText(context, it, Toast.LENGTH_SHORT).show()
            onToastShown()
        }
    }
}
