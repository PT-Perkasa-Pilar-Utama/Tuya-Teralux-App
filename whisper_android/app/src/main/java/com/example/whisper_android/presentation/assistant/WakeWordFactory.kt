
package com.example.whisper_android.presentation.assistant

import android.content.Context
import android.speech.SpeechRecognizer

object WakeWordFactory {
    fun getManager(
        context: Context,
        onDetected: () -> Unit
    ): WakeWordListener =
        if (SpeechRecognizer.isRecognitionAvailable(context)) {
            GoogleSpeechWakeWordManager(context, onDetected)
        } else {
            SensioWakeWordManager(context, onDetected)
        }
}
