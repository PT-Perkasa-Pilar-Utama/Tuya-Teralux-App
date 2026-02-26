
package com.example.whisper_android.presentation.assistant

import android.content.Context

object WakeWordFactory {
    fun getManager(
        context: Context,
        onDetected: () -> Unit
    ): WakeWordListener = SensioWakeWordManager(context, onDetected)
}
