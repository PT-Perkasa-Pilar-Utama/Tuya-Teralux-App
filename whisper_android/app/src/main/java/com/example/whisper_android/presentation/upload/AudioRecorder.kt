package com.example.whisper_android.presentation.upload

import android.content.Context
import android.media.MediaRecorder
import android.os.Build
import android.util.Log
import java.io.File
import java.io.FileOutputStream

/**
 * A simple manager for recording audio using MediaRecorder.
 * Records in MPEG_4 format with AAC encoding.
 */
class AudioRecorder(private val context: Context) {

    private var recorder: MediaRecorder? = null

    fun start(outputFile: File) {
        createRecorder().apply {
            setAudioSource(MediaRecorder.AudioSource.MIC)
            setOutputFormat(MediaRecorder.OutputFormat.MPEG_4)
            setAudioEncoder(MediaRecorder.AudioEncoder.AAC)
            setOutputFile(FileOutputStream(outputFile).fd)

            prepare()
            start()

            recorder = this
            Log.d("AudioRecorder", "Recording started: ${outputFile.absolutePath}")
        }
    }

    fun stop() {
        recorder?.stop()
        recorder?.reset()
        recorder?.release()
        recorder = null
        Log.d("AudioRecorder", "Recording stopped")
    }

    private fun createRecorder(): MediaRecorder {
        return if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
            MediaRecorder(context)
        } else {
            @Suppress("DEPRECATION")
            MediaRecorder()
        }
    }
}
