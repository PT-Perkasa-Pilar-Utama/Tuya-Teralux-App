package com.example.whisperandroid.data.local

import android.content.Context
import java.io.File
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale
import java.util.UUID

/**
 * Utility to manage app-private storage for meeting audio files.
 * Ensures consistent file naming and stable storage for True Resume functionality.
 */
object MeetingAudioFileStore {
    private const val DIRECTORY_NAME = "meetings"

    /**
     * Creates a new, uniquely named WAV file for microphone recordings.
     */
    fun createMicAudioFile(context: Context): File {
        val dir = getMeetingsDirectory(context)
        val timestamp = SimpleDateFormat("yyyyMMdd_HHmmss", Locale.US).format(Date())
        val uuid = UUID.randomUUID().toString()
        return File(dir, "meeting_mic_${timestamp}_$uuid.wav")
    }

    /**
     * Creates a new, uniquely named file for imported audio.
     */
    fun createImportedAudioFile(context: Context, extension: String): File {
        val dir = getMeetingsDirectory(context)
        val timestamp = SimpleDateFormat("yyyyMMdd_HHmmss", Locale.US).format(Date())
        val uuid = UUID.randomUUID().toString()
        val cleanExt = extension.removePrefix(".")
        return File(dir, "meeting_file_${timestamp}_$uuid.$cleanExt")
    }

    private fun getMeetingsDirectory(context: Context): File {
        val dir = File(context.filesDir, DIRECTORY_NAME)
        if (!dir.exists()) {
            dir.mkdirs()
        }
        return dir
    }

    /**
     * Retrieves a list of all meeting audio files sorted by newest first.
     */
    fun listMeetingAudioFiles(context: Context): List<File> {
        val dir = getMeetingsDirectory(context)
        return dir.listFiles()
            ?.filter { it.isFile && isSupportedAudio(it) }
            ?.sortedByDescending { it.lastModified() }
            ?: emptyList()
    }

    /**
     * Checks if the given file has a supported audio extension.
     */
    fun isSupportedAudio(file: File): Boolean {
        val name = file.name.lowercase()
        return name.endsWith(".wav") ||
            name.endsWith(".m4a") ||
            name.endsWith(".mp3") ||
            name.endsWith(".ogg") ||
            name.endsWith(".flac") ||
            name.endsWith(".mp4")
    }

    /**
     * Formats file size in MB.
     */
    fun formatSize(sizeInBytes: Long): String {
        val sizeInMb = sizeInBytes.toDouble() / (1024 * 1024)
        return String.format(Locale.getDefault(), "%.1f MB", sizeInMb)
    }

    /**
     * Formats last modified time.
     */
    fun formatTime(lastModified: Long): String {
        val sdf = SimpleDateFormat("MMM dd, yyyy HH:mm", Locale.getDefault())
        return sdf.format(Date(lastModified))
    }
}
