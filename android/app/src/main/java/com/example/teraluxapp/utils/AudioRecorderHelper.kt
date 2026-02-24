package com.example.teraluxapp.utils

import android.content.Context
import android.media.MediaRecorder
import android.os.Build
import android.util.Log
import java.io.File
import java.io.IOException

class AudioRecorderHelper(private val context: Context) {
    private var mediaRecorder: MediaRecorder? = null
    private var outputFile: File? = null
    private val TAG = "AudioRecorderHelper"

    fun startRecording(): File? {
        return try {
            val cacheDir = context.cacheDir
            outputFile = File.createTempFile("voice_record_", ".m4a", cacheDir)
            
            mediaRecorder = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
                MediaRecorder(context)
            } else {
                @Suppress("DEPRECATION")
                MediaRecorder()
            }

            mediaRecorder?.apply {
                setAudioSource(MediaRecorder.AudioSource.MIC)
                setOutputFormat(MediaRecorder.OutputFormat.MPEG_4)
                setAudioEncoder(MediaRecorder.AudioEncoder.AAC)
                setAudioSamplingRate(44100)
                setAudioEncodingBitRate(128000)
                setOutputFile(outputFile?.absolutePath)
                prepare()
                start()
            }
            Log.d(TAG, "Recording started: ${outputFile?.absolutePath}")
            outputFile
        } catch (e: Exception) {
            Log.e(TAG, "startRecording failed: ${e.message}", e)
            mediaRecorder?.release()
            mediaRecorder = null
            null
        }
    }

    fun stopRecording(): File? {
        val currentRecorder = mediaRecorder ?: return null
        return try {
            currentRecorder.stop()
            currentRecorder.release()
            mediaRecorder = null
            Log.d(TAG, "Recording stopped")
            outputFile
        } catch (e: Exception) {
            Log.e(TAG, "stopRecording failed: ${e.message}")
            try { currentRecorder.release() } catch (ignored: Exception) {}
            mediaRecorder = null
            null
        }
    }
}
