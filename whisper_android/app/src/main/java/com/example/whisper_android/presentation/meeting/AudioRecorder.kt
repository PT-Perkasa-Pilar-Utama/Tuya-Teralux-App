package com.example.whisper_android.presentation.meeting

import android.content.Context
import android.media.MediaRecorder
import android.os.Build
import android.util.Log
import java.io.File
import java.io.IOException

class AudioRecorder(private val context: Context) {

    private var recorder: MediaRecorder? = null

    private fun createRecorder(): MediaRecorder {
        return if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
            MediaRecorder(context)
        } else {
            MediaRecorder()
        }
    }

    fun start(outputFile: File) {
        createRecorder().apply {
            setAudioSource(MediaRecorder.AudioSource.MIC)
            setOutputFormat(MediaRecorder.OutputFormat.MPEG_4)
            setAudioEncoder(MediaRecorder.AudioEncoder.AAC)
            setOutputFile(outputFile.absolutePath)

            try {
                prepare()
                start()
                recorder = this
            } catch (e: IOException) {
                Log.e("AudioRecorder", "prepare() failed", e)
                e.printStackTrace()
            }
        }
    }

    fun stop() {
        recorder?.apply {
            try {
                stop()
                reset()
                release()
            } catch (e: Exception) {
                Log.e("AudioRecorder", "stop() failed", e)
                e.printStackTrace()
            }
        }
        recorder = null
    }

    // Optional: Release resources if the recorder is destroyed without stopping
    fun release() {
        recorder?.release()
        recorder = null
    }
}
