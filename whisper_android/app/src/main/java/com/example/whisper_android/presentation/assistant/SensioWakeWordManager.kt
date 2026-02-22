package com.example.whisper_android.presentation.assistant

import android.content.Context
import android.os.Handler
import android.os.Looper
import android.util.Log
import java.io.IOException
import org.vosk.Model
import org.vosk.Recognizer
import org.vosk.android.RecognitionListener
import org.vosk.android.SpeechService
import org.vosk.android.StorageService

class SensioWakeWordManager(
    private val context: Context,
    private val onWakeWordDetected: () -> Unit
) : WakeWordListener {
    private var model: Model? = null
    private var speechService: SpeechService? = null
    private val handler = Handler(Looper.getMainLooper())

    // Key phrases to listen for + noise absorbers
    private val keywords =
        "[" +
            "\"hey census\", \"hey essence\", \"hey sense\", \"hey essential\", " +
            "\"census\", \"essence\", \"hello\", \"hi\", \"yes\", \"no\", " +
            "\"computer\", \"assistant\", \"[unk]\"" +
            "]"

    init {
        initModel()
    }

    private fun initModel() {
        StorageService.unpack(
            context,
            "vosk-model",
            "model",
            { model: Model ->
                this.model = model
                Log.d("SensioWakeWord", "Vosk model unpacked and loaded successfully")
            },
            { exception: IOException ->
                Log.e("SensioWakeWord", "Failed to unpack Vosk model: ${exception.message}")
            }
        )
    }

    private val recognitionListener =
        object : RecognitionListener {
            override fun onPartialResult(hypothesis: String?) {
                hypothesis?.let {
                    if (containsWakeWord(it)) {
                        Log.d("SensioWakeWord", "Wake word detected (partial): $it")
                        triggerWakeWord()
                    }
                }
            }

            override fun onResult(hypothesis: String?) {
                hypothesis?.let {
                    if (containsWakeWord(it)) {
                        Log.d("SensioWakeWord", "Wake word detected: $it")
                        triggerWakeWord()
                    }
                }
            }

            override fun onFinalResult(hypothesis: String?) {
                hypothesis?.let {
                    if (containsWakeWord(it)) {
                        Log.d("SensioWakeWord", "Wake word detected (final): $it")
                        triggerWakeWord()
                    }
                }
            }

            override fun onError(exception: Exception?) {
                Log.e("SensioWakeWord", "Vosk Error: ${exception?.message}")
            }

            override fun onTimeout() {
                Log.d("SensioWakeWord", "Vosk Timeout")
            }
        }

    private fun containsWakeWord(hypothesis: String): Boolean {
        val lower = hypothesis.lowercase()
        // Since 'sensio' is missing from Vosk's dictionary, it often hears 'census' or 'essence'
        return lower.contains("hey census") ||
            lower.contains("hey essence") ||
            lower.contains("hey sense") ||
            lower.contains("hey essential") ||
            (lower.contains("census") && lower.contains("hey")) ||
            (lower.contains("essence") && lower.contains("hey"))
    }

    private fun triggerWakeWord() {
        // Stop listening temporarily to avoid double trigger
        stopListening()
        onWakeWordDetected()
    }

    override fun startListening() {
        handler.post {
            if (model == null) {
                Log.w("SensioWakeWord", "Model not ready yet, retrying in 1s")
                handler.postDelayed({ startListening() }, 1000)
                return@post
            }

            try {
                if (speechService == null) {
                    val rec = Recognizer(model, 16000.0f, keywords)
                    speechService = SpeechService(rec, 16000.0f)
                }
                speechService?.startListening(recognitionListener)
                Log.d("SensioWakeWord", "Started offline listening with Vosk")
            } catch (e: Exception) {
                Log.e("SensioWakeWord", "Failed to start Vosk: ${e.message}")
            }
        }
    }

    override fun stopListening() {
        handler.post {
            speechService?.stop()
            speechService?.setPause(true)
        }
    }

    override fun destroy() {
        speechService?.stop()
        speechService?.shutdown()
        speechService = null
    }
}
