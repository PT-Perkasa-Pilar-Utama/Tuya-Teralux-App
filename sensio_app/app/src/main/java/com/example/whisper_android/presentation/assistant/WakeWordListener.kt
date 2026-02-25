
package com.example.whisper_android.presentation.assistant

interface WakeWordListener {
    fun startListening()

    fun stopListening()

    fun destroy()
}
