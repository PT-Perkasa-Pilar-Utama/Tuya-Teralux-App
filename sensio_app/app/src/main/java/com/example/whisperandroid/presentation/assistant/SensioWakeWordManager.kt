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

    private var isDestroyed = false
    private var isListeningRequested = false
    private var isWakeCycleActive = false
    private var lastWakeDetectedAtMs = 0L

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
                // Triggering from partial results can cause repeated false starts.
                // We only use stable onResult / onFinalResult for triggers in this flow.
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
        // Strict phrase matching using regex word boundaries (\b)
        val patterns = listOf(
            "sensio",
            "senso",
            "census",
            "sensei",
            "santio",
            "essence",
            "essential",
            "sencyo"
        ).joinToString("|")
        val regex = Regex("\\bhey\\s+($patterns)\\b")
        return regex.containsMatchIn(lower)
    }

    private fun triggerWakeWord() {
        val now = System.currentTimeMillis()
        if (now - lastWakeDetectedAtMs < 3000L) {
            Log.d("SensioWakeWord", "Wake word trigger ignored: Debounce active (3s)")
            return
        }

        if (isWakeCycleActive) {
            Log.d("SensioWakeWord", "Wake word trigger ignored: Cycle already active")
            return
        }

        isWakeCycleActive = true
        lastWakeDetectedAtMs = now

        // Stop listening immediately to avoid double trigger during pipeline start
        stopListening()
        onWakeWordDetected()
    }

    override fun startListening() {
        Log.d("SensioWakeWord", "Request to START listening (requested=$isListeningRequested, destroyed=$isDestroyed)")
        isListeningRequested = true
        isWakeCycleActive = false // Reset cycle to allow new trigger

        handler.post {
            if (isDestroyed) {
                Log.w("SensioWakeWord", "startListening ignored: already destroyed")
                return@post
            }

            if (!isListeningRequested) {
                Log.d("SensioWakeWord", "startListening aborted: no longer requested")
                return@post
            }

            if (model == null) {
                Log.w("SensioWakeWord", "Model not ready yet, retrying in 1s")
                handler.postDelayed({
                    if (isListeningRequested && !isDestroyed) {
                        startListening()
                    }
                }, 1000)
                return@post
            }

            try {
                if (speechService == null) {
                    Log.d("SensioWakeWord", "Creating new SpeechService")
                    val rec = Recognizer(model, 16000.0f, keywords)
                    speechService = SpeechService(rec, 16000.0f)
                }

                speechService?.setPause(false)
                var started = speechService?.startListening(recognitionListener) ?: false
                Log.d("SensioWakeWord", "Started offline listening (result=$started, serviceExists=${speechService != null})")

                if (!started && isListeningRequested && !isDestroyed) {
                    Log.w("SensioWakeWord", "Start failed, attempting self-healing (recreate service)")
                    speechService?.stop()
                    speechService?.shutdown()
                    speechService = null

                    // One-time synchronous retry with new service
                    try {
                        val rec = Recognizer(model, 16000.0f, keywords)
                        speechService = SpeechService(rec, 16000.0f)
                        speechService?.setPause(false)
                        started = speechService?.startListening(recognitionListener) ?: false
                        Log.d("SensioWakeWord", "Self-healing sync retry result: $started")
                    } catch (e: Exception) {
                        Log.e("SensioWakeWord", "Self-healing sync retry failed: ${e.message}")
                    }

                    // If still failed, schedule a delayed retry to avoid permanent death
                    if (!started && isListeningRequested && !isDestroyed) {
                        Log.w("SensioWakeWord", "Self-healing failed again, scheduling backoff retry in 1.5s")
                        handler.postDelayed({
                            if (isListeningRequested && !isDestroyed) {
                                startListening()
                            }
                        }, 1500)
                    }
                }
            } catch (e: Exception) {
                Log.e("SensioWakeWord", "Failed to start Vosk: ${e.message}")
            }
        }
    }

    override fun stopListening() {
        Log.d("SensioWakeWord", "Request to STOP listening")
        isListeningRequested = false
        handler.removeCallbacksAndMessages(null)

        handler.post {
            speechService?.setPause(true)
            speechService?.stop()
            Log.d("SensioWakeWord", "Stopped and paused SpeechService")
        }
    }

    override fun destroy() {
        Log.d("SensioWakeWord", "Destroying WakeWordManager")
        isDestroyed = true
        isListeningRequested = false
        handler.removeCallbacksAndMessages(null)

        speechService?.stop()
        speechService?.shutdown()
        speechService = null

        model?.close()
        model = null
    }
}
