package com.example.whisperandroid.presentation.assistant

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
            "\"hey sensio\", \"hey census\", \"hey essence\", \"hey sense\", \"hey essential\", " +
            "\"hey sensei\", \"hey santio\", \"hey senso\", " +
            "\"[unk]\"" +
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
        // Must contain 'hey' and one of the sensio-like variants (Vosk often misreads 'sensio')
        if (!lower.contains("hey")) return false

        return lower.contains("sensio") ||
            lower.contains("senso") ||
            lower.contains("census") ||
            lower.contains("essence") ||
            lower.contains("sensei") ||
            lower.contains("santio") ||
            lower.contains("essential")
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
