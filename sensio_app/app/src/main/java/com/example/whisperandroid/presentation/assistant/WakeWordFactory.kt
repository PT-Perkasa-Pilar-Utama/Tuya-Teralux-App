
package com.example.whisperandroid.presentation.assistant

import android.content.Context

object WakeWordFactory {
    fun getManager(
        context: Context,
        onDetected: () -> Unit
    ): WakeWordListener = SensioWakeWordManager(context, onDetected)
}
