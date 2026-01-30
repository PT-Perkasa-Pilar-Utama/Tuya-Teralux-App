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
        val cacheDir = context.cacheDir
        // Use m4a (AAC) for efficient compression supported by Android
        outputFile = File.createTempFile("voice_record_", ".m4a", cacheDir)
        
        mediaRecorder = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
            MediaRecorder(context)
        } else {
            MediaRecorder()
        }

        return try {
            mediaRecorder?.apply {
                setAudioSource(MediaRecorder.AudioSource.MIC)
                setOutputFormat(MediaRecorder.OutputFormat.MPEG_4)
                setAudioEncoder(MediaRecorder.AudioEncoder.AAC)
                setOutputFile(outputFile?.absolutePath)
                prepare()
                start()
            }
            Log.d(TAG, "Recording started: ${outputFile?.absolutePath}")
            outputFile
        } catch (e: IOException) {
            Log.e(TAG, "prepare() failed")
            null
        } catch (e: Exception) {
            Log.e(TAG, "startRecording failed: ${e.message}")
            null
        }
    }

    fun stopRecording(): File? {
        return try {
            mediaRecorder?.apply {
                stop()
                release()
            }
            mediaRecorder = null
            Log.d(TAG, "Recording stopped")
            outputFile
        } catch (e: Exception) {
            Log.e(TAG, "stopRecording failed: ${e.message}")
            null
        }
    }
}
