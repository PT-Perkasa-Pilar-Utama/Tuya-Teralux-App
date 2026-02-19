package com.example.whisper_android.presentation.meeting

import android.content.Context
import android.media.AudioFormat
import android.media.AudioRecord
import android.media.MediaRecorder
import android.util.Log
import java.io.DataOutputStream
import java.io.File
import java.io.FileOutputStream
import java.io.RandomAccessFile
import java.util.concurrent.atomic.AtomicBoolean

class AudioRecorder(
    private val context: Context,
) {
    private var audioRecord: AudioRecord? = null
    private var recordingThread: Thread? = null
    private val isRecording = AtomicBoolean(false)
    private var bytesWritten: Long = 0

    private val sampleRate = 16000
    private val channelConfig = AudioFormat.CHANNEL_IN_MONO
    private val audioFormat = AudioFormat.ENCODING_PCM_16BIT

    @android.annotation.SuppressLint("MissingPermission")
    fun start(outputFile: File) {
        if (isRecording.get()) return

        val minBufferSize = AudioRecord.getMinBufferSize(sampleRate, channelConfig, audioFormat)
        val bufferSize = (minBufferSize * 2).coerceAtLeast(sampleRate)

        val recorder =
            AudioRecord(
                MediaRecorder.AudioSource.MIC,
                sampleRate,
                channelConfig,
                audioFormat,
                bufferSize,
            )

        if (recorder.state != AudioRecord.STATE_INITIALIZED) {
            Log.e("AudioRecorder", "AudioRecord init failed")
            recorder.release()
            return
        }

        audioRecord = recorder
        bytesWritten = 0

        val outputStream = DataOutputStream(FileOutputStream(outputFile))
        writeWavHeader(outputStream, 1, sampleRate, 16, 0)

        isRecording.set(true)
        recorder.startRecording()

        recordingThread =
            Thread {
                val buffer = ByteArray(bufferSize)
                try {
                    while (isRecording.get()) {
                        val read = recorder.read(buffer, 0, buffer.size)
                        if (read > 0) {
                            outputStream.write(buffer, 0, read)
                            bytesWritten += read
                        }
                    }
                } catch (e: Exception) {
                    Log.e("AudioRecorder", "recording loop failed", e)
                } finally {
                    try {
                        outputStream.flush()
                        outputStream.close()
                    } catch (e: Exception) {
                        Log.e("AudioRecorder", "output close failed", e)
                    }
                }
            }.also { it.start() }
    }

    fun stop() {
        if (!isRecording.get()) return

        isRecording.set(false)
        try {
            audioRecord?.stop()
        } catch (e: Exception) {
            Log.e("AudioRecorder", "stop() failed", e)
        }
        audioRecord?.release()
        audioRecord = null

        try {
            recordingThread?.join()
        } catch (e: InterruptedException) {
            Log.e("AudioRecorder", "join() interrupted", e)
        }
        recordingThread = null
    }

    fun finalizeWav(outputFile: File) {
        try {
            RandomAccessFile(outputFile, "rw").use { raf ->
                writeWavHeader(raf, 1, sampleRate, 16, bytesWritten)
            }
        } catch (e: Exception) {
            Log.e("AudioRecorder", "finalizeWav failed", e)
        }
    }

    fun release() {
        if (isRecording.get()) {
            stop()
        }
    }

    private fun writeWavHeader(
        output: java.io.DataOutput,
        channels: Int,
        sampleRate: Int,
        bitsPerSample: Int,
        dataSize: Long,
    ) {
        val byteRate = sampleRate * channels * bitsPerSample / 8
        val blockAlign = (channels * bitsPerSample / 8).toShort()
        val chunkSize = 36 + dataSize

        output.writeBytes("RIFF")
        writeIntLE(output, chunkSize.toInt())
        output.writeBytes("WAVE")
        output.writeBytes("fmt ")
        writeIntLE(output, 16)
        writeShortLE(output, 1.toShort())
        writeShortLE(output, channels.toShort())
        writeIntLE(output, sampleRate)
        writeIntLE(output, byteRate)
        writeShortLE(output, blockAlign)
        writeShortLE(output, bitsPerSample.toShort())
        output.writeBytes("data")
        writeIntLE(output, dataSize.toInt())
    }

    private fun writeIntLE(
        output: java.io.DataOutput,
        value: Int,
    ) {
        output.writeByte(value and 0xFF)
        output.writeByte((value shr 8) and 0xFF)
        output.writeByte((value shr 16) and 0xFF)
        output.writeByte((value shr 24) and 0xFF)
    }

    private fun writeShortLE(
        output: java.io.DataOutput,
        value: Short,
    ) {
        output.writeByte(value.toInt() and 0xFF)
        output.writeByte((value.toInt() shr 8) and 0xFF)
    }
}
