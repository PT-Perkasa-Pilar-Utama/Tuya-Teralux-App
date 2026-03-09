
package com.example.whisperandroid.presentation.assistant

interface WakeWordListener {
    fun startListening()

    fun stopListening()

    fun isListeningRequested(): Boolean

    fun destroy()
}
