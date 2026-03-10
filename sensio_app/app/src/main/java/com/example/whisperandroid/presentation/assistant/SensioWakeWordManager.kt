package com.example.whisperandroid.presentation.assistant

import android.content.Context
import android.os.Handler
import android.os.Looper
import com.example.whisperandroid.util.AppLog
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
    private val TAG = "WakeWord"
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
                AppLog.i(TAG, "Vosk model unpacked and loaded")
            },
            { exception: IOException ->
                AppLog.e(TAG, "Failed to unpack Vosk model", exception)
            }
        )
    }

    private val recognitionListener =
        object : RecognitionListener {
            override fun onPartialResult(hypothesis: String?) {}

            override fun onResult(hypothesis: String?) {
                hypothesis?.let {
                    if (containsWakeWord(it)) {
                        AppLog.d(TAG, "Wake word detected: $it")
                        triggerWakeWord()
                    }
                }
            }

            override fun onFinalResult(hypothesis: String?) {
                hypothesis?.let {
                    if (containsWakeWord(it)) {
                        AppLog.d(TAG, "Wake word detected (final): $it")
                        triggerWakeWord()
                    }
                }
            }

            override fun onError(exception: Exception?) {
                AppLog.e(TAG, "Vosk Error: ${exception?.message}")
            }

            override fun onTimeout() {}
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
        if (now - lastWakeDetectedAtMs < 3000L) return

        if (isWakeCycleActive) return

        isWakeCycleActive = true
        lastWakeDetectedAtMs = now

        AppLog.i(TAG, "Wake word triggered")
        // Stop listening immediately to avoid double trigger during pipeline start
        stopListening()
        onWakeWordDetected()
    }

    override fun startListening() {
        isListeningRequested = true
        isWakeCycleActive = false // Reset cycle to allow new trigger

        handler.post {
            if (isDestroyed || !isListeningRequested) return@post

            if (model == null) {
                handler.postDelayed({
                    if (isListeningRequested && !isDestroyed) {
                        startListening()
                    }
                }, 1000)
                return@post
            }

            try {
                if (speechService == null) {
                    val rec = Recognizer(model, 16000.0f, keywords)
                    speechService = SpeechService(rec, 16000.0f)
                }

                speechService?.setPause(false)
                val started = speechService?.startListening(recognitionListener) ?: false
                if (started) {
                    AppLog.d(TAG, "Started offline listening")
                } else if (isListeningRequested && !isDestroyed) {
                    AppLog.w(TAG, "Start failed, attempting self-healing")
                    speechService?.stop()
                    speechService?.shutdown()
                    speechService = null

                    // One-time synchronous retry with new service
                    try {
                        val rec = Recognizer(model, 16000.0f, keywords)
                        speechService = SpeechService(rec, 16000.0f)
                        speechService?.setPause(false)
                        val retryStarted = speechService?.startListening(recognitionListener) ?: false
                        AppLog.d(TAG, "Self-healing retry result: $retryStarted")
                    } catch (e: Exception) {
                        AppLog.e(TAG, "Self-healing retry failed", e)
                    }
                }
            } catch (e: Exception) {
                AppLog.e(TAG, "Failed to start Vosk", e)
            }
        }
    }

    override fun stopListening() {
        isListeningRequested = false
        handler.removeCallbacksAndMessages(null)

        handler.post {
            speechService?.setPause(true)
            speechService?.stop()
            AppLog.d(TAG, "Stopped SpeechService")
        }
    }

    override fun isListeningRequested(): Boolean = isListeningRequested

    override fun destroy() {
        AppLog.i(TAG, "Destroying WakeWordManager")
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
