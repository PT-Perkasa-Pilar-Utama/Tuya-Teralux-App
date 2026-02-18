
package com.example.whisper_android.presentation.assistant

import android.content.Context
import android.content.Intent
import android.os.Bundle
import android.os.Handler
import android.os.Looper
import android.speech.RecognitionListener
import android.speech.RecognizerIntent
import android.speech.SpeechRecognizer
import android.util.Log

class GoogleSpeechWakeWordManager(
    private val context: Context,
    private val onWakeWordDetected: () -> Unit
) : WakeWordListener {

    private var speechRecognizer: SpeechRecognizer? = null
    private val handler = Handler(Looper.getMainLooper())
    private var isListening = false
    private var isDestroyed = false

    // Similar keywords to Vosk to maintain consistency
    private val wakeWords = listOf("hey sensio", "sensio", "sensyo", "sensus", "essence", "hi sensio")

    private val recognitionListener = object : RecognitionListener {
        override fun onReadyForSpeech(params: Bundle?) {
            Log.d("GoogleWakeWord", "Ready for speech")
        }

        override fun onBeginningOfSpeech() {
            Log.d("GoogleWakeWord", "Speech beginning")
        }

        override fun onRmsChanged(rmsdB: Float) {}

        override fun onBufferReceived(buffer: ByteArray?) {}

        override fun onEndOfSpeech() {
            Log.d("GoogleWakeWord", "End of speech")
        }

        override fun onError(error: Int) {
            val errorMessage = when (error) {
                SpeechRecognizer.ERROR_AUDIO -> "Audio recording error"
                SpeechRecognizer.ERROR_CLIENT -> "Client side error"
                SpeechRecognizer.ERROR_INSUFFICIENT_PERMISSIONS -> "Insufficient permissions"
                SpeechRecognizer.ERROR_NETWORK -> "Network error"
                SpeechRecognizer.ERROR_NETWORK_TIMEOUT -> "Network timeout"
                SpeechRecognizer.ERROR_NO_MATCH -> "No match"
                SpeechRecognizer.ERROR_RECOGNIZER_BUSY -> "Recognizer busy"
                SpeechRecognizer.ERROR_SERVER -> "Server error"
                SpeechRecognizer.ERROR_SPEECH_TIMEOUT -> "Speech timeout"
                else -> "Unknown error"
            }
            Log.d("GoogleWakeWord", "Error: $errorMessage ($error)")

            // Restart if not destroyed and was listening
            if (!isDestroyed && isListening) {
                restartListening()
            }
        }

        override fun onResults(results: Bundle?) {
            val matches = results?.getStringArrayList(SpeechRecognizer.RESULTS_RECOGNITION)
            processResults(matches)
            
            // Restart listening
            if (!isDestroyed && isListening) {
                restartListening()
            }
        }

        override fun onPartialResults(partialResults: Bundle?) {
            val matches = partialResults?.getStringArrayList(SpeechRecognizer.RESULTS_RECOGNITION)
            processResults(matches)
        }

        override fun onEvent(eventType: Int, params: Bundle?) {}
    }

    private fun processResults(matches: ArrayList<String>?) {
        matches?.forEach { phrase ->
            Log.d("GoogleWakeWord", "Heard: $phrase")
            if (containsWakeWord(phrase)) {
                Log.d("GoogleWakeWord", "Wake word detected: $phrase")
                triggerWakeWord()
                return // Stop processing current results
            }
        }
    }

    private fun containsWakeWord(text: String): Boolean {
        val lower = text.lowercase()
        return wakeWords.any { lower.contains(it) }
    }

    private fun triggerWakeWord() {
        stopListening() // Stop temporarily
        onWakeWordDetected()
    }

    private fun restartListening() {
        handler.postDelayed({
            if (!isDestroyed && isListening) {
                try {
                    startListeningInternal()
                } catch (e: Exception) {
                    Log.e("GoogleWakeWord", "Restart failed", e)
                }
            }
        }, 100) // Small delay to prevent tight loops
    }
    
    private fun startListeningInternal() {
         if (speechRecognizer == null) {
            if (SpeechRecognizer.isRecognitionAvailable(context)) {
                speechRecognizer = SpeechRecognizer.createSpeechRecognizer(context)
                speechRecognizer?.setRecognitionListener(recognitionListener)
            } else {
                Log.e("GoogleWakeWord", "SpeechRecognizer not available")
                return
            }
        }

        val intent = Intent(RecognizerIntent.ACTION_RECOGNIZE_SPEECH).apply {
            putExtra(RecognizerIntent.EXTRA_LANGUAGE_MODEL, RecognizerIntent.LANGUAGE_MODEL_FREE_FORM)
            putExtra(RecognizerIntent.EXTRA_PARTIAL_RESULTS, true)
            putExtra(RecognizerIntent.EXTRA_MAX_RESULTS, 3)
            // Optional: Request offline if possible to save battery/data, though accuracy might drop
            // putExtra(RecognizerIntent.EXTRA_PREFER_OFFLINE, true) 
        }

        try {
            speechRecognizer?.startListening(intent)
        } catch (e: Exception) {
            Log.e("GoogleWakeWord", "Start listening failed", e)
        }
    }

    override fun startListening() {
        isListening = true
        handler.post {
            startListeningInternal()
        }
    }

    override fun stopListening() {
        isListening = false
        handler.post {
            speechRecognizer?.stopListening()
        }
    }

    override fun destroy() {
        isDestroyed = true
        isListening = false
        handler.post {
            speechRecognizer?.destroy()
            speechRecognizer = null
        }
    }
}
